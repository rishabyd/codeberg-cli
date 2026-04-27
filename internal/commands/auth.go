package commands

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rishabyd/codeberg-cli/internal/codeberg"
	"github.com/rishabyd/codeberg-cli/internal/config"
	"github.com/rishabyd/codeberg-cli/internal/constants"
	"github.com/rishabyd/codeberg-cli/internal/gitcred"
	"github.com/rishabyd/codeberg-cli/internal/oauth"
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

	verifier, challenge, err := oauth.GeneratePKCE()
	if err != nil {
		return err
	}
	authURL := oauth.AuthorizationURL(challenge)

	fmt.Println("Opening browser for Codeberg authentication...")
	fmt.Println("If browser does not open, use this URL:")
	fmt.Println(authURL)

	_ = openBrowser(authURL)

	loginCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	code, err := oauth.WaitForAuthCode(loginCtx)
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
