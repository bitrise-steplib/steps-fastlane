package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
)

func formattedError(err error) string {
	var formatted string

	i := -1
	for {
		i++

		reason := err.Error()

		if err = errors.Unwrap(err); err == nil {
			formatted = appendError(formatted, reason, i, true)
			return formatted
		}

		reason = strings.TrimSuffix(reason, err.Error())
		reason = strings.TrimRight(reason, " ")
		reason = strings.TrimSuffix(reason, ":")

		formatted = appendError(formatted, reason, i, false)
	}
}

func appendError(errorMessage, reason string, i int, last bool) string {
	if i == 0 {
		errorMessage = indentedReason(reason, i)
	} else {
		errorMessage += "\n"
		errorMessage += indentedReason(reason, i)
	}

	if !last {
		errorMessage += ":"
	}

	return errorMessage
}

func indentedReason(reason string, level int) string {
	var lines []string
	split := strings.Split(reason, "\n")
	for _, line := range split {
		line = strings.TrimLeft(line, " ")
		line = strings.TrimRight(line, "\n")
		line = strings.TrimRight(line, " ")
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	var indented string
	for i, line := range lines {
		indented += strings.Repeat("  ", level)
		indented += line
		if i != len(lines)-1 {
			indented += "\n"
		}
	}
	return indented
}

func (f FastlaneRunner) formattedCommandErrorMessage(cmd command.Command, err error) string {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Sprintf("command failed with exit status %d (%s): %v", exitErr.ExitCode(), cmd.PrintableCommandArgs(), errors.New("check the command's output for details"))
	}

	return fmt.Sprintf("executing command failed (%s): %v", cmd.PrintableCommandArgs(), err)
}
