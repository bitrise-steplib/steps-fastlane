package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-init/scanners/android"
	androidCache "github.com/bitrise-io/go-android/cache"
	"github.com/bitrise-io/go-steputils/cache"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

type depsFunc func(dir string) ([]string, []string, error)

var depsFuncs = []depsFunc{
	cocoapodsDeps,
	carthageDeps,
	androidDeps,
}

func iosDeps(dir string, buildDirName, lockFileName string) ([]string, []string, error) {
	files, err := pathutil.ListPathInDirSortedByComponents(dir, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to search for files in (%s), error: %s", dir, err)
	}

	locks, err := pathutil.FilterPaths(files, pathutil.BaseFilter(lockFileName, true))
	if err != nil {
		return nil, nil, err
	}

	buildDirToLockFile := map[string]string{}
	for _, lock := range locks {
		buildDir := filepath.Join(filepath.Dir(lock), buildDirName)
		exist, err := pathutil.IsPathExists(buildDir)
		if err != nil {
			return nil, nil, err
		}

		if exist {
			buildDirToLockFile[buildDir] = lock
		}
	}

	if len(buildDirToLockFile) > 1 {
		var locks []string
		for _, lock := range buildDirToLockFile {
			locks = append(locks, lock)
		}
		log.Debugf("Multiple %s found: %s", lockFileName, strings.Join(locks, ", "))
	}

	var include []string
	for buildDir, lockFile := range buildDirToLockFile {
		include = append(include, fmt.Sprintf("%s -> %s", buildDir, lockFile))
	}

	return include, nil, nil
}

func cocoapodsDeps(dir string) ([]string, []string, error) {
	return iosDeps(dir, "Pods", "Podfile.lock")
}

func carthageDeps(dir string) ([]string, []string, error) {
	return iosDeps(dir, "Carthage", "Cartfile.resolved")
}

func androidDeps(dir string) ([]string, []string, error) {
	scanner := android.NewScanner()
	detected, err := scanner.DetectPlatform(dir)
	if err != nil {
		return nil, nil, err
	}
	log.Debugf("android platform detected: %v", detected)

	var include []string
	var exclude []string
	for _, dir := range scanner.ProjectRoots {
		i, e, err := androidCache.NewAndroidGradleCacheItemCollector(command.NewFactory(env.NewRepository())).Collect(dir, cache.LevelDeps)
		if err != nil {
			return nil, nil, err
		}

		include = append(include, i...)
		exclude = append(exclude, e...)
	}

	return include, exclude, err
}
