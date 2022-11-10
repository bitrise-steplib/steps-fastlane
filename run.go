package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/command/gems"
	"github.com/bitrise-io/go-steputils/v2/ruby"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-xcode/appleauth"
)

// RunOpts ...
type RunOpts struct {
	WorkDir         string
	AuthCredentials appleauth.Credentials
	LaneOptions     []string
	UseBundler      bool
	GemVersions     gemVersions
	EnableCache     bool
}

// Run ...
func (f FastlaneRunner) Run(opts RunOpts) error {
	// Run fastlane
	f.logger.Println()
	f.logger.Infof("Run Fastlane")

	var envs []string
	authEnvs, err := FastlaneAuthParams(opts.AuthCredentials)
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
		f.logger.Warnf("Fastlane authentication-related environment varibale(s) (%s) are set, overriding.", globallySetAuthEnvs)
		f.logger.Infof("To stop overriding authentication-related environment variables, please set Bitrise Apple Developer Connection input to 'off' and leave authentication-related inputs empty.")
	}

	buildlogPth := ""
	if tempDir, err := pathutil.NormalizedOSTempDirPath("fastlane_logs"); err != nil {
		f.logger.Errorf("Failed to create temp dir for fastlane logs, error: %s", err)
	} else {
		buildlogPth = tempDir
		envs = append(envs, "FL_BUILDLOG_PATH="+buildlogPth)
	}

	name := "fastlane"
	args := opts.LaneOptions
	options := &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Dir:    opts.WorkDir,
		Env:    append(os.Environ(), envs...),
	}
	var cmd command.Command
	if opts.UseBundler {
		cmd = f.rbyFactory.CreateBundleExec(name, args, opts.GemVersions.bundler.Version, options)
	} else {
		cmd = f.rbyFactory.Create(name, args, options)
	}

	f.logger.Donef("$ %s", cmd.PrintableCommandArgs())

	deployDir := os.Getenv("BITRISE_DEPLOY_DIR")
	if deployDir == "" {
		f.logger.Warnf("No BITRISE_DEPLOY_DIR found")
	}
	deployPth := filepath.Join(deployDir, "fastlane_env.log")

	if err := cmd.Run(); err != nil {
		f.logger.Println()
		f.logger.Errorf("Fastlane command: (%s) failed", cmd.PrintableCommandArgs())
		f.logger.Errorf("If you want to send an issue report to fastlane (https://github.com/fastlane/fastlane/issues/new), you can find the output of fastlane env in the following log file:")
		f.logger.Println()
		f.logger.Infof(deployPth)
		f.logger.Println()

		if fastlaneDebugInfo, err := f.fastlaneDebugInfo(opts.WorkDir, opts.UseBundler, opts.GemVersions.bundler); err != nil {
			f.logger.Warnf("%s", err)
		} else if fastlaneDebugInfo != "" {
			if err := fileutil.WriteStringToFile(deployPth, fastlaneDebugInfo); err != nil {
				f.logger.Warnf("Failed to write fastlane env log file, error: %s", err)
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
			f.logger.Errorf("Failed to walk directory, error: %s", err)
		}
		return fmt.Errorf("command failed with %s (%s)", err, cmd.PrintableCommandArgs())
	}

	f.cacheDeps(opts)

	return nil
}

func (f FastlaneRunner) fastlaneDebugInfo(workDir string, useBundler bool, bundlerVersion gems.Version) (string, error) {
	factory, err := ruby.NewCommandFactory(f.cmdFactory, f.cmdLocator)
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

	f.logger.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return "", fmt.Errorf("Fastlane command (%s) failed, output: %s", cmd.PrintableCommandArgs(), outBuffer.String())
		}
		return "", fmt.Errorf("Fastlane command (%s) failed: %v", cmd.PrintableCommandArgs(), err)
	}

	return outBuffer.String(), nil
}
