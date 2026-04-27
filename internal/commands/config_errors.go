package commands

import (
	"errors"

	"github.com/rishabyd/codeberg-cli/internal/config"
)

func mapConfigError(err error) error {
	if err == nil {
		return nil
	}
	if config.IsCorruptConfigError(err) {
		return errors.New("Local configuration is corrupt. Run `cb auth logout` then `cb auth login`")
	}
	return err
}
