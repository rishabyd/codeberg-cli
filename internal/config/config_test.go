package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_MissingFile(t *testing.T) {
	// Override config dir to a temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should not error on missing file: %v", err)
	}
	if cfg != nil {
		t.Error("Load should return nil for missing file")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := AuthConfig{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		Username:     "testuser",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load returned nil for saved config")
	}
	if loaded.AccessToken != cfg.AccessToken {
		t.Errorf("AccessToken: got %q, want %q", loaded.AccessToken, cfg.AccessToken)
	}
	if loaded.RefreshToken != cfg.RefreshToken {
		t.Errorf("RefreshToken: got %q, want %q", loaded.RefreshToken, cfg.RefreshToken)
	}
	if loaded.Username != cfg.Username {
		t.Errorf("Username: got %q, want %q", loaded.Username, cfg.Username)
	}
}

func TestLoad_CorruptFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Manually write invalid JSON
	configDir := filepath.Join(tmpDir, ".config", "cb")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	configFile := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configFile, []byte("{invalid json"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("Load should error on corrupt file")
	}
	if !IsCorruptConfigError(err) {
		t.Errorf("Expected CorruptConfigError, got %T: %v", err, err)
	}
}

func TestLoad_EmptyToken(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := AuthConfig{RefreshToken: "refresh"}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded != nil {
		t.Error("Load should return nil when AccessToken is empty")
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := AuthConfig{AccessToken: "token"}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	if err := Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded != nil {
		t.Error("Load should return nil after Clear")
	}
}

func TestClear_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Clearing with no config file should not error
	if err := Clear(); err != nil {
		t.Errorf("Clear should not error on missing file: %v", err)
	}
}

func TestExpiryRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := AuthConfig{
		AccessToken: "test-token",
		Expiry:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(tmpDir, ".config", "cb", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["expiry"]; !ok {
		t.Error("expiry field missing from saved config")
	}
}
