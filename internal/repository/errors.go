package repository

import "errors"

type UsageError struct {
	Message string
}

func (e *UsageError) Error() string {
	if e == nil {
		return "Invalid usage"
	}
	return e.Message
}

func usageError(message string) error {
	return &UsageError{Message: message}
}

func IsUsageError(err error) bool {
	var target *UsageError
	return errors.As(err, &target)
}
