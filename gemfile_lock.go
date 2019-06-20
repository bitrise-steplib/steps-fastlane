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

	gemVersions.fastlane = parseFastlaneVersion(content)
	if gemVersions.fastlane.found {
		log.Infof("Gemfile.lock defined fastlane version: %s", gemVersions.fastlane.version)
	} else {
		log.Infof("No fastlane version defined in Gemfile.lock")
	}

	gemVersions.bundler = parseBundlerVersion(content)
	if gemVersions.bundler.found {
		log.Infof("Gemfile.lock defined bundler version: %s", gemVersions.bundler.version)
	} else {
		log.Infof("No bundler version defined in Gemfile.lock")
	}

	return gemVersions, nil
}

func parseFastlaneVersion(gemfileLockContent string) gemVersion {
	return parseGemVersion("fastlane", gemfileLockContent)
}

// parseGemVersion returns the gem version parsed from a Gemfile.lock on a best effort basis, for logging purposes only.
//
// parseGemVersion("fastlane", ...), for the following Gemfile.lock example, it returns: ">= 2.0)"
//   specs:
//     CFPropertyList (3.0.0)
//     addressable (2.6.0)
//       public_suffix (>= 2.0.2, < 4.0)
//     atomos (0.1.3)
//     babosa (1.0.2)
//     badge (0.8.5)
//       curb (~> 0.9)
//       fastimage (>= 1.6)
//       fastlane (>= 2.0)
//       mini_magick (>= 4.5)
//     claide (1.0.2)
func parseGemVersion(gemName string, content string) gemVersion {
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

func parseBundlerVersion(gemfileLockContent string) gemVersion {
	/*
		BUNDLED WITH
			1.17.1
	*/
	bundlerRegexp := regexp.MustCompile(`(?m)^BUNDLED WITH\n\s+(\S+)`)
	match := bundlerRegexp.FindStringSubmatch(gemfileLockContent)
	if match == nil {
		log.Warnf("failed to parse bundler version in Gemfile.lock: %s", gemfileLockContent)
		fmt.Println()
		return gemVersion{}
	}
	if len(match) != 2 {
		log.Warnf("unexpected regexp match: %v", match)
		return gemVersion{}
	}

	return gemVersion{
		version: match[1],
		found:   true,
	}
}
