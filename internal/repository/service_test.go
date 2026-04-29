package repository

import "testing"

func TestTimeAgo_Seconds(t *testing.T) {
	result := timeAgo("2024-01-01T00:00:30Z")
	if result == "-" {
		t.Error("timeAgo should not return '-' for valid time")
	}
}

func TestTimeAgo_Empty(t *testing.T) {
	result := timeAgo("")
	if result != "-" {
		t.Errorf("timeAgo for empty string should return '-', got %q", result)
	}
}

func TestTimeAgo_InvalidFormat(t *testing.T) {
	result := timeAgo("not-a-date")
	if result != "-" {
		t.Errorf("timeAgo for invalid format should return '-', got %q", result)
	}
}

func TestUsageError_IsUsageError(t *testing.T) {
	err := usageError("bad input")
	if !IsUsageError(err) {
		t.Error("IsUsageError should return true for usageError")
	}
	if IsUsageError(nil) {
		t.Error("IsUsageError should return false for nil")
	}
}
