package commands

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rishabyd/codeberg-cli/internal/constants"
	"github.com/rishabyd/codeberg-cli/internal/output"
	"github.com/rodaine/table"
)

type healthService struct {
	label    string
	badgeURL string
}

var healthServices []healthService

func init() {
	healthServices = []healthService{
		{label: "Codeberg.org", badgeURL: fmt.Sprintf("https://status.codeberg.org/api/badge/%d/status", constants.StatusBadgeIDs[0])},
		{label: "Hosted CI/CD: Forgejo Actions", badgeURL: fmt.Sprintf("https://status.codeberg.org/api/badge/%d/status", constants.StatusBadgeIDs[1])},
		{label: "Hosted CI/CD: Woodpecker", badgeURL: fmt.Sprintf("https://status.codeberg.org/api/badge/%d/status", constants.StatusBadgeIDs[2])},
	}
}

type healthResult struct {
	label  string
	status string
}

func runHealth() error {
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

	tbl := table.New("Service", "Status")
	allUnknown := true
	for _, r := range results {
		var styled string
		switch r.status {
		case "Up":
			styled = output.StatusUp()
			allUnknown = false
		case "Down":
			styled = output.StatusDown()
			allUnknown = false
		default:
			styled = output.StatusUnknown()
		}
		tbl.AddRow(r.label, styled)
	}
	tbl.Print()

	fmt.Println()
	if allUnknown {
		fmt.Println(output.Warning("Could not reach status servers. Check your internet connection."))
	}
	fmt.Println(output.Dim("Details: https://status.codeberg.org/status/codeberg"))
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
