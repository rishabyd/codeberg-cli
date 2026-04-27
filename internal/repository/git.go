package repository

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Git interface {
	Clone(ctx context.Context, cloneURL, directory string) error
}

type ExecGit struct{}

func (ExecGit) Clone(ctx context.Context, cloneURL, directory string) error {
	args := []string{"clone", "--", cloneURL}
	if strings.TrimSpace(directory) != "" {
		args = append(args, directory)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
	}
	return nil
}
