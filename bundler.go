package main

import (
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/rubycommand"
)

func installBundler(gemfileLockVersion gemVersion) (*command.Model, error) {
	installBundlerCmdParams := []string{"gem", "install", "bundler", "--force"}
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
