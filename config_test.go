package main

import (
	"testing"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/appleauth"
	"github.com/stretchr/testify/assert"
)

func Test_GivenAutomaticConnection_WhenParseAuthSources_ThenReceiveAllSources(t *testing.T) {
	step := FastlaneRunner{}
	expectedValue := []appleauth.Source{
		&appleauth.ConnectionAPIKeySource{},
		&appleauth.ConnectionAppleIDFastlaneSource{},
		&appleauth.InputAPIKeySource{},
		&appleauth.InputAppleIDFastlaneSource{},
	}

	actualValue, err := step.parseAuthSources(automatic)

	assert.NoError(t, err)
	assert.Equal(t, actualValue, expectedValue)
}

func Test_GivenAPIKeyConnection_WhenParseAuthSources_ThenReceiveConnectionAPIKeySource(t *testing.T) {
	step := FastlaneRunner{}
	expectedValue := []appleauth.Source{
		&appleauth.ConnectionAPIKeySource{},
	}

	actualValue, err := step.parseAuthSources(apiKey)

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, actualValue)
}

func Test_GivenAppleIDConnection_WhenParseAuthSources_ThenReceiveAppleIDFastlaneSource(t *testing.T) {
	step := FastlaneRunner{}
	expectedValue := []appleauth.Source{
		&appleauth.ConnectionAppleIDFastlaneSource{},
	}

	actualValue, err := step.parseAuthSources(appleID)

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, actualValue)
}

func Test_GivenOffConnection_WhenParseAuthSources_ThenReceiveInputAPIKeyAndAppleIdFastlaneSources(t *testing.T) {
	step := FastlaneRunner{}
	expectedValue := []appleauth.Source{

		&appleauth.InputAPIKeySource{},
		&appleauth.InputAppleIDFastlaneSource{},
	}

	actualValue, err := step.parseAuthSources(off)

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, actualValue)
}

func Test_GivenGemHomeEnvironmentVariableIsEmpty_WhenValidateGemHome_ThenLogNothing(t *testing.T) {
	mockLogger := MockLogger{}
	step := FastlaneRunner{logger: &mockLogger}
	expectedGemHome := ""
	inputs := Inputs{GemHome: expectedGemHome}
	config := Config{Inputs: inputs}

	step.validateGemHome(config)

	mockLogger.AssertNotCalled(t, "Warnf")
}

func Test_GivenGemHomeEnvironmentVariableIsEmpty_WhenValidateGemHome_ThenLogWarning(t *testing.T) {
	var mockedLogger MockLogger
	step := FastlaneRunner{logger: &mockedLogger}
	expectedGemHome := "/Users/test/.gem/"
	expectedGemHomeArray := []interface{}{expectedGemHome}
	expectedWarningMessage := "GEM_HOME environment variable is set to:\n%s\nThis can lead to errors as gem lookup path may not contain GEM_HOME."
	inputs := Inputs{GemHome: expectedGemHome}
	config := Config{Inputs: inputs}

	mockedLogger.On("Warnf", expectedWarningMessage, expectedGemHomeArray)

	step.validateGemHome(config)

	mockedLogger.AssertCalled(t, "Warnf", expectedWarningMessage, expectedGemHomeArray)
}

func Test_GivenLaneParams_WhenProcessConfig_ThenLaneOptionsIncludeParams(t *testing.T) {
	envRepo := env.NewRepository()
	envRepo.Set("lane", "deploy")
	envRepo.Set("lane_params", "track:beta")
	envRepo.Set("work_dir", ".")
	envRepo.Set("connection", "automatic")
	envRepo.Set("update_fastlane", "true")
	envRepo.Set("verbose_log", "no")
	envRepo.Set("enable_cache", "yes")

	inputParser := stepconf.NewInputParser(envRepo)
	logger := log.NewLogger()
	tracker := newStepTracker(envRepo, logger)
	step := NewFastlaneRunner(inputParser, logger, env.NewCommandLocator(), command.NewFactory(envRepo), nil, nil, pathutil.NewPathModifier(), tracker)

	config, err := step.ProcessConfig()

	assert.NoError(t, err)
	assert.Equal(t, []string{"deploy", "track:beta"}, config.LaneOptions)
}
