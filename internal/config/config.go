package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rishabyd/codeberg-cli/internal/constants"
)

type AuthConfig struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	Username     string    `json:"username,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
}

type CorruptConfigError struct {
	Path string
	Err  error
}

func (e *CorruptConfigError) Error() string {
	if e == nil {
		return "config file is invalid"
	}
	if e.Path == "" {
		return fmt.Sprintf("config file is invalid: %v", e.Err)
	}
	return fmt.Sprintf("config file is invalid at %s: %v", e.Path, e.Err)
}

func (e *CorruptConfigError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func IsCorruptConfigError(err error) bool {
	var target *CorruptConfigError
	return errors.As(err, &target)
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", constants.ConfigDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, constants.ConfigFileName), nil
}

func Load() (*AuthConfig, error) {
	p, err := configPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var c AuthConfig
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, &CorruptConfigError{Path: p, Err: err}
	}
	if c.AccessToken == "" {
		return nil, nil
	}
	return &c, nil
}

func Save(c AuthConfig) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(p, b, 0o600); err != nil {
		return err
	}
	return os.Chmod(p, 0o600)
}

func Clear() error {
	p, err := configPath()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
