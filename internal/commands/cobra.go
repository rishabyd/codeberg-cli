package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rishabyd/codeberg-cli/internal/logging"
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
	var verbose bool

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
	root.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")

	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if verbose {
			logging.SetVerbose(true)
		}
	}

	root.AddCommand(newAuthCmd())
	root.AddCommand(newRepoCmd())
	root.AddCommand(newUpdateCmd(version))
	root.AddCommand(newHealthCmd())

	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return usageError(err.Error())
	})

	return root
}

func newUpdateCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update to the latest release",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return checkAndUpdate(version)
		},
	}
}

func newHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check Codeberg service status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runHealth()
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
