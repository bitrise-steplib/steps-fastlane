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
