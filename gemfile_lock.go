package main

import (
	"errors"

	"github.com/bitrise-io/go-steputils/v2/ruby"
)

type gemVersions struct {
	fastlane, bundler ruby.Version
}

func (f FastlaneRunner) parseGemfileLock(searchDir string) (gemVersions, error) {
	content, err := ruby.GemFileLockContent(searchDir)
	if err != nil {
		if errors.Is(err, ruby.ErrGemLockNotFound) {
			f.logger.Printf("Gem lockfile does not exist")
			return gemVersions{}, nil
		}
		return gemVersions{}, err
	}

	var gemVersions gemVersions

	gemVersions.fastlane, err = ruby.ParseVersionFromBundle("fastlane", content)
	if err != nil {
		return gemVersions, err
	}
	if gemVersions.fastlane.Found {
		f.logger.Printf("Gem lockfile defined Fastlane version: %s", gemVersions.fastlane.Version)
	} else {
		f.logger.Printf("No Fastlane version defined in gem lockfile")
	}

	gemVersions.bundler, err = ruby.ParseBundlerVersion(content)
	if err != nil {
		return gemVersions, err
	}
	if gemVersions.bundler.Found {
		f.logger.Printf("Gem lockfile defined bundler version: %s", gemVersions.bundler.Version)
	} else {
		f.logger.Printf("No bundler version defined in gem lockfile")
	}

	return gemVersions, nil
}
