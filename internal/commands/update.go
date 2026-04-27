package commands

import (
	"os"
	"os/exec"
)

func runUpdate() error {
	cmd := exec.Command("bash", "-c",
		"curl -fsSL https://raw.githubusercontent.com/rishabyd/codeberg-cli/main/install.sh | bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
