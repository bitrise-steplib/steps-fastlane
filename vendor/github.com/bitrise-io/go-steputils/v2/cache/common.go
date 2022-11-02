package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// We need this prefix because there could be multiple restore steps in one workflow with multiple cache keys
const cacheHitEnvVarPrefix = "BITRISE_CACHE_HIT__"

func checksumOfFile(path string) (string, error) {
	hash := sha256.New()

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close() //nolint:errcheck

	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
