package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/command"
)

// EnsureDependenciesOpts ...
type EnsureDependenciesOpts struct {
	GemVersions gemVersions
	UseBundler  bool
}

// InstallDependencies ...
func (s FastlaneRunner) InstallDependencies(config Config, opts EnsureDependenciesOpts) error {
	// Install desired Fastlane version
	if opts.UseBundler {
		s.logger.Infof("Install bundler")

		// install bundler with `gem install bundler [-v version]`
		// in some configurations, the command "bundler _1.2.3_" can return 'Command not found', installing bundler solves this
		cmds := s.rbyFactory.CreateGemInstall("bundler", opts.GemVersions.bundler.Version, false, true, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Stdin:  nil,
			Env:    []string{},
			Dir:    config.WorkDir,
		})
		for _, cmd := range cmds {
			s.logger.Donef("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("command failed, error: %s", err)
			}
		}

		// install Gemfile.lock gems with `bundle [_version_] install ...`
		fmt.Println()
		s.logger.Infof("Install Fastlane with bundler")

		cmd := s.rbyFactory.CreateBundleInstall(opts.GemVersions.bundler.Version, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Stdin:  nil,
			Env:    []string{},
			Dir:    config.WorkDir,
		})

		s.logger.Donef("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Command failed, error: %s", err)
		}
	} else if config.UpdateFastlane {
		s.logger.Infof("Update system installed Fastlane")

		cmds := s.rbyFactory.CreateGemInstall("fastlane", "", false, false, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Stdin:  nil,
			Env:    []string{},
			Dir:    config.WorkDir,
		})
		for _, cmd := range cmds {
			s.logger.Donef("$ %s", cmd.PrintableCommandArgs())

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Command failed, error: %s", err)
			}
		}
	} else {
		s.logger.Infof("Using system installed Fastlane")
	}

	fmt.Println()
	s.logger.Infof("Fastlane version")

	name := "fastlane"
	args := []string{"--version"}
	options := &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  nil,
		Env:    []string{},
		Dir:    config.WorkDir,
	}
	var cmd command.Command
	if opts.UseBundler {
		cmd = s.rbyFactory.CreateBundleExec(name, args, opts.GemVersions.bundler.Version, options)
	} else {
		cmd = s.rbyFactory.Create(name, args, options)
	}

	s.logger.Donef("$ %s", cmd.PrintableCommandArgs())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Command failed, error: %s", err)
	}

	return nil
}
