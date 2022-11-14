package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

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
			Dir:    opts.WorkDir,
		})
		for _, cmd := range cmds {
			f.logger.Donef("$ %s", cmd.PrintableCommandArgs())
			f.logger.Println()

			if err := cmd.Run(); err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					return fmt.Errorf("command failed with exit status %d (%s): %w", exitErr.ExitCode(), cmd.PrintableCommandArgs(), errors.New("check the command's output for details"))
				}

				return fmt.Errorf("executing command failed (%s): %w", cmd.PrintableCommandArgs(), err)
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
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return fmt.Errorf("command failed with exit status %d (%s): %w", exitErr.ExitCode(), cmd.PrintableCommandArgs(), errors.New("check the command's output for details"))
			}

			return fmt.Errorf("executing command failed (%s): %w", cmd.PrintableCommandArgs(), err)
		}
	} else if opts.UpdateFastlane {
		f.logger.Infof("Update system installed Fastlane")

		cmds := f.rbyFactory.CreateGemInstall("fastlane", "", false, false, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Dir:    opts.WorkDir,
		})
		for _, cmd := range cmds {
			f.logger.Donef("$ %s", cmd.PrintableCommandArgs())

			if err := cmd.Run(); err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					return fmt.Errorf("command failed with exit status %d (%s): %w", exitErr.ExitCode(), cmd.PrintableCommandArgs(), errors.New("check the command's output for details"))
				}

				return fmt.Errorf("executing command failed (%s): %w", cmd.PrintableCommandArgs(), err)
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
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("command failed with exit status %d (%s): %w", exitErr.ExitCode(), cmd.PrintableCommandArgs(), errors.New("check the command's output for details"))
		}

		return fmt.Errorf("executing command failed (%s): %w", cmd.PrintableCommandArgs(), err)
	}

	return nil
}
