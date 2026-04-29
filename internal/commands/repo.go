package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/rodaine/table"
	"github.com/rishabyd/codeberg-cli/internal/codeberg"
	"github.com/rishabyd/codeberg-cli/internal/repository"
	"github.com/rishabyd/codeberg-cli/internal/validation"
	"github.com/spf13/cobra"
)

var repoSvc = &repository.Service{}

func newRepoCmd() *cobra.Command {
	repoCmd := &cobra.Command{
		Use:     "repo",
		Short:   "Manage repositories",
		Aliases: []string{"repository"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				_ = cmd.Help()
				return usageError("repo action required")
			}
			return unknownActionError(cmd, "repo action", args[0])
		},
	}
	repoCmd.SuggestionsMinimumDistance = 2

	var listLimit int
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List repositories",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRepoList(cmd.Context(), listLimit)
		},
	}
	listCmd.Flags().IntVar(&listLimit, "limit", 30, "Number of repositories")
	repoCmd.AddCommand(listCmd)

	var createDescription string
	var createPublic bool
	var createPrivate bool
	var createAddReadme bool
	var createClone bool
	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a repository",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
				return usageError("repository name is required: cb repo create <name> [flags]")
			}
			if err := validation.RepoName(strings.TrimSpace(args[0])); err != nil {
				return usageError(err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if createPublic && createPrivate {
				return usageError("--public and --private cannot be used together")
			}

			privateValue := true
			if createPublic {
				privateValue = false
			}
			if createPrivate {
				privateValue = true
			}

			return runRepoCreate(cmd.Context(), repository.CreateInput{
				Name:        strings.TrimSpace(args[0]),
				Description: createDescription,
				Private:     privateValue,
				AddReadme:   createAddReadme,
				CloneAfter:  createClone,
			})
		},
	}
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "Repository description")
	createCmd.Flags().BoolVar(&createPublic, "public", false, "Create as public")
	createCmd.Flags().BoolVar(&createPrivate, "private", false, "Create as private")
	createCmd.Flags().BoolVar(&createAddReadme, "add-readme", false, "Initialize with README")
	createCmd.Flags().BoolVarP(&createClone, "clone", "c", false, "Clone after creating")
	repoCmd.AddCommand(createCmd)

	var migrateClone bool
	migrateCmd := &cobra.Command{
		Use:   "migrate <owner/repo>",
		Short: "Migrate public GitHub repository to Codeberg",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
				return usageError("repository is required: cb repo migrate <owner/repo> [--clone]")
			}
			if err := validation.OwnerRepo(strings.TrimSpace(args[0])); err != nil {
				return usageError("invalid repository format, expected owner/repo")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepoMigrate(cmd.Context(), strings.TrimSpace(args[0]), migrateClone)
		},
	}
	migrateCmd.Flags().BoolVar(&migrateClone, "clone", false, "Clone after migrating")
	repoCmd.AddCommand(migrateCmd)

	return repoCmd
}

func runRepoCreate(ctx context.Context, input repository.CreateInput) error {
	cfg, err := requireAuth()
	if err != nil {
		return err
	}

	out, err := repoSvc.Create(ctx, cfg, input)
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

func runRepoList(ctx context.Context, limit int) error {
	cfg, err := requireAuth()
	if err != nil {
		return err
	}

	result, err := repoSvc.List(ctx, cfg, limit)
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

	tbl := table.New("Name", "Description", "Visibility", "Updated")
	for _, item := range result.Items {
		tbl.AddRow(item.Name, item.Desc, item.Visibility, item.UpdatedAgo)
	}
	tbl.Print()

	return nil
}

func runRepoMigrate(ctx context.Context, source string, cloneAfter bool) error {
	cfg, err := requireAuth()
	if err != nil {
		return err
	}

	out, err := repoSvc.Migrate(ctx, cfg, repository.MigrateInput{
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

func printRepoResult(action string, repo *codeberg.Repo) {
	if repo == nil {
		return
	}
	fmt.Printf("%s %s\n", action, repo.FullName)
	fmt.Printf("HTTPS: %s\n", repo.CloneURL)
	fmt.Printf("SSH:   %s\n", repo.SSHURL)
	fmt.Printf("Web:   %s\n", repo.HTMLURL)
}

func printClonePath(name string) {
	if name == "" {
		return
	}
	fmt.Printf("Cloned to ./%s\n", name)
}
