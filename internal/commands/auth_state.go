package commands

import (
	"errors"

	"github.com/rishabyd/codeberg-cli/internal/config"
)

func requireAuth() (*config.AuthConfig, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, mapConfigError(err)
	}
	if cfg == nil || cfg.AccessToken == "" {
		return nil, errors.New("Not logged in. Run `cb auth login`")
	}
	return cfg, nil
}
