package main

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command/gems"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

type gemVersions struct {
	fastlane, bundler gems.Version
}

func parseGemfileLock(searchDir string) (gemVersions, error) {
	gemfileLockPth := filepath.Join(searchDir, "Gemfile.lock")
	log.Printf("Checking Gemfile.lock (%s) for fastlane and bundler gem", gemfileLockPth)

	if exist, err := pathutil.IsPathExists(gemfileLockPth); err != nil {
		return gemVersions{}, fmt.Errorf("failed to check if Gemfile.lock exist at (%s), error: %s", gemfileLockPth, err)
	} else if !exist {
		log.Printf("Gemfile.lock does not exist")
		return gemVersions{}, nil
	}

	content, err := fileutil.ReadStringFromFile(gemfileLockPth)
	if err != nil {
		return gemVersions{}, err
	}

	var gemVersions gemVersions

	gemVersions.fastlane, err = gems.ParseVersionFromBundle("fastlane", content)
	if err != nil {
		return gemVersions, err
	}
	if gemVersions.fastlane.Found {
		log.Infof("Gemfile.lock defined fastlane version: %s", gemVersions.fastlane.Version)
	} else {
		log.Infof("No fastlane version defined in Gemfile.lock")
	}

	gemVersions.bundler, err = gems.ParseBundlerVersion(content)
	if err != nil {
		return gemVersions, err
	}
	if gemVersions.bundler.Found {
		log.Infof("Gemfile.lock defined bundler version: %s", gemVersions.bundler.Version)
	} else {
		log.Infof("No bundler version defined in Gemfile.lock")
	}

	return gemVersions, nil
}
