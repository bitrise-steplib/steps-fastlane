package cache

import (
	"io/fs"
	"time"

	"github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
)

type stepTracker struct {
	tracker analytics.Tracker
	logger  log.Logger
}

func newStepTracker(stepId string, envRepo env.Repository, logger log.Logger) stepTracker {
	p := analytics.Properties{
		"step_id":     stepId,
		"build_slug":  envRepo.Get("BITRISE_BUILD_SLUG"),
		"app_slug":    envRepo.Get("BITRISE_APP_SLUG"),
		"workflow":    envRepo.Get("BITRISE_TRIGGERED_WORKFLOW_ID"),
		"is_pr_build": envRepo.Get("IS_PR") == "true",
	}
	return stepTracker{
		tracker: analytics.NewDefaultTracker(logger, p),
		logger:  logger,
	}
}

func (t *stepTracker) logArchiveUploaded(uploadTime time.Duration, info fs.FileInfo, pathCount int) {
	properties := analytics.Properties{
		"upload_time_s":     uploadTime.Truncate(time.Second).Seconds(),
		"upload_size_bytes": info.Size(),
		"path_count":        pathCount,
	}
	t.tracker.Enqueue("step_save_cache_archive_uploaded", properties)
}

func (t *stepTracker) logArchiveCompressed(compressionTime time.Duration, pathCount int) {
	properties := analytics.Properties{
		"compression_time_s": compressionTime.Truncate(time.Second).Seconds(),
		"path_count":         pathCount,
	}
	t.tracker.Enqueue("step_save_cache_archive_compressed", properties)
}

func (t *stepTracker) logArchiveDownloaded(downloadTime time.Duration, info fs.FileInfo, keyCount int) {
	properties := analytics.Properties{
		"download_time_s":     downloadTime.Truncate(time.Second).Seconds(),
		"download_size_bytes": info.Size(),
		"key_count":           keyCount,
	}
	t.tracker.Enqueue("step_restore_cache_archive_downloaded", properties)
}

func (t *stepTracker) logArchiveExtracted(extractionTime time.Duration, keyCount int) {
	properties := analytics.Properties{
		"extraction_time_s": extractionTime.Truncate(time.Second).Seconds(),
		"key_count":         keyCount,
	}
	t.tracker.Enqueue("step_restore_cache_archive_extracted", properties)
}

func (t *stepTracker) logRestoreResult(isMatch bool, matchedKey string, evaluatedKeys []string) {
	if len(evaluatedKeys) == 0 {
		return
	}

	properties := analytics.Properties{
		"is_match":             isMatch,
		"is_first_key_matched": matchedKey == evaluatedKeys[0],
		"key_count":            len(evaluatedKeys),
	}
	t.tracker.Enqueue("step_restore_cache_result", properties)
}

func (t *stepTracker) logSkipSaveResult(isSaveSkipped bool, reason skipReason) {

	properties := analytics.Properties{
		"is_save_skipped": isSaveSkipped,
		"reason":          reason.String(),
	}
	t.tracker.Enqueue("step_save_cache_save_skipped", properties)
}

func (t *stepTracker) logSkipUploadResult(isUploadSkipped bool, reason skipReason) {

	properties := analytics.Properties{
		"is_upload_skipped": isUploadSkipped,
		"reason":            reason.String(),
	}
	t.tracker.Enqueue("step_save_cache_upload_skipped", properties)
}

func (t *stepTracker) wait() {
	t.tracker.Wait()
}
