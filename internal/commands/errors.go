package commands

import (
	"errors"

	"github.com/rishabyd/codeberg-cli/internal/codeberg"
	"github.com/rishabyd/codeberg-cli/internal/config"
)

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

func mapConfigError(err error) error {
	if err == nil {
		return nil
	}
	if config.IsCorruptConfigError(err) {
		return errors.New("Local configuration is corrupt. Run `cb auth logout` then `cb auth login`")
	}
	return err
}

func mapCodebergError(err error) error {
	if err == nil {
		return nil
	}
	if codeberg.IsNetworkError(err) {
		return errors.New("Could not reach Codeberg (network or DNS issue)")
	}
	if codeberg.IsAuthError(err) {
		return errors.New("Session expired. Run `cb auth login`")
	}
	return err
}
