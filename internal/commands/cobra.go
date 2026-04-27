package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rishabyd/codeberg-cli/internal/gitcred"
	"github.com/rishabyd/codeberg-cli/internal/repository"
	"github.com/rishabyd/codeberg-cli/internal/validation"
	"github.com/spf13/cobra"
)

func Execute(version string, args []string) error {
	root := newRootCmd(version)
	root.SetArgs(args)
	if err := root.ExecuteContext(context.Background()); err != nil {
		if IsUsageError(err) {
			return usageError(err.Error())
		}
		if isCobraUsageError(err) {
			return usageError(err.Error())
		}
		return err
	}
	return nil
}

func newRootCmd(version string) *cobra.Command {
	var showVersion bool

	root := &cobra.Command{
		Use:           "cb",
		Short:         "Codeberg CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				fmt.Fprintf(cmd.OutOrStdout(), "v%s\n", version)
				return nil
			}
			if len(args) == 0 {
				return cmd.Help()
			}
			return usageError(fmt.Sprintf("unknown command %q", args[0]))
		},
	}

	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.SuggestionsMinimumDistance = 2
	root.CompletionOptions.DisableDefaultCmd = true
	root.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "Show version")

	root.AddCommand(newAuthCmd())
	root.AddCommand(newRepoCmd())
	root.AddCommand(newUpdateCmd())

	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return usageError(err.Error())
	})

	return root
}

func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Codeberg",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				_ = cmd.Help()
				return usageError("auth action required")
			}
			return unknownActionError(cmd, "auth action", args[0])
		},
	}
	authCmd.SuggestionsMinimumDistance = 2

	authCmd.AddCommand(&cobra.Command{
		Use:   "login",
		Short: "Log in via OAuth (opens browser)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return authLogin(cmd.Context())
		},
	})

	authCmd.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Clear the local session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return authLogout(cmd.Context())
		},
	})

	authCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return authStatus(cmd.Context())
		},
	})

	gitCredCmd := &cobra.Command{
		Use:    "git-credential [operation]",
		Hidden: true,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 1 {
				return usageError("auth git-credential accepts at most one argument")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			op := "get"
			if len(args) == 1 {
				op = args[0]
			}
			return gitcred.HandleCredentialOperation(cmd.Context(), op, os.Stdin, os.Stdout)
		},
	}
	authCmd.AddCommand(gitCredCmd)

	return authCmd
}

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

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update to the latest release",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			return runUpdate()
		},
	}
}

func isCobraUsageError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.TrimSpace(strings.ToLower(err.Error()))
	if strings.HasPrefix(msg, "unknown command") {
		return true
	}
	if strings.HasPrefix(msg, "unknown flag") || strings.HasPrefix(msg, "unknown shorthand flag") {
		return true
	}
	if strings.Contains(msg, "accepts") && strings.Contains(msg, "arg") {
		return true
	}
	if strings.Contains(msg, "required flag") {
		return true
	}
	return false
}

func unknownActionError(cmd *cobra.Command, kind string, typed string) error {
	msg := fmt.Sprintf("unknown %s %q", kind, typed)
	suggestions := cmd.SuggestionsFor(typed)
	if len(suggestions) > 0 {
		msg += fmt.Sprintf("\n\nDid you mean:\n  %s", suggestions[0])
	}
	return usageError(msg)
}
