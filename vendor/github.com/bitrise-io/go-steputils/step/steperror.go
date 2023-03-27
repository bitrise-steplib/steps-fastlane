package step

import "errors"

// Error is an error occuring top level in a step
type Error struct {
	StepID, Tag, ShortMsg string
	Err                   error
	Recommendations       Recommendation
}

// Recommendation interface
type Recommendation map[string]interface{}

// NewError constructs a step.Error
func NewError(stepID, tag string, err error, shortMsg string) *Error {
	return &Error{
		StepID:   stepID,
		Tag:      tag,
		Err:      err,
		ShortMsg: shortMsg,
	}
}

// NewErrorWithRecommendations constructs a step.Error
func NewErrorWithRecommendations(stepID, tag string, err error, shortMsg string, recommendations Recommendation) *Error {
	return &Error{
		StepID:          stepID,
		Tag:             tag,
		Err:             err,
		ShortMsg:        shortMsg,
		Recommendations: recommendations,
	}
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	if err := errors.Unwrap(e.Err); err != nil {
		return err
	}

	return nil
}
