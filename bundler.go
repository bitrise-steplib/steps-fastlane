package main

import (
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/rubycommand"
)

func getInstallBundlerCommand(gemfileLockVersion gemVersion) (*command.Model, error) {
	installBundlerCmdParams := []string{"gem", "install", "bundler", "--force", "--no-document"}
	if gemfileLockVersion.found {
		installBundlerCmdParams = append(installBundlerCmdParams, []string{"-v", gemfileLockVersion.version}...)
	}

	return command.NewFromSlice(installBundlerCmdParams)
}

func getBundleInstallCommand(gemfileLockVersion gemVersion) (*command.Model, error) {
	bundleInstallCmdParams := []string{"bundle"}
	if gemfileLockVersion.found {
		bundleInstallCmdParams = append(bundleInstallCmdParams, "_"+gemfileLockVersion.version+"_")
	}
	bundleInstallCmdParams = append(bundleInstallCmdParams, []string{"install", "--jobs", "20", "--retry", "5"}...)

	return rubycommand.NewFromSlice(bundleInstallCmdParams)
}

func getRbenvVersionsCommand() (*command.Model, error) {
	if _, err := command.New("which", "rbenv").RunAndReturnTrimmedCombinedOutput(); err != nil {
		return nil, err
	}

	return command.New("rbenv", "versions"), nil
}
