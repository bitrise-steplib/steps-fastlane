package cache

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/cache"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

// AndroidGradleCacheItemCollector ...
type AndroidGradleCacheItemCollector struct {
	cmdFactory command.Factory
}

// NewAndroidGradleCacheItemCollector ...
func NewAndroidGradleCacheItemCollector(cmdFactory command.Factory) cache.ItemCollector {
	return AndroidGradleCacheItemCollector{cmdFactory: cmdFactory}
}

// Collect ...
func (c AndroidGradleCacheItemCollector) Collect(dir string, cacheLevel cache.Level) ([]string, []string, error) {
	if cacheLevel == cache.LevelNone {
		return nil, nil, nil
	}

	homeDir := pathutil.UserHomeDir()

	projectRoot, err := filepath.Abs(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("cache collection skipped: failed to determine project root path")
	}

	includePths, err := c.collectIncludePaths(homeDir, projectRoot, cacheLevel)
	if err != nil {
		return nil, nil, err
	}

	excludePths := c.collectExcludePaths(homeDir, projectRoot)

	return includePths, excludePths, nil
}

// Collect walks the directory tree underneath projectRoot and registers matching
// paths for caching based on the value of cacheLevel. Returns an error if there
// was an underlying error that would lead to a corrupted cache file, otherwise
// the given path is skipped.
func Collect(projectRoot string, cacheLevel cache.Level, cmdFactory command.Factory) error {
	cacheItemCollector := NewAndroidGradleCacheItemCollector(cmdFactory)
	includes, excludes, err := cacheItemCollector.Collect(projectRoot, cacheLevel)
	if err != nil {
		return err
	}

	if len(includes) == 0 && len(excludes) == 0 {
		return nil
	}

	gradleCache := cache.New()
	gradleCache.IncludePath(includes...)
	gradleCache.ExcludePath(excludes...)
	if err := gradleCache.Commit(); err != nil {
		return fmt.Errorf("failed to commit cache paths: %s", err)
	}

	return nil
}

func (c AndroidGradleCacheItemCollector) collectIncludePaths(homeDir, projectDir string, cacheLevel cache.Level) ([]string, error) {
	var includePths []string

	lockFilePath := filepath.Join(projectDir, "gradle.deps")

	lockfileContent := ""
	if err := filepath.Walk(projectDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk %s: %s", path, err)
		}

		if f.IsDir() || strings.Contains(path, "node_modules") {
			return nil
		}

		if !strings.HasSuffix(f.Name(), ".gradle") && !strings.HasSuffix(f.Name(), ".gradle.kts") && f.Name() != "gradlew-wrapper.properties" {
			return nil
		}

		unmodified, err := c.prepareUnmodifiedIndicator(path)
		if err != nil {
			log.Debugf(err.Error())
			unmodified = path
		}

		md5Hash, err := computeMD5String(unmodified)
		if err != nil {
			log.Warnf("Failed to compute MD5 hash of %s -> %s: %s", path, unmodified, err)
			return nil
		}

		lockfileContent += md5Hash

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to create cache indicator file: %s", err)
	}
	if err := fileutil.WriteStringToFile(lockFilePath, lockfileContent); err != nil {
		return nil, fmt.Errorf("failed to write indicator file: %s", err)
	}

	includePths = append(includePths, fmt.Sprintf("%s -> %s", filepath.Join(homeDir, ".gradle"), lockFilePath))
	includePths = append(includePths, fmt.Sprintf("%s -> %s", filepath.Join(homeDir, ".kotlin"), lockFilePath))
	includePths = append(includePths, fmt.Sprintf("%s -> %s", filepath.Join(homeDir, ".m2"), lockFilePath))

	if cacheLevel == cache.LevelAll {
		includePths = append(includePths, fmt.Sprintf("%s -> %s", filepath.Join(homeDir, ".android", "build-cache"), lockFilePath))

		if err := filepath.Walk(projectDir, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("failed to walk %s: %s", path, err)
			}

			if f.IsDir() {
				if f.Name() == "build" {
					includePths = append(includePths, path)
				}

				if f.Name() == ".gradle" {
					includePths = append(includePths, path)
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("failed to collect build cache: %s", err)
		}
	}

	return includePths, nil
}

