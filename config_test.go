package main

import (
	"testing"

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
