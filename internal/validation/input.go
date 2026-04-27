package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	repoNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)
	ownerPattern    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)
)

func RepoName(name string) error {
	if hasControlChars(name) {
		return errors.New("Repository name contains invalid control characters")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("Repository name is required")
	}
	if !repoNamePattern.MatchString(name) {
		return errors.New("Repository name must match [A-Za-z0-9][A-Za-z0-9._-]{0,99}")
	}
	return nil
}

func OwnerRepo(value string) error {
	if hasControlChars(value) {
		return errors.New("Repository contains invalid control characters")
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("Repository must be in owner/repo format")
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return errors.New("Repository must be in owner/repo format")
	}
	if !ownerPattern.MatchString(parts[0]) || !repoNamePattern.MatchString(parts[1]) {
		return errors.New("Repository must be in owner/repo format")
	}
	return nil
}

func NonEmptyTrimmed(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	if hasControlChars(value) {
		return fmt.Errorf("%s contains invalid control characters", name)
	}
	return nil
}

func Description(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	if hasControlChars(value) {
		return errors.New("Description contains invalid control characters")
	}
	if len(value) > 2048 {
		return errors.New("Description is too long (max 2048 characters)")
	}
	return nil
}

func PositiveLimit(limit int, max int) error {
	if limit <= 0 {
		return errors.New("Limit must be greater than 0")
	}
	if max > 0 && limit > max {
		return fmt.Errorf("Limit must be <= %d", max)
	}
	return nil
}

func hasControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}
