package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/command/gems"
	"github.com/bitrise-io/go-steputils/ruby"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

// Run ...
func (s FastlaneRunner) Run(config Config, opts EnsureDependenciesOpts) error {
	// Run fastlane
	fmt.Println()
	s.logger.Infof("Run Fastlane")

	var envs []string
	authEnvs, err := FastlaneAuthParams(config.AuthCredentials)
	if err != nil {
		return fmt.Errorf("Failed to set up Fastlane authentication parameters: %v", err)
	}
	var globallySetAuthEnvs []string
	for envKey, envValue := range authEnvs {
		if _, set := os.LookupEnv(envKey); set {
			globallySetAuthEnvs = append(globallySetAuthEnvs, envKey)
		}

		envs = append(envs, fmt.Sprintf("%s=%s", envKey, envValue))
	}
	if len(globallySetAuthEnvs) != 0 {
		s.logger.Warnf("Fastlane authentication-related environment varibale(s) (%s) are set, overriding.", globallySetAuthEnvs)
		s.logger.Infof("To stop overriding authentication-related environment variables, please set Bitrise Apple Developer Connection input to 'off' and leave authentication-related inputs empty.")
	}

	buildlogPth := ""
	if tempDir, err := pathutil.NormalizedOSTempDirPath("fastlane_logs"); err != nil {
		s.logger.Errorf("Failed to create temp dir for fastlane logs, error: %s", err)
	} else {
		buildlogPth = tempDir
		envs = append(envs, "FL_BUILDLOG_PATH="+buildlogPth)
	}

	name := "fastlane"
	args := config.LaneOptions
	options := &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Dir:    config.WorkDir,
		Env:    append(os.Environ(), envs...),
	}
	var cmd command.Command
	if opts.UseBundler {
		cmd = s.rbyFactory.CreateBundleExec(name, args, opts.GemVersions.bundler.Version, options)
	} else {
		cmd = s.rbyFactory.Create(name, args, options)
	}

	s.logger.Donef("$ %s", cmd.PrintableCommandArgs())

	deployDir := os.Getenv("BITRISE_DEPLOY_DIR")
	if deployDir == "" {
		s.logger.Warnf("No BITRISE_DEPLOY_DIR found")
	}
	deployPth := filepath.Join(deployDir, "fastlane_env.log")

	if err := cmd.Run(); err != nil {
		fmt.Println()
		s.logger.Errorf("Fastlane command: (%s) failed", cmd.PrintableCommandArgs())
		s.logger.Errorf("If you want to send an issue report to fastlane (https://github.com/fastlane/fastlane/issues/new), you can find the output of fastlane env in the following log file:")
		fmt.Println()
		s.logger.Infof(deployPth)
		fmt.Println()

		if fastlaneDebugInfo, err := s.fastlaneDebugInfo(config.WorkDir, opts.UseBundler, opts.GemVersions.bundler); err != nil {
			s.logger.Warnf("%s", err)
		} else if fastlaneDebugInfo != "" {
			if err := fileutil.WriteStringToFile(deployPth, fastlaneDebugInfo); err != nil {
				s.logger.Warnf("Failed to write fastlane env log file, error: %s", err)
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
			s.logger.Errorf("Failed to walk directory, error: %s", err)
		}
		return fmt.Errorf("Command failed, error: %s", err)
	}

	return nil
}

func (s FastlaneRunner) fastlaneDebugInfo(workDir string, useBundler bool, bundlerVersion gems.Version) (string, error) {
	factory, err := ruby.NewCommandFactory(command.NewFactory(env.NewRepository()), env.NewCommandLocator())
	if err != nil {
		return "", err
	}

	name := "fastlane"
	args := []string{"env"}
	var outBuffer bytes.Buffer
	opts := &command.Opts{
		Stdin:  strings.NewReader("n"),
		Stdout: bufio.NewWriter(&outBuffer),
		Stderr: os.Stderr,
		Dir:    workDir,
	}
	var cmd command.Command
	if useBundler {
		cmd = factory.CreateBundleExec(name, args, bundlerVersion.Version, opts)
	} else {
		cmd = factory.Create(name, args, opts)
	}

	s.logger.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return "", fmt.Errorf("Fastlane command (%s) failed, output: %s", cmd.PrintableCommandArgs(), outBuffer.String())
		}
		return "", fmt.Errorf("Fastlane command (%s) failed: %v", cmd.PrintableCommandArgs(), err)
	}

	return outBuffer.String(), nil
}
