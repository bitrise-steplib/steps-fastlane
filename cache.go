package main

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/bitrise-io/bitrise-init/scanners/android"
	androidCache "github.com/bitrise-io/go-android/v2/cache"
	"github.com/bitrise-io/go-steputils/cache"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
)

type depsFunc func(dir string) ([]string, []string, error)

func (f FastlaneRunner) cacheDeps(opts RunOpts) {
	if opts.EnableCache {
		f.logger.Println()
		f.logger.Infof("Collecting cache")

		var depsFuncs = []depsFunc{
			f.cocoapodsDeps,
			f.carthageDeps,
			f.androidDeps,
		}

		c := cache.New()
		for _, depFunc := range depsFuncs {
			includes, excludes, err := depFunc(opts.WorkDir)
			f.logger.Debugf("%s found include path:\n%s\nexclude paths:\n%s", f.functionName(depFunc), strings.Join(includes, "\n"), strings.Join(excludes, "\n"))
			if err != nil {
				f.logger.Warnf("failed to collect dependencies: %s", err.Error())
				continue
			}

			for _, item := range includes {
				c.IncludePath(item)
			}

			for _, item := range excludes {
				c.ExcludePath(item)
			}
		}
		if err := c.Commit(); err != nil {
			f.logger.Warnf("failed to commit paths to cache: %s", err)
		}
	}
}

func (f FastlaneRunner) functionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func (f FastlaneRunner) iosDeps(dir string, buildDirName, lockFileName string) ([]string, []string, error) {
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

func (f FastlaneRunner) cocoapodsDeps(dir string) ([]string, []string, error) {
	return f.iosDeps(dir, "Pods", "Podfile.lock")
}

func (f FastlaneRunner) carthageDeps(dir string) ([]string, []string, error) {
	return f.iosDeps(dir, "Carthage", "Cartfile.resolved")
}

func (f FastlaneRunner) androidDeps(dir string) ([]string, []string, error) {
	scanner := android.NewScanner()
	detected, err := scanner.DetectPlatform(dir)
	if err != nil {
		return nil, nil, err
	}
	log.Debugf("android platform detected: %v", detected)

	var include []string
	var exclude []string
	for _, project := range scanner.Projects {
		i, e, err := androidCache.NewAndroidGradleCacheItemCollector(command.NewFactory(env.NewRepository())).Collect(project.RelPath, cache.LevelDeps)
		if err != nil {
			return nil, nil, err
		}

		include = append(include, i...)
		exclude = append(exclude, e...)
	}

	return include, exclude, err
}
