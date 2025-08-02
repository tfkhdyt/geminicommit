/*
Copyright © 2025 Christina Sørensen <ces@fem.gg>
*/
package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tfkhdyt/geminicommit/internal/delivery/cli/handler"
	"github.com/tfkhdyt/geminicommit/internal/service"
)

var (
	prHandler = handler.NewPRHandler()
	draft     = false
)

// prCmd represents the pr command
var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create a pull request with a conventional commit title",
	Long:  `Create a pull request with a conventional commit title`,
	Run: prHandler.PRCommand(
		context.Background(),
		&model,
		&noConfirm,
		&quiet,
		&dryRun,
		&showDiff,
		&maxLength,
		&language,
		&userContext,
		&draft,
	),
}

func init() {
	RootCmd.AddCommand(prCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// prCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	prCmd.Flags().
		BoolVarP(&noConfirm, "yes", "y", noConfirm, "skip confirmation prompt")
	prCmd.Flags().
		BoolVarP(&quiet, "quiet", "q", quiet, "suppress output (only works with --yes)")
	prCmd.Flags().
		StringVarP(&model, "model", "m", service.DefaultModel, "google gemini model to use")
	prCmd.Flags().
		BoolVarP(&dryRun, "dry-run", "d", dryRun, "run the command without making any changes")
	prCmd.Flags().
		BoolVarP(&showDiff, "show-diff", "s", showDiff, "show the diff before creating the pull request")
	prCmd.Flags().
		IntVarP(&maxLength, "max-length", "l", maxLength, "maximum length of the pull request title")
	prCmd.Flags().
		StringVarP(&language, "language", "g", language, "language of the pull request title")
	prCmd.Flags().
		StringVarP(&userContext, "context", "c", "", "additional context to be added to the pull request title")
	prCmd.Flags().
		BoolVar(&draft, "draft", draft, "create a draft pull request")
}
