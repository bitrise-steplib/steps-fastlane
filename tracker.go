package main

import (
	"github.com/bitrise-io/go-utils/v2/analytics"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
)

type stepTracker struct {
	tracker analytics.Tracker
	logger  log.Logger
}

func newStepTracker(envRepo env.Repository, logger log.Logger) stepTracker {
	p := analytics.Properties{
		"build_slug":        envRepo.Get("BITRISE_BUILD_SLUG"),
		"step_execution_id": envRepo.Get("BITRISE_STEP_EXECUTION_ID"),
		"step_id":           "fastlane",
	}
	return stepTracker{
		tracker: analytics.NewDefaultTracker(logger, p),
		logger:  logger,
	}
}

func (t *stepTracker) logEffectiveRubyVersion(versionString string) {
	properties := analytics.Properties{
		"effective_ruby_version": versionString,
	}
	t.tracker.Enqueue("step_ruby_version_selected", properties)
}

func (t *stepTracker) wait() {
	t.tracker.Wait()
}
