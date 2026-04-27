package gitcred

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/rishabyd/codeberg-cli/internal/codeberg"
	"github.com/rishabyd/codeberg-cli/internal/config"
	"github.com/rishabyd/codeberg-cli/internal/constants"
)

func ConfigureGlobalHelper(ctx context.Context) error {
	path, err := os.Executable()
	if err != nil || path == "" {
		path, err = exec.LookPath("cb")
		if err != nil {
			path = "cb"
		}
	}
	helper := "!" + path + " auth git-credential"
	cmd := exec.CommandContext(ctx, "git", "config", "--global", constants.GitCredentialTarget, helper)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to configure git credential helper: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func RemoveGlobalHelper(ctx context.Context) {
	_ = exec.CommandContext(ctx, "git", "config", "--global", "--unset", constants.GitCredentialTarget).Run()
}

type request struct {
	Protocol string
	Host     string
}

func readRequest(r io.Reader) (request, error) {
	var req request
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			break
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "protocol":
			req.Protocol = parts[1]
		case "host":
			req.Host = parts[1]
		}
	}
	return req, s.Err()
}

func HandleCredentialOperation(ctx context.Context, op string, stdin io.Reader, stdout io.Writer) error {
	if op != "get" {
		return nil
	}

	req, err := readRequest(stdin)
	if err != nil {
		return err
	}
	if req.Protocol != "https" || req.Host != constants.CodebergHost {
		return nil
	}

	cfg, err := config.Load()
	if err != nil || cfg == nil || cfg.AccessToken == "" {
		return nil
	}

	user, err := codeberg.GetCurrentUser(ctx, cfg)
	if err != nil {
		if codeberg.IsNetworkError(err) {
			return nil
		}
		if codeberg.IsAuthError(err) {
			return nil
		}
		return nil
	}

	username := cfg.Username
	if username == "" && user != nil {
		username = user.Login
	}
	token := cfg.AccessToken

	if username == "" {
		username = "oauth2"
	}

	fmt.Fprintf(stdout, "username=%s\n", username)
	fmt.Fprintf(stdout, "password=%s\n\n", token)
	return nil
}
