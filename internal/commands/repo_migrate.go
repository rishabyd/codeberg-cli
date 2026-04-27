package commands

import (
	"context"
	"fmt"

	"github.com/rishabyd/codeberg-cli/internal/repository"
)

func runRepoMigrate(ctx context.Context, source string, cloneAfter bool) error {
	cfg, err := requireAuth()
	if err != nil {
		return err
	}

	out, err := repoService().Migrate(ctx, cfg, repository.MigrateInput{
		Source:     source,
		CloneAfter: cloneAfter,
	})
	if err != nil {
		if repository.IsUsageError(err) {
			return usageError(err.Error())
		}
		return mapCodebergError(err)
	}
	if out == nil || out.Repo == nil {
		return nil
	}

	fmt.Printf("Migrated %s to Codeberg\n", source)
	fmt.Printf("HTTPS: %s\n", out.Repo.CloneURL)
	fmt.Printf("SSH:   %s\n", out.Repo.SSHURL)
	fmt.Printf("Web:   %s\n", out.Repo.HTMLURL)
	printClonePath(out.ClonedTo)

	return nil
}
