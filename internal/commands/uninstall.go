package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rishabyd/codeberg-cli/internal/config"
	"github.com/rishabyd/codeberg-cli/internal/constants"
	"github.com/rishabyd/codeberg-cli/internal/gitcred"
	"github.com/rishabyd/codeberg-cli/internal/output"
	"github.com/spf13/cobra"
)

func runUninstall(ctx context.Context) error {
	fmt.Print("Uninstall cb? [y/N] ")
	var answer string
	if _, err := fmt.Scanln(&answer); err != nil {
		return nil
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println(output.Dim("Aborted."))
		return nil
	}

	target := filepath.Join(constants.ReleaseInstallDir, constants.ReleaseBinaryName)
	if err := removeBinary(target); err != nil {
		fmt.Println(err)
	}

	gitcred.RemoveGlobalHelper(ctx)

	if err := config.Clear(); err != nil {
		fmt.Printf("Could not remove config: %v\n", err)
	}

	fmt.Println(output.Success("Uninstalled cb and removed local config from ~/.config/" + constants.ConfigDirName))
	return nil
}

func removeBinary(target string) error {
	if err := os.Remove(target); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		if os.IsPermission(err) {
			fmt.Printf("Removing %s (sudo may prompt)...\n", target)
			cmd := exec.Command("sudo", "rm", "-f", target)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
		return fmt.Errorf("could not remove %s: %v", target, err)
	}
	return nil
}

func newUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the cb binary and local config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runUninstall(cmd.Context())
		},
	}
}
