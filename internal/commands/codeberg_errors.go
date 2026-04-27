package commands

import (
	"errors"

	"github.com/rishabyd/codeberg-cli/internal/codeberg"
)

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
