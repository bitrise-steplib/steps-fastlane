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
			log.Printf("Gem lockfile does not exist")
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
		log.Infof("Gem lockfile defined Fastlane version: %s", gemVersions.fastlane.Version)
	} else {
		log.Infof("No Fastlane version defined in gem lockfile")
	}

	gemVersions.bundler, err = gems.ParseBundlerVersion(content)
	if err != nil {
		return gemVersions, err
	}
	if gemVersions.bundler.Found {
		log.Infof("Gem lockfile defined bundler version: %s", gemVersions.bundler.Version)
	} else {
		log.Infof("No bundler version defined in gem lockfile")
	}

	return gemVersions, nil
}
