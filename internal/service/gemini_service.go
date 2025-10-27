package service

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/huh/spinner"
	"github.com/fatih/color"
	"google.golang.org/genai"
)

//go:embed system_prompt.md
var systemPrompt string

type GeminiService struct {
	systemPrompt string
}

// CommitOptions contains options for commit generation
type CommitOptions struct {
	StageAll    *bool
	AutoSelect  *bool
	UserContext *string
	Model       *string
	NoConfirm   *bool
	Quiet       *bool
	Push        *bool
	DryRun      *bool
	ShowDiff    *bool
	MaxLength   *int
	Language    *string
	Issue       *string
	NoVerify    *bool
}

// PreCommitData contains data about the changes to be committed
type PreCommitData struct {
	Files        []string
	Diff         string
	RelatedFiles map[string]string
	Issue        string
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

// GenerateCommitMessage creates a commit message using AI analysis with UI feedback
func (g *GeminiService) GenerateCommitMessage(
	client *genai.Client,
	ctx context.Context,
	data *PreCommitData,
	opts *CommitOptions,
) (string, error) {
	// Handle --auto flag
	if *opts.AutoSelect {
		selectedFiles, err := g.SelectFilesUsingAI(client, ctx, data.Diff, opts.UserContext, opts.Model)
		if err != nil {
			return "", err
		}

		interactionService := NewInteractionService()
		action, confirmedFiles, err := interactionService.ConfirmAutoSelectedFiles(selectedFiles)
		if err != nil || action == ActionCancel {
			return "", fmt.Errorf("operation cancelled")
		}
		if action == ActionAutoSelect || action == ActionConfirm {
			data.Files = confirmedFiles
		}
	}

	messageChan := make(chan string, 1)

	if !*opts.Quiet {
		if err := spinner.New().
			Title(fmt.Sprintf("AI is analyzing your changes. (Model: %s)", *opts.Model)).
			Action(func() {
				g.analyzeToChannel(client, ctx, data, opts, messageChan)
			}).
			Run(); err != nil {
			return "", err
		}
	} else {
		g.analyzeToChannel(client, ctx, data, opts, messageChan)
	}

	message := <-messageChan
	if !*opts.Quiet {
		underline := color.New(color.Underline)
		underline.Println("\nChanges analyzed!")
	}

	message = strings.TrimSpace(message)
	if message == "" {
		return "", fmt.Errorf("no commit messages were generated. try again")
	}

	return message, nil
}

// analyzeToChannel performs the actual AI analysis and sends result to channel
func (g *GeminiService) analyzeToChannel(
	client *genai.Client,
	ctx context.Context,
	data *PreCommitData,
	opts *CommitOptions,
	messageChan chan string,
) {
	message, err := g.AnalyzeChanges(
		client,
		ctx,
		data.Diff,
		opts.UserContext,
		&data.RelatedFiles,
		opts.Model,
		opts.MaxLength,
		opts.Language,
		&data.Issue,
	)
	if err != nil {
		messageChan <- ""
	} else {
		messageChan <- message
	}
}

func (g *GeminiService) GetUserPrompt(
	context *string,
	diff string,
	files []string,
	maxLength *int,
	language *string,
	issue *string,
	// lastCommits []string,
) (string, error) {
	if *context != "" {
		temp := fmt.Sprintf("Use the following context to understand intent: %s", *context)
		context = &temp
	} else {
		*context = ""
	}

	prompt := fmt.Sprintf(
		`%s

Code diff:
%s

Neighboring files:
%s

Requirements:
- Maximum commit message length: %d characters
- Language: %s`,
		*context,
		diff,
		strings.Join(files, ", "),
		*maxLength,
		*language,
	)

	if *issue != "" {
		prompt += fmt.Sprintf("\n- Reference issue: %s", *issue)
	}

	return prompt, nil
}

func (g *GeminiService) AnalyzeChanges(
	geminiClient *genai.Client,
	ctx context.Context,
	diff string,
	userContext *string,
	relatedFiles *map[string]string,
	modelName *string,
	maxLength *int,
	language *string,
	issue *string,
	// lastCommits []string,
) (string, error) {
	// format relatedFiles to be dir : files
	relatedFilesArray := make([]string, 0, len(*relatedFiles))
	for dir, ls := range *relatedFiles {
		relatedFilesArray = append(relatedFilesArray, fmt.Sprintf("%s/%s", dir, ls))
	}

	userPrompt, err := g.GetUserPrompt(userContext, diff, relatedFilesArray, maxLength, language, issue)
	if err != nil {
		return "", err
	}

	// Update system prompt to include language and length requirements
	enhancedSystemPrompt := g.systemPrompt
	if *language != "english" {
		enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Generate the commit message in %s language.", *language)
	}
	enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Keep the commit message under %d characters.", *maxLength)
	if *issue != "" {
		enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Reference issue %s in the commit message.", *issue)
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
			Parts: []*genai.Part{{Text: enhancedSystemPrompt}},
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

// SelectFilesUsingAI lets the AI determine which files to stage based on the diff and context
func (g *GeminiService) SelectFilesUsingAI(
	geminiClient *genai.Client,
	ctx context.Context,
	diff string,
	userContext *string,
	modelName *string,
) ([]string, error) {
	prompt := fmt.Sprintf(
		`%s
Here's the code diff:
%s`,
		*userContext,
		diff,
	)

	// Use empty system prompt since we're not using g.systemPrompt
	enhancedSystemPrompt := "You are an assistant that helps developers decide which files to include in a git commit. Respond ONLY with the list of files to stage in the format: \"FILES: file1, file2, ...\"."

	temp := float32(0.2)
	resp, err := geminiClient.Models.GenerateContent(ctx, *modelName, genai.Text(prompt), &genai.GenerateContentConfig{
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
			Parts: []*genai.Part{{Text: enhancedSystemPrompt}},
		},
	})
	if err != nil {
		return nil, err
	}

	result := strings.TrimSpace(resp.Candidates[0].Content.Parts[0].Text)

	// Look for the file list in the response with more flexible matching
	var filesStr string
	if after, ok := strings.CutPrefix(result, "FILES:"); ok {
		// Direct match
		filesStr = after
	} else {
		// Try to find "FILES: ..." anywhere in the response
		lines := strings.Split(result, "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if after, ok := strings.CutPrefix(trimmedLine, "FILES:"); ok {
				filesStr = after
				break
			}
		}

		// If still not found, try to find any line that starts with a file path pattern
		if filesStr == "" {
			// Try to find pattern like: file1, file2, file3 or list of individual files
			// Look for common patterns in the response
			if idx := strings.Index(result, ":"); idx != -1 {
				potential := result[idx+1:]
				// Check if this contains common file extensions or paths
				if strings.Contains(potential, ".") || strings.Contains(potential, "/") || strings.Contains(potential, ",") {
					filesStr = potential
				}
			}
		}
	}

	if filesStr == "" {
		return nil, fmt.Errorf("AI response did not include file list in expected format. Response was: %s", result)
	}

	// Parse the file list
	files := strings.Split(filesStr, ",")
	for i, f := range files {
		// Remove any markdown formatting like backticks
		f = strings.Trim(f, "` \t\n\r")
		files[i] = strings.TrimSpace(f)
	}

	// Filter out empty strings
	var validFiles []string
	for _, f := range files {
		if f != "" {
			validFiles = append(validFiles, f)
		}
	}

	return validFiles, nil
}
