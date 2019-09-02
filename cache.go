package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-init/scanners/android"
	"github.com/bitrise-io/bitrise-init/utility"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	androidCache "github.com/bitrise-steplib/bitrise-step-android-unit-test/cache"
)

type depsFunc func(dir string) ([]string, []string, error)

var depsFuncs = []depsFunc{
	cocoapodsDeps,
	cacheCarthageDeps,
	cacheAndroidDeps,
}

func cocoapodsDeps(dir string) ([]string, []string, error) {
	files, err := utility.ListPathInDirSortedByComponents(dir, true)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to search for files in (%s), error: %s", dir, err)
	}

	locks, err := utility.FilterPaths(files, utility.BaseFilter("Podfile.lock", true))
	if err != nil {
		return nil, nil, err
	}

	var relevant []string
	for _, lock := range locks {
		podsPth := filepath.Join(filepath.Dir(lock), "Pods")
		exist, err := pathutil.IsPathExists(podsPth)
		if err != nil {
			return nil, nil, err
		}

		if exist {
			relevant = append(relevant, lock)
		}
	}

	if len(relevant) > 0 {
		log.Debugf("Multiple Podfile.lock found: %s", strings.Join(relevant, ", "))
	}

	var include []string
	for _, lock := range locks {
		podsPth := filepath.Join(filepath.Dir(lock), "Pods")
		include = append(include, fmt.Sprintf("%s -> %s", podsPth, lock))
	}

	return include, nil, nil
}

func cacheCarthageDeps(dir string) ([]string, []string, error) {
	files, err := utility.ListPathInDirSortedByComponents(dir, true)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to search for files in (%s), error: %s", dir, err)
	}

	locks, err := utility.FilterPaths(files, utility.BaseFilter("Cartfile.resolved", true))
	if err != nil {
		return nil, nil, err
	}

	var relevant []string
	for _, lock := range locks {
		carthagePth := filepath.Join(filepath.Dir(lock), "Carthage")
		exist, err := pathutil.IsPathExists(carthagePth)
		if err != nil {
			return nil, nil, err
		}

		if exist {
			relevant = append(relevant, lock)
		}
	}

	if len(relevant) > 0 {
		log.Debugf("Multiple Cartfile.resolved found: %s", strings.Join(relevant, ", "))
	}

	var include []string
	for _, lock := range locks {
		carthagePth := filepath.Join(filepath.Dir(lock), "Carthage")
		include = append(include, fmt.Sprintf("%s -> %s", carthagePth, lock))
	}

	return include, nil, nil
}

func cacheAndroidDeps(dir string) ([]string, []string, error) {
	scanner := android.NewScanner()
	_, err := scanner.DetectPlatform(dir)
	if err != nil {
		return nil, nil, err
	}

	var include []string
	var exclude []string
	for _, dir := range scanner.ProjectRoots {
		i, e, err := androidCache.Deps(dir, androidCache.LevelDeps)
		if err != nil {
			return nil, nil, err
		}

		include = append(include, i...)
		exclude = append(exclude, e...)
	}

	return include, exclude, err
}
