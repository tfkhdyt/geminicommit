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

type RootHandler struct {
	useCase *usecase.RootUsecase
}

var (
	rootHandlerInstance *RootHandler
	rootHandlerOnce     sync.Once
)

func NewRootHandler() *RootHandler {
	rootHandlerOnce.Do(func() {
		useCase := usecase.NewRootUsecase()

		rootHandlerInstance = &RootHandler{useCase}
	})

	return rootHandlerInstance
}

func (r *RootHandler) RootCommand(
	ctx context.Context,
	stageAll *bool,
	userContext *string,
	model *string,
) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {
		modelFromConfig := viper.GetString("api.model")
		if modelFromConfig != "" && *model == "" {
			*model = modelFromConfig
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

		err := r.useCase.RootCommand(ctx, apiKey, stageAll, userContext, model)
		cobra.CheckErr(err)
	}
}
