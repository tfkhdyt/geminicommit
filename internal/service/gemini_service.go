package service

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/genai"
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

	userPrompt, err := g.GetUserPrompt(userContext, diff, relatedFilesArray)
	if err != nil {
		return "", err
	}

	temp := float32(0.2)
	resp, err := geminiClient.Models.GenerateContent(ctx, *modelName, genai.Text(userPrompt), &genai.GenerateContentConfig{
		Temperature: &temp,
		SafetySettings: []*genai.SafetySetting{
			{
				Category:  genai.HarmCategoryHarassment,
				Threshold: genai.HarmBlockThresholdBlockNone,
			},
			{
				Category:  genai.HarmCategoryHateSpeech,
				Threshold: genai.HarmBlockThresholdBlockNone,
			},
			{
				Category:  genai.HarmCategoryDangerousContent,
				Threshold: genai.HarmBlockThresholdBlockNone,
			},
			{
				Category:  genai.HarmCategorySexuallyExplicit,
				Threshold: genai.HarmBlockThresholdBlockNone,
			},
		},
		SystemInstruction: &genai.Content{
			Role:  genai.RoleModel,
			Parts: []*genai.Part{{Text: g.systemPrompt}},
		},
	})
	if err != nil {
		fmt.Println("Error:", err)
		return "", nil
	}

	result := resp.Candidates[0].Content.Parts[0].Text
	result = strings.ReplaceAll(result, "```", "")
	result = strings.TrimSpace(result)

	return result, nil
}
