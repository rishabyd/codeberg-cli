package commands

import (
	"context"

	"github.com/rishabyd/codeberg-cli/internal/repository"
)

func runRepoCreate(ctx context.Context, input repository.CreateInput) error {
	cfg, err := requireAuth()
	if err != nil {
		return err
	}

	out, err := repoService().Create(ctx, cfg, input)
	if err != nil {
		if repository.IsUsageError(err) {
			return usageError(err.Error())
		}
		return mapCodebergError(err)
	}

	if out == nil || out.Repo == nil {
		return nil
	}
	printRepoResult("Created", out.Repo)
	printClonePath(out.ClonedTo)

	return nil
}
