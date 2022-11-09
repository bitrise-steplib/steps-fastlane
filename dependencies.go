package main

import (
	"fmt"
	"os"

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
	// Install desired Fastlane version
	if opts.UseBundler {
		f.logger.Infof("Install bundler")

		// install bundler with `gem install bundler [-v version]`
		// in some configurations, the command "bundler _1.2.3_" can return 'Command not found', installing bundler solves this
		cmds := f.rbyFactory.CreateGemInstall("bundler", opts.GemVersions.bundler.Version, false, true, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Stdin:  nil,
			Env:    []string{},
			Dir:    opts.WorkDir,
		})
		for _, cmd := range cmds {
			f.logger.Donef("$ %s", cmd.PrintableCommandArgs())
			f.logger.Println()

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("command failed with %s (%s)", err, cmd.PrintableCommandArgs())
			}
		}

		// install Gemfile.lock gems with `bundle [_version_] install ...`
		f.logger.Println()
		f.logger.Infof("Install Fastlane with bundler")

		cmd := f.rbyFactory.CreateBundleInstall(opts.GemVersions.bundler.Version, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Stdin:  nil,
			Env:    []string{},
			Dir:    opts.WorkDir,
		})

		f.logger.Donef("$ %s", cmd.PrintableCommandArgs())
		f.logger.Println()

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed with %s (%s)", err, cmd.PrintableCommandArgs())
		}
	} else if opts.UpdateFastlane {
		f.logger.Infof("Update system installed Fastlane")

		cmds := f.rbyFactory.CreateGemInstall("fastlane", "", false, false, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Stdin:  nil,
			Env:    []string{},
			Dir:    opts.WorkDir,
		})
		for _, cmd := range cmds {
			f.logger.Donef("$ %s", cmd.PrintableCommandArgs())

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("command failed with %s (%s)", err, cmd.PrintableCommandArgs())
			}
		}
	} else {
		f.logger.Infof("Using system installed Fastlane")
	}

	f.logger.Println()
	f.logger.Infof("Fastlane version")

	name := "fastlane"
	args := []string{"--version"}
	options := &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  nil,
		Env:    []string{},
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
		return fmt.Errorf("command failed with %s (%s)", err, cmd.PrintableCommandArgs())
	}

	return nil
}
