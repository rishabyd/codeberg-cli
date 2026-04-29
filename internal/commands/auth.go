package commands

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rishabyd/codeberg-cli/internal/codeberg"
	"github.com/rishabyd/codeberg-cli/internal/config"
	"github.com/rishabyd/codeberg-cli/internal/constants"
	"github.com/rishabyd/codeberg-cli/internal/gitcred"
	"github.com/rishabyd/codeberg-cli/internal/oauth"
	"github.com/spf13/cobra"
)

func authLogin(ctx context.Context) error {
	existing, err := config.Load()
	if err != nil {
		return mapConfigError(err)
	}
	if existing != nil && existing.AccessToken != "" {
		fmt.Println("Already authenticated. Run `cb auth logout` to re-authenticate.")
		return nil
	}

	verifier := codeberg.GenerateVerifier()
	state := generateState()
	authURL := codeberg.AuthorizationURL(state, verifier)

	fmt.Println("Opening browser for Codeberg authentication...")
	fmt.Println("If browser does not open, use this URL:")
	fmt.Println(authURL)

	_ = openBrowser(authURL)

	loginCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	code, err := oauth.WaitForAuthCode(loginCtx, state)
	if err != nil {
		return err
	}

	tok, err := codeberg.ExchangeCode(ctx, code, verifier)
	if err != nil {
		return err
	}

	user, _, err := codeberg.GetCurrentUserByToken(ctx, tok.AccessToken)
	if err != nil {
		return err
	}

	if err := config.Save(config.AuthConfig{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		Username:     user.Login,
		Expiry:       time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second),
	}); err != nil {
		return err
	}

	if err := gitcred.ConfigureGlobalHelper(ctx); err != nil {
		return err
	}

	fmt.Printf("\n✓ Logged in as %s\n", user.Login)
	fmt.Println("Git credential helper configured.")
	return nil
}

func authLogout(ctx context.Context) error {
	gitcred.RemoveGlobalHelper(ctx)
	if err := config.Clear(); err != nil {
		return err
	}
	fmt.Println("✓ Logged out")
	return nil
}

func authStatus(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return mapConfigError(err)
	}
	if cfg == nil || cfg.AccessToken == "" {
		return errors.New("Not logged in. Run `cb auth login`")
	}

	user, err := codeberg.GetCurrentUser(ctx, cfg)
	if err != nil {
		return mapCodebergError(err)
	}
	if user == nil || strings.TrimSpace(user.Login) == "" {
		return errors.New("Failed to verify authenticated user")
	}

	fmt.Printf("%s\n", constants.CodebergHost)
	fmt.Printf("  ✓ Logged in as %s\n", user.Login)
	return nil
}

func openBrowser(url string) error {
	cmd := exec.Command("xdg-open", url)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to open browser: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func generateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

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

func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Codeberg",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				_ = cmd.Help()
				return usageError("auth action required")
			}
			return unknownActionError(cmd, "auth action", args[0])
		},
	}
	authCmd.SuggestionsMinimumDistance = 2

	authCmd.AddCommand(&cobra.Command{
		Use:   "login",
		Short: "Log in via OAuth (opens browser)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return authLogin(cmd.Context())
		},
	})

	authCmd.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Clear the local session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return authLogout(cmd.Context())
		},
	})

	authCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return authStatus(cmd.Context())
		},
	})

	gitCredCmd := &cobra.Command{
		Use:    "git-credential [operation]",
		Hidden: true,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 1 {
				return usageError("auth git-credential accepts at most one argument")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			op := "get"
			if len(args) == 1 {
				op = args[0]
			}
			return gitcred.HandleCredentialOperation(cmd.Context(), op, os.Stdin, os.Stdout)
		},
	}
	authCmd.AddCommand(gitCredCmd)

	return authCmd
}
