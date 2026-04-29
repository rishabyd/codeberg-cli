package commands

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rishabyd/codeberg-cli/internal/codeberg"
	"github.com/rishabyd/codeberg-cli/internal/config"
)

func TestIsUsageError(t *testing.T) {
	if !IsUsageError(usageError("bad input")) {
		t.Error("usageError should be detected as UsageError")
	}
	if IsUsageError(fmt.Errorf("other error")) {
		t.Error("regular error should not be detected as UsageError")
	}
	if IsUsageError(nil) {
		t.Error("nil should not be UsageError")
	}
}

func TestMapCodebergError_Network(t *testing.T) {
	err := mapCodebergError(&codeberg.NetworkError{Err: errors.New("timeout")})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "Could not reach Codeberg (network or DNS issue)" {
		t.Errorf("unexpected message: %v", err)
	}
}

func TestMapCodebergError_Auth(t *testing.T) {
	err := mapCodebergError(&codeberg.AuthError{Message: "bad token"})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "Session expired. Run `cb auth login`" {
		t.Errorf("unexpected message: %v", err)
	}
}

func TestMapCodebergError_Nil(t *testing.T) {
	if err := mapCodebergError(nil); err != nil {
		t.Errorf("nil should return nil: %v", err)
	}
}

func TestMapConfigError_Corrupt(t *testing.T) {
	err := mapConfigError(&config.CorruptConfigError{Path: "/tmp/fake", Err: errors.New("parse error")})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "Local configuration is corrupt. Run `cb auth logout` then `cb auth login`" {
		t.Errorf("unexpected message: %v", err)
	}
}

func TestMapConfigError_Nil(t *testing.T) {
	if err := mapConfigError(nil); err != nil {
		t.Errorf("nil should return nil: %v", err)
	}
}

func TestUsageErrorMessage(t *testing.T) {
	u := usageError("something wrong")
	if u.Error() != "something wrong" {
		t.Errorf("unexpected message: %v", u)
	}
	var nilPtr *UsageError
	if nilPtr.Error() != "Invalid usage" {
		t.Error("nil UsageError should return default message")
	}
}
