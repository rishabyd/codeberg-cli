package commands

import (
	"context"
	"fmt"

	"github.com/rishabyd/codeberg-cli/internal/repository"
)

func runRepoList(ctx context.Context, limit int) error {
	cfg, err := requireAuth()
	if err != nil {
		return err
	}

	result, err := repoService().List(ctx, cfg, limit)
	if err != nil {
		if repository.IsUsageError(err) {
			return usageError(err.Error())
		}
		return mapCodebergError(err)
	}
	if result == nil || len(result.Items) == 0 {
		fmt.Println("No repositories.")
		return nil
	}

	for _, item := range result.Items {
		fmt.Printf("%-*s  %-40s  %-7s  %s\n", result.NameWidth+2, item.Name, item.Desc, item.Visibility, item.UpdatedAgo)
	}

	return nil
}
