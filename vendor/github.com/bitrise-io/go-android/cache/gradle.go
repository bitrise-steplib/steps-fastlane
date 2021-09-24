package cache

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/hashicorp/go-version"
)

func parseGradleVersion(out string) (string, error) {
	/*
	   ------------------------------------------------------------
	   Gradle 6.1.1
	   ------------------------------------------------------------

	   Build time:   2020-01-24 22:30:24 UTC
	   Revision:     a8c3750babb99d1894378073499d6716a1a1fa5d

	   Kotlin:       1.3.61
	   Groovy:       2.5.8
	   Ant:          Apache Ant(TM) version 1.10.7 compiled on September 1 2019
	   JVM:          1.8.0_241 (Oracle Corporation 25.241-b07)
	   OS:           Mac OS X 10.15.5 x86_64
	*/

	pattern := `-+\sGradle (.*)\s-+`
	exp := regexp.MustCompile(pattern)
	matches := exp.FindStringSubmatch(out)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to find Gradle version in output:\n%s\nusing a pattern:%s", out, pattern)
	}
	return matches[1], nil
}

func projectGradleVersion(projectPth string, cmdFactory command.Factory) (string, error) {
	gradlewPth := filepath.Join(projectPth, "gradlew")
	exist, err := pathutil.IsPathExists(gradlewPth)
	if err != nil {
		return "", fmt.Errorf("failed to check if %s exists: %s", gradlewPth, err)
	}
	if !exist {
		return "", fmt.Errorf("no gradlew found at: %s", gradlewPth)
	}

	versionCmdOpts := command.Opts{Dir: filepath.Dir(gradlewPth)}
	versionCmd := cmdFactory.Create("./gradlew", []string{"-version"}, &versionCmdOpts)
	out, err := versionCmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return "", fmt.Errorf("%s failed: %s", versionCmd.PrintableCommandArgs(), out)
		}
		return "", fmt.Errorf("%s failed: %s", versionCmd.PrintableCommandArgs(), err)
	}

	return parseGradleVersion(out)
}

func gradleUserHomeExcludePaths(gradleUserHome, currentGradleVersion string) ([]string, error) {
	var excludes []string

	{
		// exclude old wrappers, like ~/.gradle/wrapper/dists/gradle-5.1.1-all
		wrapperDistrDir := filepath.Join(gradleUserHome, "wrapper", "dists")
		entries, err := ioutil.ReadDir(wrapperDistrDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read entries of %s: %s", wrapperDistrDir, err)
		}
		for _, e := range entries {
			if !strings.HasPrefix(e.Name(), "gradle-"+currentGradleVersion) {
				excludes = append(excludes, "!"+filepath.Join(wrapperDistrDir, e.Name()))
			}
		}
	}

	{
		// exclude old caches, like ~/.gradle/caches/5.1.1
		cachesDir := filepath.Join(gradleUserHome, "caches")
		entries, err := ioutil.ReadDir(cachesDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read entries of %s: %s", cachesDir, err)
		}
		for _, e := range entries {
			v, err := version.NewVersion(e.Name())
			if err != nil || v == nil {
				continue
			}

			if e.Name() != currentGradleVersion {
				excludes = append(excludes, "!"+filepath.Join(cachesDir, e.Name()))
			}
		}
	}

	{
		// exclude old daemon, like ~/.gradle/daemon/5.1.1
		daemonDir := filepath.Join(gradleUserHome, "daemon")
		entries, err := ioutil.ReadDir(daemonDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read entries of %s: %s", daemonDir, err)
		}
		for _, e := range entries {
			v, err := version.NewVersion(e.Name())
			if err != nil || v == nil {
				continue
			}

			if e.Name() != currentGradleVersion {
				excludes = append(excludes, "!"+filepath.Join(daemonDir, e.Name()))
			}
		}
	}

	return excludes, nil
}

func projectGradleExcludePaths(projectDir, currentGradleVersion string) ([]string, error) {
	var excludes []string

	gradleDir := filepath.Join(projectDir, ".gradle")
	entries, err := ioutil.ReadDir(gradleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read entries of %s: %s", gradleDir, err)
	}
	for _, e := range entries {
		v, err := version.NewVersion(e.Name())
		if err != nil || v == nil {
			continue
		}

		if e.Name() != currentGradleVersion {
			excludes = append(excludes, "!"+filepath.Join(gradleDir, e.Name()))
		}
	}

	return excludes, nil
}
