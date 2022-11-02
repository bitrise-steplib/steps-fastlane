package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitrise-io/go-steputils/v2/cache/compression"
	"github.com/bitrise-io/go-steputils/v2/cache/keytemplate"
	"github.com/bitrise-io/go-steputils/v2/cache/network"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/docker/go-units"
)

// SaveCacheInput is the information that comes from the cache steps that call this shared implementation
type SaveCacheInput struct {
	// StepId identifies the exact cache step. Used for logging events.
	StepId  string
	Verbose bool
	Key     string
	Paths   []string
	// IsKeyUnique indicates that the cache key is enough for knowing the cache archive is different from
	// another cache archive.
	// This can be set to true if the cache key contains a checksum that changes when any of the cached files change.
	// Example of such key: my-cache-key-{{ checksum "package-lock.json" }}
	// Example where this is not true: my-cache-key-{{ .OS }}-{{ .Arch }}
	IsKeyUnique bool
}

// Saver ...
type Saver interface {
	Save(input SaveCacheInput) error
}

type saveCacheConfig struct {
	Verbose        bool
	Key            string
	Paths          []string
	APIBaseURL     stepconf.Secret
	APIAccessToken stepconf.Secret
}

type saver struct {
	envRepo      env.Repository
	logger       log.Logger
	pathProvider pathutil.PathProvider
	pathModifier pathutil.PathModifier
	pathChecker  pathutil.PathChecker
}

// NewSaver ...
func NewSaver(
	envRepo env.Repository,
	logger log.Logger,
	pathProvider pathutil.PathProvider,
	pathModifier pathutil.PathModifier,
	pathChecker pathutil.PathChecker,
) *saver {
	return &saver{
		envRepo:      envRepo,
		logger:       logger,
		pathProvider: pathProvider,
		pathModifier: pathModifier,
		pathChecker:  pathChecker,
	}
}

// Save ...
func (s *saver) Save(input SaveCacheInput) error {
	config, err := s.createConfig(input)
	if err != nil {
		return fmt.Errorf("failed to parse inputs: %w", err)
	}

	tracker := newStepTracker(input.StepId, s.envRepo, s.logger)
	defer tracker.wait()

	canSkipSave, reason := s.canSkipSave(input.Key, config.Key, input.IsKeyUnique)
	tracker.logSkipSaveResult(canSkipSave, reason)
	s.logger.Println()
	if canSkipSave {
		s.logger.Donef("Cache save can be skipped, reason: %s", reason.description())
		return nil
	} else {
		s.logger.Infof("Can't skip saving the cache, reason: %s", reason.description())
		if reason == reasonNoRestoreThisKey {
			s.logOtherHits()
		}
	}

	s.logger.Println()
	s.logger.Infof("Creating archive...")
	compressionStartTime := time.Now()
	archivePath, err := s.compress(config.Paths)
	if err != nil {
		return fmt.Errorf("compression failed: %s", err)
	}
	compressionTime := time.Since(compressionStartTime).Round(time.Second)
	tracker.logArchiveCompressed(compressionTime, len(config.Paths))
	s.logger.Donef("Archive created in %s", compressionTime)

	fileInfo, err := os.Stat(archivePath)
	if err != nil {
		return err
	}
	s.logger.Printf("Archive size: %s", units.HumanSizeWithPrecision(float64(fileInfo.Size()), 3))
	s.logger.Debugf("Archive path: %s", archivePath)

	archiveChecksum, err := checksumOfFile(archivePath)
	if err != nil {
		s.logger.Warnf(err.Error())
		// fail silently and continue
	}
	canSkipUpload, reason := s.canSkipUpload(config.Key, archiveChecksum)
	tracker.logSkipUploadResult(canSkipUpload, reason)
	s.logger.Println()
	if canSkipUpload {
		s.logger.Donef("Cache upload can be skipped, reason: %s", reason.description())
		return nil
	}
	s.logger.Infof("Can't skip uploading the cache, reason: %s", reason.description())

	s.logger.Println()
	s.logger.Infof("Uploading archive...")
	uploadStartTime := time.Now()
	err = s.upload(archivePath, fileInfo.Size(), config)
	if err != nil {
		return fmt.Errorf("cache upload failed: %w", err)
	}
	uploadTime := time.Since(uploadStartTime).Round(time.Second)
	s.logger.Donef("Archive uploaded in %s", uploadTime)
	tracker.logArchiveUploaded(uploadTime, fileInfo, len(config.Paths))

	return nil
}

