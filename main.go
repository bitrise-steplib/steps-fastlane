package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/rubycommand"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	shellquote "github.com/kballard/go-shellquote"
)

// ConfigsModel ...
type ConfigsModel struct {
	WorkDir        string
	Lane           string
	UpdateFastlane string
}

func createConfigsModelFromEnvs() ConfigsModel {
	return ConfigsModel{
		WorkDir:        os.Getenv("work_dir"),
		Lane:           os.Getenv("lane"),
		UpdateFastlane: os.Getenv("update_fastlane"),
	}
}

func (configs ConfigsModel) print() {
	log.Infof("Configs:")
	log.Printf("- WorkDir: %s", configs.WorkDir)
	log.Printf("- Lane: %s", configs.Lane)
	log.Printf("- UpdateFastlane: %s", configs.UpdateFastlane)
}

func (configs ConfigsModel) validate() error {
	if configs.Lane == "" {
		return errors.New("no Lane parameter specified")
	}

	if configs.WorkDir != "" {
		if exist, err := pathutil.IsDirExists(configs.WorkDir); err != nil {
			return fmt.Errorf("failed to check if WorkDir exist at: %s, error: %s", configs.WorkDir, err)
		} else if !exist {
			return fmt.Errorf("WorkDir not exist at: %s", configs.WorkDir)
		}
	}

	if configs.UpdateFastlane == "" {
		return errors.New("no UpdateFastlane parameter specified")
	} else if configs.UpdateFastlane != "true" && configs.UpdateFastlane != "false" {
		return fmt.Errorf(`invalid UpdateFastlane parameter specified: %s, available: ["true", "false"]`, configs.UpdateFastlane)
	}

	return nil
}

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func fastlaneVersionFromGemfileLockContent(content string) string {
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
	exp := regexp.MustCompile(`fastlane \((.+)\)`)
	for _, line := range relevantLines {
		match := exp.FindStringSubmatch(line)
		if match != nil && len(match) == 2 {
			return match[1]
		}
	}

	return ""
}

func fastlaneVersionFromGemfileLock(gemfileLockPth string) (string, error) {
	content, err := fileutil.ReadStringFromFile(gemfileLockPth)
	if err != nil {
		return "", err
	}
	return fastlaneVersionFromGemfileLockContent(content), nil
}

func main() {
	configs := createConfigsModelFromEnvs()

	fmt.Println()
	configs.print()

	if err := configs.validate(); err != nil {
		failf("Issue with input: %s", err)
	}

	// Expand WorkDir
	fmt.Println()
	log.Infof("Expand WorkDir")

	workDir := configs.WorkDir
	if workDir == "" {
		log.Printf("WorkDir not set, using CurrentWorkingDirectory...")
		currentDir, err := pathutil.CurrentWorkingDirectoryAbsolutePath()
		if err != nil {
			failf("Failed to get current dir, error: %s", err)
		}
		workDir = currentDir
	} else {
		absWorkDir, err := pathutil.AbsPath(workDir)
		if err != nil {
			failf("Failed to expand path (%s), error: %s", workDir, err)
		}
		workDir = absWorkDir
	}

	log.Donef("Expanded WorkDir: %s", workDir)

	// Split lane option
	laneOptions, err := shellquote.Split(configs.Lane)
	if err != nil {
		failf("Failed to parse lane (%s), error: %s", configs.Lane, err)
	}

	// Determine desired Fastlane version
	fmt.Println()
	log.Infof("Determine desired Fastlane version")

	useBundler := false

	gemfileLockPth := filepath.Join(workDir, "Gemfile.lock")
	log.Printf("Checking Gemfile.lock (%s) for fastlane gem", gemfileLockPth)

	if exist, err := pathutil.IsPathExists(gemfileLockPth); err != nil {
		failf("Failed to check if Gemfile.lock exist at (%s), error: %s", gemfileLockPth, err)
	} else if exist {
		version, err := fastlaneVersionFromGemfileLock(gemfileLockPth)
		if err != nil {
			failf("Failed to read Fastlane versiom from Gemfile.lock (%s), error: %s", gemfileLockPth, err)
		}

		if version != "" {
			log.Printf("Gemfile.lock defined fastlane version: %s", version)

			useBundler = true
		} else {
			log.Printf("No fastlane version defined in Gemfile.lock")
		}
	} else {
		log.Printf("Gemfile.lock does not exist")
	}

	fmt.Println()

	// Install desired Fastlane version
	if useBundler {
		log.Infof("Install Fastlane with bundler")

		bundleInstallCmd := []string{"bundle", "install", "--jobs", "20", "--retry", "5"}

		log.Donef("$ %s", command.PrintableCommandArgs(false, bundleInstallCmd))

		cmd, err := rubycommand.NewFromSlice(bundleInstallCmd...)
		if err != nil {
			failf("Failed to create command model, error: %s", err)
		}

		cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
		cmd.SetDir(workDir)

		if err := cmd.Run(); err != nil {
			failf("Command failed, error: %s", err)
		}
	} else if configs.UpdateFastlane == "true" {
		log.Infof("Update system installed Fastlane")

		cmds, err := rubycommand.GemInstall("fastlane", "")
		if err != nil {
			failf("Failed to create command model, error: %s", err)
		}

		for _, cmd := range cmds {
			log.Donef("$ %s", cmd.PrintableCommandArgs())

			cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
			cmd.SetDir(workDir)

			if err := cmd.Run(); err != nil {
				failf("Command failed, error: %s", err)
			}
		}
	} else {
		log.Infof("Using system installed Fastlane")
	}

	fmt.Println()
	log.Infof("Fastlane version:")

	versionCmd := []string{"fastlane", "--version"}
	if useBundler {
		versionCmd = append([]string{"bundle", "exec"}, versionCmd...)
	}

	log.Donef("$ %s", command.PrintableCommandArgs(false, versionCmd))

	cmd, err := rubycommand.NewFromSlice(versionCmd...)
	if err != nil {
		failf("Command failed, error: %s", err)
	}

	cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
	cmd.SetDir(workDir)

	if err := cmd.Run(); err != nil {
		failf("Command failed, error: %s", err)
	}

	// Run fastlane
	fmt.Println()
	log.Infof("Run Fastlane")

	fastlaneCmd := []string{"fastlane"}
	fastlaneCmd = append(fastlaneCmd, laneOptions...)
	if useBundler {
		fastlaneCmd = append([]string{"bundle", "exec"}, fastlaneCmd...)
	}

	log.Donef("$ %s", command.PrintableCommandArgs(false, fastlaneCmd))

	cmd, err = rubycommand.NewFromSlice(fastlaneCmd...)
	if err != nil {
		failf("Failed to create command model, error: %s", err)
	}

	cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
	cmd.SetDir(workDir)

	if err := cmd.Run(); err != nil {
		failf("Command failed, error: %s", err)
	}
}
