package main

import (
	"fmt"
	"path/filepath"

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
	iosDir, err := pathutil.AbsPath(filepath.Join(dir, "ios"))
	if err != nil {
		return nil, nil, err
	}

	podfileLockPth := filepath.Join(iosDir, "Podfile.lock")
	if exist, err := pathutil.IsPathExists(podfileLockPth); err != nil {
		return nil, nil, err
	} else if !exist {
		return nil, nil, nil
	}

	return []string{fmt.Sprintf("%s -> %s", filepath.Join(iosDir, "Pods"), podfileLockPth)}, nil, nil
}

func cacheCarthageDeps(dir string) ([]string, []string, error) {
	iosDir, err := pathutil.AbsPath(filepath.Join(dir, "ios"))
	if err != nil {
		return nil, nil, err
	}

	cartfileResolvedPth := filepath.Join(iosDir, "Cartfile.resolved")
	if exist, err := pathutil.IsPathExists(cartfileResolvedPth); err != nil {
		return nil, nil, err
	} else if !exist {
		return nil, nil, nil
	}

	return []string{fmt.Sprintf("%s -> %s", filepath.Join(iosDir, "Carthage"), cartfileResolvedPth)}, nil, nil
}

func cacheAndroidDeps(dir string) ([]string, []string, error) {
	androidDir := filepath.Join(dir, "android")

	exist, err := pathutil.IsDirExists(androidDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check if directory (%s) exists, error: %s", androidDir, err)
	}
	if !exist {
		return nil, nil, nil
	}
	return androidCache.Deps(androidDir, androidCache.LevelDeps)
}
