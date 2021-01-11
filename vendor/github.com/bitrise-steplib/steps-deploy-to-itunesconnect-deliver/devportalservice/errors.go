package devportalservice

import "fmt"

// NetworkError represents a networking issue.
type NetworkError struct {
	Status int
	Body   string
}

func (e NetworkError) Error() string {
	return fmt.Sprintf("response %d %s", e.Status, e.Body)
}

// CIEnvMissingError represents an issue caused by missing environment variables,
// which environment variables are exposed in builds on Bitrise.io.
type CIEnvMissingError struct {
	Key string
}

func (e CIEnvMissingError) Error() string {
	return fmt.Sprintf("%s env is not exported", e.Key)
}
