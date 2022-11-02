package network

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/retryhttp"
)

// DownloadParams ...
type DownloadParams struct {
	APIBaseURL   string
	Token        string
	CacheKeys    []string
	DownloadPath string
}

// ErrCacheNotFound ...
var ErrCacheNotFound = errors.New("no cache archive found for the provided keys")

// Download archive from the cache API based on the provided keys in params.
// If there is no match for any of the keys, the error is ErrCacheNotFound.
func Download(params DownloadParams, logger log.Logger) (matchedKey string, err error) {
	if params.APIBaseURL == "" {
		return "", fmt.Errorf("API base URL is empty")
	}

	if params.Token == "" {
		return "", fmt.Errorf("API token is empty")
	}

	if len(params.CacheKeys) == 0 {
		return "", fmt.Errorf("cache key list is empty")
	}

	client := newAPIClient(retryhttp.NewClient(logger), params.APIBaseURL, params.Token)

	logger.Debugf("Get download URL")
	restoreResponse, err := client.restore(params.CacheKeys)
	if err != nil {
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}

	logger.Debugf("Download archive")
	file, err := os.Create(params.DownloadPath)
	if err != nil {
		return "", fmt.Errorf("can't open download location: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Warnf(err.Error())
		}
	}(file)

	respBody, err := client.downloadArchive(restoreResponse.URL)
	if err != nil {
		return "", fmt.Errorf("failed to download archive: %w", err)
	}
	defer func(respBody io.ReadCloser) {
		err := respBody.Close()
		if err != nil {
			logger.Warnf(err.Error())
		}
	}(respBody)
	_, err = io.Copy(file, respBody)
	if err != nil {
		return "", fmt.Errorf("failed to save archive to disk: %w", err)
	}

	return restoreResponse.MatchedKey, nil
}
