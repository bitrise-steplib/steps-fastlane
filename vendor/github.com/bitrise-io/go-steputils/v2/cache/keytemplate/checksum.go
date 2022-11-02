package keytemplate

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bmatcuk/doublestar/v4"
)

// checksum returns a hex-encoded SHA-256 checksum of one or multiple files. Each file path can contain glob patterns,
// including "doublestar" patterns (such as `**/*.gradle`).
// The path list is sorted alphabetically to produce consistent output.
// Errors are logged as warnings and an empty string is returned in that case.
func (m Model) checksum(paths ...string) string {
	files := m.evaluateGlobPatterns(paths)
	m.logger.Debugf("Files included in checksum:")
	for _, path := range files {
		m.logger.Debugf("- %s", path)
	}

	if len(files) == 0 {
		m.logger.Warnf("No files to include in the checksum")
		return ""
	} else if len(files) == 1 {
		checksum, err := checksumOfFile(files[0])
		if err != nil {
			m.logger.Warnf("Error while computing checksum %s: %s", files[0], err)
			return ""
		}
		return hex.EncodeToString(checksum)
	}

	finalChecksum := sha256.New()
	sort.Strings(files)
	for _, path := range files {
		checksum, err := checksumOfFile(path)
		if err != nil {
			m.logger.Warnf("Error while hashing %s: %s", path, err)
			continue
		}

		finalChecksum.Write(checksum)
	}

	return hex.EncodeToString(finalChecksum.Sum(nil))
}

func (m Model) evaluateGlobPatterns(paths []string) []string {
	var finalPaths []string

	for _, path := range paths {
		if strings.Contains(path, "*") {
			base, pattern := doublestar.SplitPattern(path)
			absBase, err := pathutil.NewPathModifier().AbsPath(base)
			if err != nil {
				m.logger.Warnf("Failed to convert %s to an absolute path: %s", path, err)
				continue
			}
			matches, err := doublestar.Glob(os.DirFS(absBase), pattern)
			if matches == nil {
				m.logger.Warnf("No match for pattern: %s", path)
				continue
			}
			if err != nil {
				m.logger.Warnf("Error in pattern '%s': %s", path, err)
				continue
			}
			for _, match := range matches {
				finalPaths = append(finalPaths, filepath.Join(base, match))
			}
		} else {
			finalPaths = append(finalPaths, path)
		}
	}

	return filterFilesOnly(finalPaths)
}

func checksumOfFile(path string) ([]byte, error) {
	hash := sha256.New()
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:errcheck

	_, err = io.Copy(hash, file)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func filterFilesOnly(paths []string) []string {
	var files []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		files = append(files, path)
	}

	return files
}
