package main

import (
	"fmt"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/rubycommand"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
)

func installBundler(gemfileLockVersion gemVersion) error {
	installBundlerCmdParams := []string{"gem", "install", "bundler"}
	if gemfileLockVersion.found {
		installBundlerCmdParams = append(installBundlerCmdParams, []string{"-v", gemfileLockVersion.version}...)
	}

	log.Debugf("$ %s", installBundlerCmdParams)
	fmt.Println()

	installBundlerCommand, err := command.NewFromSlice(installBundlerCmdParams)
	if err != nil {
		return err
	}

	out, err := installBundlerCommand.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("gem command exited with error: %s, out: %s", err, out)
		}
		return fmt.Errorf("failed to run gem command, error: %s", err)
	}

	return nil
}

func getBundleInstallCommand(gemfileLockVersion gemVersion) (*command.Model, error) {
	bundleInstallCmdParams := []string{"bundle"}
	if gemfileLockVersion.found {
		bundleInstallCmdParams = append(bundleInstallCmdParams, "_"+gemfileLockVersion.version+"_")
	}
	bundleInstallCmdParams = append(bundleInstallCmdParams, []string{"install", "--jobs", "20", "--retry", "5"}...)

	return rubycommand.NewFromSlice(bundleInstallCmdParams)
}
