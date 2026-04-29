package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	selfupdate "github.com/creativeprojects/go-selfupdate"
	"github.com/rishabyd/codeberg-cli/internal/constants"
)

func checkAndUpdate(currentVersion string) error {
	ctx := context.Background()
	repo := selfupdate.ParseSlug(constants.UpdateRepository)

	latest, found, err := selfupdate.DetectLatest(ctx, repo)
	if err != nil {
		fmt.Println("Could not check for updates.")
		return nil
	}
	if !found {
		fmt.Println("No releases found.")
		return nil
	}

	if !latest.GreaterThan(currentVersion) {
		fmt.Printf("Already up to date (v%s)\n", currentVersion)
		return nil
	}

	fmt.Printf("A new version is available: v%s (current: v%s)\n", latest.Version(), currentVersion)

	fmt.Print("Proceed with update? [Y/n] ")
	var answer string
	if _, err := fmt.Scanln(&answer); err != nil {
		return nil
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "" && answer != "y" && answer != "yes" {
		fmt.Println("To update manually, run:")
		fmt.Println("  curl -fsSL https://raw.githubusercontent.com/rishabyd/codeberg-cli/main/install.sh | bash")
		return nil
	}

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("could not locate executable path: %w", err)
	}

	if err := selfupdate.UpdateTo(ctx, latest.AssetURL, latest.AssetName, exe); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Updated to v%s. Please restart your shell.\n", latest.Version())
	return nil
}
