package commands

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

type healthService struct {
	label    string
	badgeURL string
}

var healthServices = []healthService{
	{label: "Codeberg.org", badgeURL: "https://status.codeberg.org/api/badge/1/status"},
	{label: "Hosted CI/CD: Forgejo Actions", badgeURL: "https://status.codeberg.org/api/badge/38/status"},
	{label: "Hosted CI/CD: Woodpecker", badgeURL: "https://status.codeberg.org/api/badge/29/status"},
}

type healthResult struct {
	label  string
	status string
}

func runHealth() error {
	labelWidth := 0
	for _, svc := range healthServices {
		if len(svc.label) > labelWidth {
			labelWidth = len(svc.label)
		}
	}
	labelWidth += 2

	client := &http.Client{Timeout: 5 * time.Second}
	results := make([]healthResult, len(healthServices))

	var wg sync.WaitGroup
	for i, svc := range healthServices {
		wg.Add(1)
		go func(idx int, svc healthService) {
			defer wg.Done()
			results[idx] = healthResult{
				label:  svc.label,
				status: fetchServiceStatus(client, svc.badgeURL),
			}
		}(i, svc)
	}
	wg.Wait()

	allUnknown := true
	for _, r := range results {
		var color, status string
		switch r.status {
		case "Up":
			allUnknown = false
			color, status = colorGreen, "Up"
		case "Down":
			allUnknown = false
			color, status = colorRed, "Down"
		default:
			color, status = colorReset, "Unknown"
		}
		fmt.Printf("%-*s %s%s%s\n", labelWidth, r.label, color, status, colorReset)
	}

	fmt.Println()
	if allUnknown {
		fmt.Println("Could not reach status servers. Check your internet connection.")
	}
	fmt.Println("Details: https://status.codeberg.org/status/codeberg")
	return nil
}

func fetchServiceStatus(client *http.Client, url string) string {
	resp, err := client.Get(url)
	if err != nil {
		return "Unknown"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "Unknown"
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return "Unknown"
	}

	svg := string(body)
	if strings.Contains(svg, `aria-label="Status: Up"`) {
		return "Up"
	}
	if strings.Contains(svg, `aria-label="Status: Down"`) {
		return "Down"
	}

	return "Unknown"
}