func (s *saver) createConfig(input SaveCacheInput) (saveCacheConfig, error) {
	if strings.TrimSpace(input.Key) == "" {
		return saveCacheConfig{}, fmt.Errorf("cache key should not be empty")
	}

	s.logger.Println()
	s.logger.Printf("Evaluating key template: %s", input.Key)
	evaluatedKey, err := s.evaluateKey(input.Key)
	if err != nil {
		return saveCacheConfig{}, fmt.Errorf("failed to evaluate key template: %s", err)
	}
	s.logger.Donef("Cache key: %s", evaluatedKey)

	finalPaths, err := s.evaluatePaths(input.Paths)
	if err != nil {
		return saveCacheConfig{}, fmt.Errorf("failed to parse paths: %w", err)
	}

	apiBaseURL := s.envRepo.Get("BITRISEIO_ABCS_API_URL")
	if apiBaseURL == "" {
		return saveCacheConfig{}, fmt.Errorf("the secret 'BITRISEIO_ABCS_API_URL' is not defined")
	}
	apiAccessToken := s.envRepo.Get("BITRISEIO_ABCS_ACCESS_TOKEN")
	if apiAccessToken == "" {
		return saveCacheConfig{}, fmt.Errorf("the secret 'BITRISEIO_ABCS_ACCESS_TOKEN' is not defined")
	}

	return saveCacheConfig{
		Verbose:        input.Verbose,
		Key:            evaluatedKey,
		Paths:          finalPaths,
		APIBaseURL:     stepconf.Secret(apiBaseURL),
		APIAccessToken: stepconf.Secret(apiAccessToken),
	}, nil
}

func (s *saver) evaluatePaths(paths []string) ([]string, error) {
	// Expand wildcard paths
	var expandedPaths []string
	for _, path := range paths {
		if !strings.Contains(path, "*") {
			expandedPaths = append(expandedPaths, path)
			continue
		}

		base, pattern := doublestar.SplitPattern(path)
		absBase, err := s.pathModifier.AbsPath(base) // resolves ~/ and expands any envs
		if err != nil {
			return nil, err
		}
		matches, err := doublestar.Glob(os.DirFS(absBase), pattern)
		if matches == nil {
			s.logger.Warnf("No match for path pattern: %s", path)
			continue
		}
		if err != nil {
			s.logger.Warnf("Error in path pattern '%s': %w", path, err)
			continue
		}

		for _, match := range matches {
			expandedPaths = append(expandedPaths, filepath.Join(base, match))
		}
	}

	// Validate and sanitize paths
	var finalPaths []string
	for _, path := range expandedPaths {
		absPath, err := s.pathModifier.AbsPath(path)
		if err != nil {
			s.logger.Warnf("Failed to parse path %s, error: %s", path, err)
			continue
		}

		exists, err := s.pathChecker.IsPathExists(absPath)
		if err != nil {
			s.logger.Warnf("Failed to check path %s, error: %s", absPath, err)
		}
		if !exists {
			s.logger.Warnf("Cache path doesn't exist: %s", path)
			continue
		}

		finalPaths = append(finalPaths, absPath)
	}

	return finalPaths, nil
}

func (s *saver) evaluateKey(keyTemplate string) (string, error) {
	model := keytemplate.NewModel(s.envRepo, s.logger)
	return model.Evaluate(keyTemplate)
}

func (s *saver) compress(paths []string) (string, error) {
	if compression.AreAllPathsEmpty(paths) {
		s.logger.Warnf("The provided paths are all empty, skipping compression and upload.")
		os.Exit(0)
	}

	fileName := fmt.Sprintf("cache-%s.tzst", time.Now().UTC().Format("20060102-150405"))
	tempDir, err := s.pathProvider.CreateTempDir("save-cache")
	if err != nil {
		return "", err
	}
	archivePath := filepath.Join(tempDir, fileName)

	err = compression.Compress(archivePath, paths, s.logger, s.envRepo)
	if err != nil {
		return "", err
	}

	return archivePath, nil
}

func (s *saver) upload(archivePath string, archiveSize int64, config saveCacheConfig) error {
	params := network.UploadParams{
		APIBaseURL:  string(config.APIBaseURL),
		Token:       string(config.APIAccessToken),
		ArchivePath: archivePath,
		ArchiveSize: archiveSize,
		CacheKey:    config.Key,
	}
	return network.Upload(params, s.logger)
}
