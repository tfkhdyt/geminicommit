package handler

import (
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

func NewRootHandler(useCase *usecase.RootUsecase) *RootHandler {
	return &RootHandler{useCase}
}

func (r *RootHandler) RootCommand(
	stageAll bool,
) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {
		if apiKey := viper.GetString("api.key"); apiKey == "" {
			fmt.Println(
				"Error: API key is still empty, run this command to set your API key",
			)
			fmt.Print("\n")
			color.New(color.Bold).Print("geminicommit config key set ")
			color.New(color.Italic, color.Bold).Print("api_key\n\n")
			os.Exit(1)
		}

		err := r.useCase.RootCommand(stageAll)
		cobra.CheckErr(err)
	}
}
