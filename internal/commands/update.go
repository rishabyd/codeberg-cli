package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const latestReleaseURL = "https://api.github.com/repos/rishabyd/codeberg-cli/releases/latest"

type ghRelease struct {
	TagName string `json:"tag_name"`
}

func checkAndUpdate(currentVersion string) error {
	latest, err := fetchLatestTag()
	if err != nil {
		fmt.Println("Could not check for updates. Proceeding...")
	} else if !isNewer(latest, currentVersion) {
		fmt.Printf("Already up to date (v%s)\n", currentVersion)
		return nil
	} else {
		fmt.Printf("A new version is available: %s (current: v%s)\n", latest, currentVersion)
	}

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

	return execUpdate()
}

func fetchLatestTag() (string, error) {
	resp, err := http.Get(latestReleaseURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

func isNewer(latest, current string) bool {
	l := parseVersion(latest)
	c := parseVersion(current)
	if l == nil || c == nil {
		return true
	}
	for i := 0; i < 3; i++ {
		if l[i] > c[i] {
			return true
		}
		if l[i] < c[i] {
			return false
		}
	}
	return false
}

func parseVersion(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) < 3 {
		return nil
	}
	nums := make([]int, 3)
	for i := 0; i < 3; i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}

func execUpdate() error {
	cmd := exec.Command("bash", "-c",
		"curl -fsSL https://raw.githubusercontent.com/rishabyd/codeberg-cli/main/install.sh | bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
