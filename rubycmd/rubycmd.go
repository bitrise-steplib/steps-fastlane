package rubycmd

import (
	"errors"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/log"
)

const (
	systemRubyPth = "/usr/bin/ruby"
	brewRubyPth   = "/usr/local/bin/ruby"
)

// ----------------------
// RubyCommand

// RubyInstallType ...
type RubyInstallType int8

const (
	// SystemRuby ...
	SystemRuby RubyInstallType = iota
	// BrewRuby ...
	BrewRuby
	// RVMRuby ...
	RVMRuby
	// RbenvRuby ...
	RbenvRuby
)

// RubyCommandModel ...
type RubyCommandModel struct {
	rubyInstallType RubyInstallType
}

// NewRubyCommandModel ...
func NewRubyCommandModel() (RubyCommandModel, error) {

	whichRuby, err := cmdex.NewCommand("which", "ruby").RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return RubyCommandModel{}, err
	}

	command := RubyCommandModel{}

	if whichRuby == systemRubyPth {
		command.rubyInstallType = SystemRuby
	} else if whichRuby == brewRubyPth {
		command.rubyInstallType = BrewRuby
	} else if cmdExist([]string{"rvm", "-v"}) {
		command.rubyInstallType = RVMRuby
	} else if cmdExist([]string{"rbenv", "-v"}) {
		command.rubyInstallType = RbenvRuby
	} else {
		return RubyCommandModel{}, errors.New("unkown ruby installation type")
	}

	return command, nil
}

func (command RubyCommandModel) sudoNeeded(cmdSlice []string) bool {
	if command.rubyInstallType != SystemRuby {
		return false
	}

	if len(cmdSlice) < 2 {
		return false
	}

	isGemManagementCmd := (cmdSlice[0] == "gem" || cmdSlice[0] == "bundle")
	isInstallOrUnintsallCmd := (cmdSlice[1] == "install" || cmdSlice[1] == "uninstall")

	return (isGemManagementCmd && isInstallOrUnintsallCmd)
}

// Execute ...
func (command RubyCommandModel) Execute(workDir string, useBundle bool, cmdSlice []string) error {
	if useBundle {
		cmdSlice = append([]string{"bundle", "exec"}, cmdSlice...)
	}

	if command.sudoNeeded(cmdSlice) {
		cmdSlice = append([]string{"sudo"}, cmdSlice...)
	}

	return execute(workDir, false, cmdSlice)
}

// ExecuteForOutput ...
func (command RubyCommandModel) ExecuteForOutput(workDir string, useBundle bool, cmdSlice []string) (string, error) {
	if useBundle {
		cmdSlice = append([]string{"bundle", "exec"}, cmdSlice...)
	}

	if command.sudoNeeded(cmdSlice) {
		cmdSlice = append([]string{"sudo"}, cmdSlice...)
	}

	return executeForOutput(workDir, false, cmdSlice)
}

// GemUpdate ...
func (command RubyCommandModel) GemUpdate(gem string) error {
	cmdSlice := []string{"gem", "update", gem, "--no-document"}

	if err := command.Execute("", false, cmdSlice); err != nil {
		return err
	}

	if command.rubyInstallType == RbenvRuby {
		cmdSlice := []string{"rbenv", "rehash"}

		if err := command.Execute("", false, cmdSlice); err != nil {
			return err
		}
	}

	return nil
}

// GemInstall ...
func (command RubyCommandModel) GemInstall(gem, version string) error {
	cmdSlice := []string{"gem", "install", gem, "--no-document"}
	if version != "" {
		cmdSlice = append(cmdSlice, "-v", version)
	}

	if err := command.Execute("", false, cmdSlice); err != nil {
		return err
	}

	if command.rubyInstallType == RbenvRuby {
		cmdSlice := []string{"rbenv", "rehash"}

		if err := command.Execute("", false, cmdSlice); err != nil {
			return err
		}
	}

	return nil
}

// IsGemInstalled ...
func (command RubyCommandModel) IsGemInstalled(gem, version string) (bool, error) {
	cmdSlice := []string{"gem", "list"}
	out, err := command.ExecuteForOutput("", false, cmdSlice)
	if err != nil {
		return false, err
	}

	regexpStr := gem + ` \((?P<versions>.*)\)`
	exp := regexp.MustCompile(regexpStr)
	matches := exp.FindStringSubmatch(out)
	if len(matches) > 1 {
		if version == "" {
			return true, nil
		}

		versionsStr := matches[1]
		versions := strings.Split(versionsStr, ", ")

		for _, v := range versions {
			if v == version {
				return true, nil
			}
		}
	}

	return false, nil
}

// ----------------------
// Common

func cmdExist(cmdSlice []string) bool {
	if len(cmdSlice) == 0 {
		return false
	}

	cmd, err := cmdex.NewCommandFromSlice(cmdSlice)
	if err != nil {
		return false
	}

	err = cmd.Run()
	return (err == nil)
}

func execute(workDir string, bundleExec bool, cmdSlice []string) error {
	if len(cmdSlice) == 0 {
		return errors.New("no command specified")
	}

	if bundleExec {
		cmdSlice = append([]string{"bundle", "exec"}, cmdSlice...)
	}

	prinatableCmd := cmdex.PrintableCommandArgs(false, cmdSlice)
	log.Donef("$ %s", prinatableCmd)

	cmd, err := cmdex.NewCommandFromSlice(cmdSlice)
	if err != nil {
		return err
	}

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	log.Printf(out)

	return err
}

func executeForOutput(workDir string, bundleExec bool, cmdSlice []string) (string, error) {
	if len(cmdSlice) == 0 {
		return "", errors.New("no command specified")
	}

	if bundleExec {
		cmdSlice = append([]string{"bundle", "exec"}, cmdSlice...)
	}

	cmd, err := cmdex.NewCommandFromSlice(cmdSlice)
	if err != nil {
		return "", err
	}

	return cmd.RunAndReturnTrimmedCombinedOutput()
}
