package service

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"github.com/google/generative-ai-go/genai"
)

//go:embed system_prompt.md
var systemPrompt string

type GeminiService struct {
	systemPrompt string
}

var (
	geminiService *GeminiService
	geminiOnce    sync.Once
)

func NewGeminiService() *GeminiService {
	geminiOnce.Do(func() {
		geminiService = &GeminiService{
			systemPrompt: systemPrompt,
		}
	})

	return geminiService
}

func (g *GeminiService) GetUserPrompt(
	context *string,
	diff string,
	files []string,
	// lastCommits []string,
) (string, error) {
	if *context != "" {
		temp := fmt.Sprintf("Use the following context to understand intent: %s", *context)
		context = &temp
	} else {
		*context = ""
	}

	return fmt.Sprintf(
		`%s

Code diff:
%s

Neighboring files:
%s`,
		*context,
		diff,
		strings.Join(files, ", "),
	), nil
}

func (g *GeminiService) AnalyzeChanges(
	geminiClient *genai.Client,
	ctx context.Context,
	diff string,
	userContext *string,
	relatedFiles *map[string]string,
	modelName *string,
	// lastCommits []string,
) (string, error) {
	// format relatedFiles to be dir : files
	relatedFilesArray := make([]string, 0, len(*relatedFiles))
	for dir, ls := range *relatedFiles {
		relatedFilesArray = append(relatedFilesArray, fmt.Sprintf("%s/%s", dir, ls))
	}

	model := geminiClient.GenerativeModel(*modelName)
	model.SetTemperature(0.2)
	safetySettings := []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockNone,
		},
	}

	// Apply safety settings to the model
	model.SafetySettings = safetySettings
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(g.systemPrompt)},
	}

	userPrompt, err := g.GetUserPrompt(userContext, diff, relatedFilesArray)
	if err != nil {
		return "", err
	}

	resp, err := model.GenerateContent(
		ctx,
		genai.Text(userPrompt),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return "", nil
	}

	return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
}
