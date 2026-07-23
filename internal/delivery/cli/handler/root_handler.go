package handler

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tfkhdyt/geminicommit/internal/usecase"
)

type RootHandler struct {
	useCase *usecase.RootUsecase
}

func NewRootHandler() *RootHandler {
	return &RootHandler{useCase: usecase.NewRootUsecase()}
}

func (r *RootHandler) RootCommand(
	ctx context.Context,
	stageAll *bool,
	autoSelect *bool,
	userContext *string,
	model *string,
	noConfirm *bool,
	quiet *bool,
	push *bool,
	dryRun *bool,
	showDiff *bool,
	maxLength *int,
	language *string,
	issue *string,
	issueFooter *string,
	noVerify *bool,
	customBaseUrl *string,
) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {
		if *quiet && !*noConfirm {
			*quiet = false
		}

		apiKey := viper.GetString("api.key")
		if apiKey == "" {
			fmt.Println(
				"Error: API key is still empty, run this command to set your API key",
			)
			fmt.Print("\n")
			color.New(color.Bold).Print("geminicommit config key set ")
			color.New(color.Italic, color.Bold).Print("api_key\n\n")
			os.Exit(1)
		}

		err := r.useCase.RootCommand(ctx, apiKey, stageAll, autoSelect, userContext, model, noConfirm, quiet, push, dryRun, showDiff, maxLength, language, issue, issueFooter, noVerify, customBaseUrl)
		cobra.CheckErr(err)
	}
}
