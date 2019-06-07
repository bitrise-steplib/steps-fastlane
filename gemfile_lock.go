package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

type gemVersion struct {
	version string
	found   bool
}

type gemVersions struct {
	fastlane, bundler gemVersion
}

func gemVersionFromGemfileLockContent(gemName string, content string) gemVersion {
	relevantLines := []string{}
	lines := strings.Split(content, "\n")

	specsStart := false
	for _, line := range lines {
		if strings.Contains(line, "specs:") {
			specsStart = true
		}

		trimmed := strings.Trim(line, " ")
		if trimmed == "" {
			specsStart = false
		}

		if specsStart {
			relevantLines = append(relevantLines, line)
		}
	}

	//     fastlane (1.109.0)
	exp := regexp.MustCompile(fmt.Sprintf(`^%s \((.+)\)`, regexp.QuoteMeta(gemName)))
	for _, line := range relevantLines {
		match := exp.FindStringSubmatch(strings.TrimSpace(line))
		if len(match) == 2 {
			return gemVersion{
				version: match[1],
				found:   true,
			}
		}
	}

	return gemVersion{}
}

func fastlaneVersionFromGemfileLock(content string) gemVersion {
	return gemVersionFromGemfileLockContent("fastlane", content)
}

func bundlerVersionFromGemfileLock(content string) gemVersion {
	return gemVersionFromGemfileLockContent("bundler", content)
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

	gemVersions.fastlane = fastlaneVersionFromGemfileLock(content)
	if gemVersions.fastlane.found {
		log.Printf("Gemfile.lock defined fastlane version: %s", gemVersions.fastlane.version)
	} else {
		log.Printf("No fastlane version defined in Gemfile.lock")
	}

	gemVersions.bundler = bundlerVersionFromGemfileLock(content)
	if gemVersions.bundler.found {
		log.Printf("Gemfile.lock defined bundler version: %s", gemVersions.bundler.version)
	} else {
		log.Printf("No bundler version defined in Gemfile.lock")
	}

	return gemVersions, nil
}
