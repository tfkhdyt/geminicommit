/*
Copyright © 2025 Christina Sørensen <ces@fem.gg>
*/
package handler

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tfkhdyt/geminicommit/internal/usecase"
)

type PRHandler struct {
	useCase *usecase.PRUsecase
}

var (
	prHandlerInstance *PRHandler
	prHandlerOnce     sync.Once
)

func NewPRHandler() *PRHandler {
	prHandlerOnce.Do(func() {
		useCase := usecase.NewPRUsecase()

		prHandlerInstance = &PRHandler{useCase}
	})

	return prHandlerInstance
}

func (p *PRHandler) PRCommand(
	ctx context.Context,
	model *string,
	noConfirm *bool,
	quiet *bool,
	dryRun *bool,
	showDiff *bool,
	maxLength *int,
	language *string,
	userContext *string,
) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {
		modelFromConfig := viper.GetString("api.model")
		if modelFromConfig != "" && *model == "gemini-2.5-flash" {
			*model = modelFromConfig
		}

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

		err := p.useCase.PRCommand(ctx, apiKey, model, noConfirm, quiet, dryRun, showDiff, maxLength, language, userContext)
		cobra.CheckErr(err)
	}
}
