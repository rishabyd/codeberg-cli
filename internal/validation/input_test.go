package validation

import "testing"

func TestRepoName_Valid(t *testing.T) {
	valid := []string{"a", "my-repo", "repo.name", "repo_name", "Repo123", "a1b2c3", "a._-b"}
	for _, name := range valid {
		if err := RepoName(name); err != nil {
			t.Errorf("Reponame(%q) should be valid: %v", name, err)
		}
	}
}

func TestRepoName_Invalid(t *testing.T) {
	invalid := []string{
		"",
		"   ",
		"-startsWithDash",
		".startsWithDot",
		"_startsWithUnderscore",
	}
	for _, name := range invalid {
		if err := RepoName(name); err == nil {
			t.Errorf("Reponame(%q) should be invalid", name)
		}
	}
}

func TestRepoName_ControlChars(t *testing.T) {
	if err := RepoName("bad\x00name"); err == nil {
		t.Error("Reponame with null byte should be invalid")
	}
}

func TestOwnerRepo_Valid(t *testing.T) {
	valid := []string{"owner/repo", "user123/my-repo", "o/r"}
	for _, v := range valid {
		if err := OwnerRepo(v); err != nil {
			t.Errorf("OwnerRepo(%q) should be valid: %v", v, err)
		}
	}
}

func TestOwnerRepo_Invalid(t *testing.T) {
	invalid := []string{"", "owner", "owner/", "/repo", "owner/repo/extra"}
	for _, v := range invalid {
		if err := OwnerRepo(v); err == nil {
			t.Errorf("OwnerRepo(%q) should be invalid", v)
		}
	}
}

func TestDescription_Valid(t *testing.T) {
	if err := Description("A valid description"); err != nil {
		t.Errorf("Description should be valid: %v", err)
	}
	if err := Description(""); err != nil {
		t.Errorf("Empty description should be valid: %v", err)
	}
}

func TestPositiveLimit_Valid(t *testing.T) {
	if err := PositiveLimit(1, 100); err != nil {
		t.Errorf("PositiveLimit(1, 100) should be valid: %v", err)
	}
	if err := PositiveLimit(100, 100); err != nil {
		t.Errorf("PositiveLimit(100, 100) should be valid: %v", err)
	}
}

func TestPositiveLimit_Invalid(t *testing.T) {
	if err := PositiveLimit(0, 100); err == nil {
		t.Error("PositiveLimit(0) should be invalid")
	}
	if err := PositiveLimit(101, 100); err == nil {
		t.Error("PositiveLimit(101, 100) should be invalid")
	}
}

func TestNonEmptyTrimmed(t *testing.T) {
	if err := NonEmptyTrimmed("test", "value"); err != nil {
		t.Errorf("NonEmptyTrimmed should be valid: %v", err)
	}
	if err := NonEmptyTrimmed("test", "  "); err == nil {
		t.Error("NonEmptyTrimmed with whitespace should be invalid")
	}
}
