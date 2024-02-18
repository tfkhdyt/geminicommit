package gemini

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

func AnalyzeChanges(diff string, messageChan chan<- string) func() {
	return func() {
		ctx := context.Background()

		client, err := genai.NewClient(
			ctx,
			option.WithAPIKey(viper.GetString("api.key")),
		)
		cobra.CheckErr(err)
		defer client.Close()

		model := client.GenerativeModel("gemini-pro")
		resp, err := model.GenerateContent(
			ctx,
			genai.Text(
				fmt.Sprintf(
					`generate conventional git commit message based on the following diff:
%s

Exclude anything unnecessary, because your entire response will be passed directly into git commit`,
					diff,
				),
			),
		)
		cobra.CheckErr(err)

		messageChan <- fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	}
}
