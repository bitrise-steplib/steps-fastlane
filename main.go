package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/tools"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/rubycommand"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-steplib/steps-deploy-to-itunesconnect-deliver/devportalservice"
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

	//
	// Fastlane session
	fmt.Println()
	log.Infof("Ensure cookies for Apple Developer Portal")

	fs, errors := devportalservice.SessionData()
	if errors != nil {
		log.Warnf("Failed to activate the Bitrise Apple Developer Portal connection: %s\nRead more: https://devcenter.bitrise.io/getting-started/signing-up/connecting-apple-dev-account/\nerrors:")
		for _, err := range errors {
			log.Errorf("%s\n", err)
		}
	} else {
		if err := tools.ExportEnvironmentWithEnvman("FASTLANE_SESSION", fs); err != nil {
			failf("Failed to export FASTLANE_SESSION, error: %s", err)
		}

		if err := os.Setenv("FASTLANE_SESSION", fs); err != nil {
			failf("Failed to set FASTLANE_SESSION env, error: %s", err)
		}

		log.Donef("Session exported")
	}

	// Split lane option
	laneOptions, err := shellquote.Split(configs.Lane)
	if err != nil {
		failf("Failed to parse lane (%s), error: %s", configs.Lane, err)
	}

	// Determine desired Fastlane version
	fmt.Println()
	log.Infof("Determine desired Fastlane version")

	gemVersions, err := parseGemfileLock(workDir)
	if err != nil {
		failf("%s", err)
	}

	useBundler := false
	if gemVersions.fastlane.found {
		useBundler = true
	}

	fmt.Println()

	// Install desired Fastlane version
	if useBundler {
		log.Infof("Install bundler")

		// install bundler with `gem install bundler [-v version]`
		installBundlerCommand, err := getInstallBundlerCommand(gemVersions.bundler)
		if err != nil {
			failf("failed to create command, error: %s", err)
		}

		log.Debugf("$ %s", installBundlerCommand.PrintableCommandArgs())
		fmt.Println()

		out, err := installBundlerCommand.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			if errorutil.IsExitStatusError(err) {
				failf("failed to install bundler, command exited with error: %s, out: %s", err, out)
			}
			failf("failed to install bundler, failed to run command, error: %s", err)
		}

		// install Gemfile.lock gems with `bundle [_version_] install ...`
		log.Infof("Install Fastlane with bundler")

		cmd, err := getBundleInstallCommand(gemVersions.bundler)
		if err != nil {
			failf("failed to create bundle command, error: %s", err)
		}

		log.Donef("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()

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

	cmd, err := rubycommand.NewFromSlice(versionCmd)
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

	cmd, err = rubycommand.NewFromSlice(fastlaneCmd)
	if err != nil {
		failf("Failed to create command model, error: %s", err)
	}

	cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
	cmd.SetDir(workDir)

	buildlogPth := ""

	if tempDir, err := pathutil.NormalizedOSTempDirPath("fastlane_logs"); err != nil {
		log.Errorf("Failed to create temp dir for fastlane logs, error: %s", err)
	} else {
		buildlogPth = tempDir
		cmd.AppendEnvs("FL_BUILDLOG_PATH=" + buildlogPth)
	}

	deployDir := os.Getenv("BITRISE_DEPLOY_DIR")
	if deployDir == "" {
		log.Warnf("No BITRISE_DEPLOY_DIR found")
	}
	deployPth := filepath.Join(deployDir, "fastlane_env.log")

	if err := cmd.Run(); err != nil {
		fmt.Println()
		log.Errorf("Fastlane command: (%s) failed", cmd.PrintableCommandArgs())
		log.Errorf("If you want to send an issue report to fastlane (https://github.com/fastlane/fastlane/issues/new), you can find the output of fastlane env in the following log file:")
		fmt.Println()
		log.Infof(deployPth)
		fmt.Println()

		if cmd, err := rubycommand.New("fastlane", "env"); err != nil {
			log.Warnf("Failed to create command model, error: %s", err)
		} else {
			inputReader := strings.NewReader("n")
			var outBuffer bytes.Buffer
			outWriter := bufio.NewWriter(&outBuffer)

			cmd.SetStdin(inputReader)
			cmd.SetStdout(outWriter).SetStderr(os.Stderr)
			cmd.SetDir(workDir)

			if errEnv := cmd.Run(); errEnv != nil {
				log.Warnf("Fastlane command: (%s) failed", cmd.PrintableCommandArgs())
			} else if outBuffer.String() != "" {
				if err := fileutil.WriteStringToFile(deployPth, outBuffer.String()); err != nil {
					log.Warnf("Failed to write fastlane env log file, error: %s", err)
				}
			}
		}

		if err := filepath.Walk(buildlogPth, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if relLogPath, err := filepath.Rel(buildlogPth, path); err != nil {
					return err
				} else if err := os.Rename(path, filepath.Join(deployDir, strings.Replace(relLogPath, "/", "_", -1))); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			log.Errorf("Failed to walk directory, error: %s", err)
		}
		failf("Command failed, error: %s", err)
	}
}