/*
If the indicator is version controlled in git and has changes, we create a copy of it with its original content.
*/
func (c AndroidGradleCacheItemCollector) prepareUnmodifiedIndicator(indicator string) (unmodified string, err error) {
	indicatorDir := filepath.Dir(indicator)
	indicatorFile := filepath.Base(indicator)
	opts := command.Opts{Dir: indicatorDir}

	lsArgs := []string{"ls-files", "--error-unmatch", indicatorFile}
	lsCmd := c.cmdFactory.Create("git", lsArgs, &opts)
	code, err := lsCmd.RunAndReturnExitCode()
	if code != 0 {
		return "", fmt.Errorf("%s is not under git version control", indicator)
	}

	diffArgs := []string{"diff", "-s", "--exit-code", indicatorFile}
	diffCmd := c.cmdFactory.Create("git", diffArgs, &opts)
	code, err = diffCmd.RunAndReturnExitCode()
	if code == 0 {
		return "", fmt.Errorf("%s has not modification compared to HEAD", indicator)
	}

	file, err := ioutil.TempFile(os.TempDir(), "indicator")
	if err != nil {
		return "", err
	}
	defer func() {
		e := file.Close()
		if err == nil {
			err = e
		}
	}()

	showArgs := []string{"show", "HEAD:" + indicatorFile}
	showOpts := command.Opts{
		Stdout: file,
		Stderr: os.Stderr,
		Dir:    indicatorDir,
	}
	showCmd := c.cmdFactory.Create("git", showArgs, &showOpts)
	code, err = showCmd.RunAndReturnExitCode()
	if err != nil || code != 0 {
		return "", err
	}
	return file.Name(), nil
}

func computeMD5String(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("Failed to close %s: %s", filePath, err)
		}
	}()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (c AndroidGradleCacheItemCollector) collectExcludePaths(homeDir, projectDir string) []string {
	excludePths := []string{
		"!~/.gradle/daemon/*/daemon-*.out.log", // excludes Gradle daemon logs, like: ~/.gradle/daemon/6.1.1/daemon-3122.out.log
		"~/.android/build-cache/**",
		"*.lock",
		"*.bin",
		"*/build/*.json",
		"*/build/*.html",
		"*/build/*.xml",
		"*/build/*.properties",
		"*/build/*/zip-cache/*",
		"*.log",
		"*.txt",
		"*.rawproto",
		"!*.ap_",
		"!*.apk",
	}

	ver, err := projectGradleVersion(projectDir, c.cmdFactory)
	if err != nil {
		log.Warnf("Failed to get project gradle version: %s", err)
		return nil
	}

	{
		gradleUserHome := filepath.Join(homeDir, ".gradle")
		exist, err := pathutil.IsPathExists(gradleUserHome)
		if err != nil {
			log.Warnf("Failed to check if gradle user home dir (%s) exists: %s", gradleUserHome, err)
			return nil
		}
		if !exist {
			log.Warnf("Gradle user home dir (%s) does not exist", gradleUserHome)
			return nil
		}

		excludes, err := gradleUserHomeExcludePaths(gradleUserHome, ver)
		if err != nil {
			log.Warnf("Failed to collect gradle user home exclude paths: %s", err)
			return nil
		}

		excludePths = append(excludePths, excludes...)
	}

	{
		excludes, err := projectGradleExcludePaths(projectDir, ver)
		if err != nil {
			log.Warnf("Failed to collect project gradle exclude paths: %s", err)
			return nil
		}

		excludePths = append(excludePths, excludes...)

	}

	return excludePths
}
