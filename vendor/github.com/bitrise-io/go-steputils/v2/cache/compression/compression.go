package compression

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
)

// Compress creates a compressed archive from the provided files and folders using absolute paths.
func Compress(archivePath string, includePaths []string, logger log.Logger, envRepo env.Repository) error {
	cmdFactory := command.NewFactory(envRepo)

	/*
		tar arguments:
		--use-compress-program: Pipe the output to zstd instead of using the built-in gzip compression
		-P: Alias for --absolute-paths in BSD tar and --absolute-names in GNU tar (step runs on both Linux and macOS)
			Storing absolute paths in the archive allows paths outside the current directory (such as ~/.gradle)
		-c: Create archive
		-f: Output file
	*/
	tarArgs := []string{
		"--use-compress-program", "zstd --threads=0", // Use CPU count threads
		"-P",
		"-c",
		"-f", archivePath,
	}
	tarArgs = append(tarArgs, includePaths...)

	cmd := cmdFactory.Create("tar", tarArgs, nil)

	logger.Debugf("$ %s", cmd.PrintableCommandArgs())

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("command failed with exit status %d (%s):\n%w", exitErr.ExitCode(), cmd.PrintableCommandArgs(), errors.New(out))
		}
		return fmt.Errorf("executing command failed (%s): %w", cmd.PrintableCommandArgs(), err)
	}

	return nil
}

// Decompress takes an archive path and extracts files. This assumes an archive created with absolute file paths.
func Decompress(archivePath string, logger log.Logger, envRepo env.Repository, additionalArgs ...string) error {
	commandFactory := command.NewFactory(envRepo)

	/*
		tar arguments:
		--use-compress-program: Pipe the input to zstd instead of using the built-in gzip compression
		-P: Alias for --absolute-paths in BSD tar and --absolute-names in GNU tar (step runs on both Linux and macOS)
			Storing absolute paths in the archive allows paths outside the current directory (such as ~/.gradle)
		-x: Extract archive
		-f: Output file
	*/
	decompressTarArgs := []string{
		"--use-compress-program", "zstd -d",
		"-x",
		"-f", archivePath,
		"-P",
	}

	if len(additionalArgs) > 0 {
		decompressTarArgs = append(decompressTarArgs, additionalArgs...)
	}

	cmd := commandFactory.Create("tar", decompressTarArgs, nil)
	logger.Debugf("$ %s", cmd.PrintableCommandArgs())

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("command failed with exit status %d (%s):\n%w", exitErr.ExitCode(), cmd.PrintableCommandArgs(), errors.New(out))
		}
		return fmt.Errorf("executing command failed (%s): %w", cmd.PrintableCommandArgs(), err)
	}

	return nil
}

// AreAllPathsEmpty checks if the provided paths are all nonexistent files or empty directories
func AreAllPathsEmpty(includePaths []string) bool {
	allEmpty := true

	for _, path := range includePaths {
		// Check if file exists at path
		fileInfo, err := os.Stat(path)
		if errors.Is(err, fs.ErrNotExist) {
			// File doesn't exist
			continue
		}

		// Check if it's a directory
		if !fileInfo.IsDir() {
			// Is a file and it exists
			allEmpty = false
			break
		}

		file, err := os.Open(path)
		if err != nil {
			continue
		}
		_, err = file.Readdirnames(1) // query only 1 child
		if errors.Is(err, io.EOF) {
			// Dir is empty
			continue
		}
		if err == nil {
			// Dir has files or dirs
			allEmpty = false
			break
		}
	}

	return allEmpty
}
