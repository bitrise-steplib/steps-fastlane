package main

import (
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
)

// EnsureDependenciesOpts ...
type EnsureDependenciesOpts struct {
	GemVersions    gemVersions
	UseBundler     bool
	WorkDir        string
	UpdateFastlane bool
}

// InstallDependencies ...
func (f FastlaneRunner) InstallDependencies(opts EnsureDependenciesOpts) error {
	f.reportRubyVersion(opts.UseBundler, opts.GemVersions.bundler.Version, opts.WorkDir)

	// Install desired Fastlane version
	if opts.UseBundler {
		f.logger.Println()
		f.logger.Infof("Install bundler")

		// install bundler with `gem install bundler [-v version]`
		// in some configurations, the command "bundler _1.2.3_" can return 'Command not found', installing bundler solves this
		cmds := f.rbyFactory.CreateGemInstall("bundler", opts.GemVersions.bundler.Version, false, true, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Dir:    opts.WorkDir,
		})
		for _, cmd := range cmds {
			f.logger.Donef("$ %s", cmd.PrintableCommandArgs())
			f.logger.Println()

			if err := cmd.Run(); err != nil {
				return err
			}
		}

		// install Gemfile.lock gems with `bundle [_version_] install ...`
		f.logger.Println()
		f.logger.Infof("Install Fastlane with bundler")

		cmd := f.rbyFactory.CreateBundleInstall(opts.GemVersions.bundler.Version, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Dir:    opts.WorkDir,
		})

		f.logger.Donef("$ %s", cmd.PrintableCommandArgs())
		f.logger.Println()

		if err := cmd.Run(); err != nil {
			return err
		}
	} else if opts.UpdateFastlane {
		f.logger.Println()
		f.logger.Infof("Update system installed Fastlane")

		cmds := f.rbyFactory.CreateGemInstall("fastlane", "", false, false, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Dir:    opts.WorkDir,
		})
		for _, cmd := range cmds {
			f.logger.Donef("$ %s", cmd.PrintableCommandArgs())

			if err := cmd.Run(); err != nil {
				return err
			}
		}
	} else {
		f.logger.Println()
		f.logger.Infof("Using system installed Fastlane")
	}

	f.logger.Println()
	f.logger.Infof("Fastlane version")

	name := "fastlane"
	args := []string{"--version"}
	options := &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Dir:    opts.WorkDir,
	}
	var cmd command.Command
	if opts.UseBundler {
		cmd = f.rbyFactory.CreateBundleExec(name, args, opts.GemVersions.bundler.Version, options)
	} else {
		cmd = f.rbyFactory.Create(name, args, options)
	}

	f.logger.Donef("$ %s", cmd.PrintableCommandArgs())

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (f FastlaneRunner) reportRubyVersion(useBundler bool, bundlerVersion string, workDir string) {
	var versionCmd command.Command
	options := &command.Opts{
		Dir: workDir,
	}
	if useBundler {
		versionCmd = f.rbyFactory.CreateBundleExec("ruby", []string{"--version"}, bundlerVersion, options)
	} else {
		versionCmd = f.rbyFactory.Create("ruby", []string{"--version"}, options)
	}
	output, err := versionCmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		f.logger.Warnf("Failed to check active Ruby version: %s", err)
		f.logger.Printf("Output: %s", output)
		return
	}
	// Example output:
	// ruby 3.2.1 (2023-02-08 revision 31819e82c8) [arm64-darwin22]
	versionSlice := strings.Split(output, " ")
	if len(versionSlice) < 2 || versionSlice[0] != "ruby" {
		f.logger.Warnf("Unrecognized Ruby version: %s", versionSlice)
	}
	version := versionSlice[1]

	f.logger.Println()
	f.logger.Infof("Active Ruby version: %s", version)

	f.tracker.logEffectiveRubyVersion(version)
}
