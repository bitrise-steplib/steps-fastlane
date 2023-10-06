package android

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise-init/analytics"
	"github.com/bitrise-io/bitrise-init/models"
	"github.com/bitrise-io/bitrise-init/utility"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

type fileGroups [][]string

var pathUtilIsPathExists = pathutil.IsPathExists
var filePathWalk = filepath.Walk

// Project is an Android project on the filesystem
type Project struct {
	RelPath               string
	UsesKotlinBuildScript bool
	Icons                 models.Icons
	Warnings              models.Warnings
}

func detect(searchDir string) ([]Project, error) {
	projectFiles := fileGroups{
		{"build.gradle", gradleKotlinBuildFile},
		{"settings.gradle", gradleKotlinSettingsFile},
	}
	skipDirs := []string{".git", "CordovaLib", "node_modules"}

	log.TInfof("Searching for android files")

	projectRoots, err := walkMultipleFileGroups(searchDir, projectFiles, skipDirs)
	if err != nil {
		return nil, fmt.Errorf("failed to search for build.gradle files, error: %s", err)
	}

	log.TPrintf("%d android files detected", len(projectRoots))
	for _, file := range projectRoots {
		log.TPrintf("- %s", file)
	}

	if len(projectRoots) == 0 {
		return nil, nil
	}
	log.TSuccessf("Platform detected")

	projects, err := parseProjects(searchDir, projectRoots)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func parseProjects(searchDir string, projectRoots []string) ([]Project, error) {
	var (
		lastErr  error
		projects []Project
	)

	for _, projectRoot := range projectRoots {
		var warnings models.Warnings

		log.TInfof("Investigating Android project: %s", projectRoot)

		exists, err := containsLocalProperties(projectRoot)
		if err != nil {
			lastErr = err
			log.TWarnf("%s", err)

			continue
		}
		if exists {
			containsLocalPropertiesWarning := fmt.Sprintf("the local.properties file should NOT be checked into Version Control Systems, as it contains information specific to your local configuration, the location of the file is: %s", filepath.Join(projectRoot, "local.properties"))
			warnings = []string{containsLocalPropertiesWarning}
		}

		if err := checkGradlew(projectRoot); err != nil {
			lastErr = err
			log.TWarnf("%s", err)

			continue
		}

		relProjectRoot, err := filepath.Rel(searchDir, projectRoot)
		if err != nil {
			lastErr = err
			log.TWarnf("%s", err)

			continue
		}

		icons, err := LookupIcons(projectRoot, searchDir)
		if err != nil {
			analytics.LogInfo("android-icon-lookup", analytics.DetectorErrorData("android", err), "Failed to lookup android icon")
		}

		kotlinBuildScriptBased := usesKotlinBuildScripts(projectRoot)
		projects = append(projects, Project{
			RelPath:               relProjectRoot,
			UsesKotlinBuildScript: kotlinBuildScriptBased,
			Icons:                 icons,
			Warnings:              warnings,
		})
	}

	if len(projects) == 0 {
		return []Project{}, lastErr
	}

	return projects, nil
}

func usesKotlinBuildScripts(projectRoot string) bool {
	return utility.FileExists(filepath.Join(projectRoot, gradleKotlinBuildFile)) && utility.FileExists(filepath.Join(projectRoot, gradleKotlinSettingsFile))
}

func walk(src string, fn func(path string, info os.FileInfo) error) error {
	return filePathWalk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == src {
			return nil
		}
		return fn(path, info)
	})
}

func checkFileGroups(path string, fileGroups fileGroups) (bool, error) {
	for _, fileGroup := range fileGroups {
		found := false
		for _, file := range fileGroup {
			exists, err := pathUtilIsPathExists(filepath.Join(path, file))
			if err != nil {
				return found, err
			}
			if exists {
				found = true
			}
		}
		if !found {
			return false, nil
		}
	}
	return true, nil
}

func walkMultipleFileGroups(searchDir string, fileGroups fileGroups, skipDirs []string) (matches []string, err error) {
	match, err := checkFileGroups(searchDir, fileGroups)
	if err != nil {
		return nil, err
	}
	if match {
		matches = append(matches, searchDir)
	}
	return matches, walk(searchDir, func(path string, info os.FileInfo) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if nameMatchSkipDirs(info.Name(), skipDirs) {
				return filepath.SkipDir
			}
			match, err := checkFileGroups(path, fileGroups)
			if err != nil {
				return err
			}
			if match {
				matches = append(matches, path)
			}
		}
		return nil
	})
}

func nameMatchSkipDirs(name string, skipDirs []string) bool {
	for _, skipDir := range skipDirs {
		if skipDir == "" {
			continue
		}
		if name == skipDir {
			return true
		}
	}
	return false
}

func containsLocalProperties(projectDir string) (bool, error) {
	return pathutil.IsPathExists(filepath.Join(projectDir, "local.properties"))
}

func checkGradlew(projectDir string) error {
	gradlewPth := filepath.Join(projectDir, "gradlew")
	exist, err := pathutil.IsPathExists(gradlewPth)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New(`<b>No Gradle Wrapper (gradlew) found.</b>
Using a Gradle Wrapper (gradlew) is required, as the wrapper is what makes sure
that the right Gradle version is installed and used for the build. More info/guide: <a>https://docs.gradle.org/current/userguide/gradle_wrapper.html</a>`)
	}
	return nil
}

type configBuildingParams struct {
	name            string
	useKotlinScript bool
}

func configBuildingParameters(projects []Project) []configBuildingParams {
	regularProjectCount := 0
	kotlinBuildScriptProjectCount := 0

	for _, project := range projects {
		if project.UsesKotlinBuildScript {
			kotlinBuildScriptProjectCount += 1
		} else {
			regularProjectCount += 1
		}
	}

	var params []configBuildingParams
	if 0 < regularProjectCount {
		params = append(params, configBuildingParams{
			name:            ConfigName,
			useKotlinScript: false,
		})
	}
	if 0 < kotlinBuildScriptProjectCount {
		params = append(params, configBuildingParams{
			name:            ConfigNameKotlinScript,
			useKotlinScript: true,
		})
	}
	return params
}
