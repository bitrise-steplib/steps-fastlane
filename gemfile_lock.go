package main

import (
	"github.com/bitrise-io/go-utils/command/gems"
	"github.com/bitrise-io/go-utils/log"
)

type gemVersions struct {
	fastlane, bundler gems.Version
}

func parseGemfileLock(searchDir string) (gemVersions, error) {
	content, err := gems.GemFileLockContent(searchDir)
	if err != nil {
		if err == gems.ErrGemLockNotFound {
			log.Printf("Gemfile.lock does not exist")
			return gemVersions{}, nil
		}
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
