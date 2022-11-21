package main

import (
	"github.com/bitrise-io/go-steputils/command/gems"
	"github.com/bitrise-io/go-utils/log"
)

type gemVersions struct {
	fastlane, bundler gems.Version
}

func (f FastlaneRunner) parseGemfileLock(searchDir string) (gemVersions, error) {
	content, err := gems.GemFileLockContent(searchDir)
	if err != nil {
		if err == gems.ErrGemLockNotFound {
			f.logger.Printf("Gem lockfile does not exist")
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
		log.Printf("Gem lockfile defined Fastlane version: %s", gemVersions.fastlane.Version)
	} else {
		log.Printf("No Fastlane version defined in gem lockfile")
	}

	gemVersions.bundler, err = gems.ParseBundlerVersion(content)
	if err != nil {
		return gemVersions, err
	}
	if gemVersions.bundler.Found {
		log.Printf("Gem lockfile defined bundler version: %s", gemVersions.bundler.Version)
	} else {
		log.Printf("No bundler version defined in gem lockfile")
	}

	return gemVersions, nil
}
