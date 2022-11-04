package main

import (
	"testing"

	"github.com/bitrise-io/go-xcode/appleauth"
	"github.com/stretchr/testify/assert"
)

// func TestFastlaneStep_ProcessInputs(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		envs map[string]string
// 		want Config
// 		err  string
// 	}{
// 		{
// 			name: "project_path should be and .xcodeproj or .xcworkspace path",
// 			envs: override(thisStepInputs(t), map[string]string{
// 				"project_path": ".",
// 				"scheme":       "My Scheme",
// 				"workdir":      "",
// 			}),
// 			want: Config{},
// 			err:  "issue with input ProjectPath: should be and .xcodeproj or .xcworkspace path",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			envRepository := MockEnvRepository{envs: tt.envs}
// 			rbyFactory, err := ruby.NewCommandFactory(command.NewFactory(env.NewRepository()), cmdLocator)
// 			s := FastlaneRunner{
// 				stepInputParser: stepconf.NewInputParser(envRepository),
// 				logger: log.NewLogger(),
// 				cmdFactory: command.NewFactory(envRepository)
// 				cmdLocator: tt.envs.NewCommandLocator()
// 				rbyFactory: rbyFactory

// 			}

// 			config, err := s.ProcessInputs()
// 			gotErr := err != nil
// 			wantErr := tt.err != ""
// 			require.Equal(t, wantErr, gotErr, fmt.Sprintf("Step.ValidateConfig() error = %v, wantErr %v", err, tt.err))
// 			require.Equal(t, tt.want, config)
// 		})
// 	}
// }

// func thisStepInputs(t *testing.T) map[string]string {
// 	_, filename, _, _ := runtime.Caller(1)
// 	thisPackageDir := filepath.Dir(filename)
// 	rootDir := filepath.Dir(thisPackageDir)
// 	stepYMLPth := filepath.Join(rootDir, "step.yml")
// 	b, err := fileutil.ReadBytesFromFile(stepYMLPth)
// 	require.NoError(t, err)

// 	var s struct {
// 		Inputs []map[string]interface{} `yaml:"inputs"`
// 	}
// 	require.NoError(t, yaml.Unmarshal(b, &s))

// 	inputKeyValues := map[string]string{}
// 	for _, in := range s.Inputs {
// 		for k, v := range in {
// 			if k != "opts" {
// 				if v == nil {
// 					inputKeyValues[k] = ""
// 				} else {
// 					v, ok := v.(string)
// 					require.True(t, ok)
// 					inputKeyValues[k] = v

// 				}
// 				break
// 			}
// 		}
// 	}

// 	return inputKeyValues
// }

// func override(orig, new map[string]string) map[string]string {
// 	inputs := map[string]string{}
// 	for k, v := range orig {
// 		inputs[k] = v
// 	}

// 	for k, v := range new {
// 		inputs[k] = v
// 	}

// 	return inputs
// }

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

func Test_GivenAPIKeyConnection_WhenParseAuthSources_ThenReceiveAPIKeySource(t *testing.T) {
	step := FastlaneRunner{}
	expectedValue := []appleauth.Source{
		&appleauth.InputAPIKeySource{},
	}

	actualValue, err := step.parseAuthSources(apiKey)

	assert.NoError(t, err)
	assert.Equal(t, actualValue, expectedValue)
}

func Test_GivenAppleIDConnection_WhenParseAuthSources_ThenReceiveAppleIDFastlaneSource(t *testing.T) {
	step := FastlaneRunner{}
	expectedValue := []appleauth.Source{
		&appleauth.ConnectionAppleIDFastlaneSource{},
	}

	actualValue, err := step.parseAuthSources(appleID)

	assert.NoError(t, err)
	assert.Equal(t, actualValue, expectedValue)
}

func Test_GivenOffConnection_WhenParseAuthSources_ThenReceiveInputAPIKeyAndAppleIdFastlaneSources(t *testing.T) {
	step := FastlaneRunner{}
	expectedValue := []appleauth.Source{

		&appleauth.InputAPIKeySource{},
		&appleauth.InputAppleIDFastlaneSource{},
	}

	actualValue, err := step.parseAuthSources(off)

	assert.NoError(t, err)
	assert.Equal(t, actualValue, expectedValue)
}

func Test_GivenGemHomeEnvironmentVariableIsEmpty_WhenValidateGemHome_ThenLogNothing(t *testing.T) {
	mockLogger := MockLogger{}
	step := FastlaneRunner{logger: &mockLogger}
	expectedGemHome := ""
	config := Config{GemHome: expectedGemHome}

	step.validateGemHome(config)

	mockLogger.AssertNotCalled(t, "Warnf")
}

func Test_GivenGemHomeEnvironmentVariableIsEmpty_WhenValidateGemHome_ThenLogWarning(t *testing.T) {
	var mockedLogger MockLogger
	step := FastlaneRunner{logger: &mockedLogger}
	expectedGemHome := "/Users/test/.gem/"
	expectedGemHomeArray := []interface{}{expectedGemHome}
	expectedWarningMessage := "GEM_HOME environment variable is set to:\n%s\nThis can lead to errors as gem lookup path may not contain GEM_HOME."
	config := Config{GemHome: expectedGemHome}

	mockedLogger.On("Warnf", expectedWarningMessage, expectedGemHomeArray)

	step.validateGemHome(config)

	mockedLogger.AssertCalled(t, "Warnf", expectedWarningMessage, expectedGemHomeArray)
}
