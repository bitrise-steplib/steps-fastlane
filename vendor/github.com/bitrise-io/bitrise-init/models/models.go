package models

import (
	"github.com/bitrise-io/go-steputils/step"
)

type BitriseConfigMap map[string]string

type Warnings []string

type Errors []string

// Icon is potential app icon.
// The name is unique (sha256 hash of relative path converted to string plus the original extension appended).
type Icon struct {
	Filename string
	Path     string
}

type Icons []Icon

type ErrorWithRecommendations struct {
	Error           string
	Recommendations step.Recommendation
}

type ErrorsWithRecommendations []ErrorWithRecommendations

type ScanResultModel struct {
	ScannerToOptionRoot                  map[string]OptionNode                `json:"options,omitempty" yaml:"options,omitempty"`
	ScannerToBitriseConfigMap            map[string]BitriseConfigMap          `json:"configs,omitempty" yaml:"configs,omitempty"`
	ScannerToWarnings                    map[string]Warnings                  `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	ScannerToErrors                      map[string]Errors                    `json:"errors,omitempty" yaml:"errors,omitempty"`
	ScannerToErrorsWithRecommendations   map[string]ErrorsWithRecommendations `json:"errors_with_recommendations,omitempty" yaml:"errors_with_recommendations,omitempty"`
	ScannerToWarningsWithRecommendations map[string]ErrorsWithRecommendations `json:"warnings_with_recommendations,omitempty" yaml:"warnings_with_recommendations,omitempty"`
	Icons                                []Icon                               `json:"-" yaml:"-"`
}

type SSHKeyActivation int

const (
	SSHKeyActivationNone = iota
	SSHKeyActivationMandatory
	SSHKeyActivationConditional
)

func (result *ScanResultModel) AddErrorWithRecommendation(platform string, recommendation ErrorWithRecommendations) {
	if result.ScannerToErrorsWithRecommendations == nil {
		result.ScannerToErrorsWithRecommendations = map[string]ErrorsWithRecommendations{}
	}
	if result.ScannerToErrorsWithRecommendations[platform] == nil {
		result.ScannerToErrorsWithRecommendations[platform] = ErrorsWithRecommendations{}
	}
	result.ScannerToErrorsWithRecommendations[platform] = append(result.ScannerToErrorsWithRecommendations[platform], recommendation)
}
