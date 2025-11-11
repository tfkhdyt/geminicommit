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

//go:embed file_selection_prompt.md
var fileSelectionPrompt string

//go:embed combined_prompt.md
var combinedPrompt string

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

// SelectFilesAndGenerateCommitOptions contains optional parameters for SelectFilesAndGenerateCommit
type SelectFilesAndGenerateCommitOptions struct {
	UserContext  *string
	RelatedFiles *map[string]string
	ModelName    *string
	MaxLength    *int
	Language     *string
	Issue        *string
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

// formatRelatedFiles formats a map of directory to files into a slice of strings
// in the format "dir/file"
func formatRelatedFiles(dirToFiles map[string]string) []string {
	relatedFilesArray := make([]string, 0, len(dirToFiles))
	for dir, ls := range dirToFiles {
		relatedFilesArray = append(relatedFilesArray, fmt.Sprintf("%s/%s", dir, ls))
	}
	return relatedFilesArray
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
	relatedFilesArray := formatRelatedFiles(*relatedFiles)

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

	enhancedSystemPrompt := fileSelectionPrompt

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

// SelectFilesAndGenerateCommit combines file selection and commit message generation in a single AI request
func (g *GeminiService) SelectFilesAndGenerateCommit(
	geminiClient *genai.Client,
	ctx context.Context,
	diff string,
	opts *SelectFilesAndGenerateCommitOptions,
) ([]string, string, error) {
	// Validate required parameters
	if opts == nil {
		return nil, "", fmt.Errorf("options cannot be nil")
	}
	if opts.ModelName == nil {
		return nil, "", fmt.Errorf("ModelName cannot be nil")
	}
	if opts.MaxLength == nil {
		return nil, "", fmt.Errorf("MaxLength cannot be nil")
	}
	if opts.Language == nil {
		return nil, "", fmt.Errorf("Language cannot be nil")
	}
	if opts.RelatedFiles == nil {
		return nil, "", fmt.Errorf("RelatedFiles cannot be nil")
	}

	// Format relatedFiles to be dir : files
	relatedFilesArray := formatRelatedFiles(*opts.RelatedFiles)

	// Build user prompt with context, diff, and requirements
	contextStr := ""
	if opts.UserContext != nil && *opts.UserContext != "" {
		contextStr = fmt.Sprintf("Use the following context to understand intent: %s\n\n", *opts.UserContext)
	}

	prompt := fmt.Sprintf(
		`%sHere's the code diff:
%s

Neighboring files:
%s

Requirements:
- Maximum commit message length: %d characters
- Language: %s`,
		contextStr,
		diff,
		strings.Join(relatedFilesArray, ", "),
		*opts.MaxLength,
		*opts.Language,
	)

	if opts.Issue != nil && *opts.Issue != "" {
		prompt += fmt.Sprintf("\n- Reference issue: %s", *opts.Issue)
	}

	// Build enhanced system prompt
	enhancedSystemPrompt := combinedPrompt
	if *opts.Language != "english" {
		enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Generate the commit message in %s language.", *opts.Language)
	}
	enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Keep the commit message under %d characters.", *opts.MaxLength)
	if opts.Issue != nil && *opts.Issue != "" {
		enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Reference issue %s in the commit message.", *opts.Issue)
	}

	temp := float32(0.2)
	resp, err := geminiClient.Models.GenerateContent(ctx, *opts.ModelName, genai.Text(prompt), &genai.GenerateContentConfig{
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
		return nil, "", err
	}

	// Defensive checks to prevent panics
	if resp == nil {
		return nil, "", fmt.Errorf("API response is nil")
	}
	if len(resp.Candidates) == 0 {
		return nil, "", fmt.Errorf("API response contains no candidates")
	}
	if resp.Candidates[0].Content == nil {
		return nil, "", fmt.Errorf("API response candidate content is nil")
	}
	if len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, "", fmt.Errorf("API response candidate content contains no parts")
	}
	if resp.Candidates[0].Content.Parts[0] == nil {
		return nil, "", fmt.Errorf("API response candidate part is nil")
	}
	if resp.Candidates[0].Content.Parts[0].Text == "" {
		return nil, "", fmt.Errorf("API response candidate part text is empty")
	}

	result := strings.TrimSpace(resp.Candidates[0].Content.Parts[0].Text)

	// Parse files from response
	var filesStr string
	lines := strings.Split(result, "\n")
	foundFilesSection := false
	var filesLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if this line starts the FILES section
		if strings.HasPrefix(trimmedLine, "FILES:") {
			foundFilesSection = true
			// Extract the part after "FILES:"
			afterPrefix := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "FILES:"))
			if afterPrefix != "" {
				filesLines = append(filesLines, afterPrefix)
			}
			continue
		}

		// If we're in the files section, collect lines until we hit COMMIT_MESSAGE
		if foundFilesSection {
			// Stop if we hit COMMIT_MESSAGE section
			if strings.HasPrefix(trimmedLine, "COMMIT_MESSAGE:") {
				break
			}
			filesLines = append(filesLines, line)
		}
	}

	if len(filesLines) > 0 {
		filesStr = strings.TrimSpace(strings.Join(filesLines, " "))
	} else {
		// Fallback: try to find files using the old method
		if after, ok := strings.CutPrefix(result, "FILES:"); ok {
			// Stop at COMMIT_MESSAGE if present
			if idx := strings.Index(after, "COMMIT_MESSAGE:"); idx != -1 {
				filesStr = strings.TrimSpace(after[:idx])
			} else {
				filesStr = strings.TrimSpace(after)
			}
		}
	}

	if filesStr == "" {
		return nil, "", fmt.Errorf("AI response did not include file list in expected format. Response was: %s", result)
	}

	// Parse the file list
	files := strings.Split(filesStr, ",")
	for i, f := range files {
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

	// Parse commit message from response
	var commitMessage string
	foundCommitSection := false
	var commitLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if this line starts the COMMIT_MESSAGE section
		if strings.HasPrefix(trimmedLine, "COMMIT_MESSAGE:") {
			foundCommitSection = true
			// Extract the part after "COMMIT_MESSAGE:"
			afterPrefix := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "COMMIT_MESSAGE:"))
			if afterPrefix != "" {
				commitLines = append(commitLines, afterPrefix)
			}
			continue
		}

		// If we're in the commit message section, collect lines
		if foundCommitSection {
			// Stop if we hit another section marker (like FILES:)
			if strings.HasPrefix(trimmedLine, "FILES:") {
				break
			}
			// Collect the line (preserve empty lines for commit message formatting)
			commitLines = append(commitLines, line)
		}
	}

	if len(commitLines) > 0 {
		commitMessage = strings.Join(commitLines, "\n")
		// Remove any markdown code blocks
		commitMessage = strings.ReplaceAll(commitMessage, "```", "")
		commitMessage = strings.TrimSpace(commitMessage)
	}

	if commitMessage == "" {
		return nil, "", fmt.Errorf("AI response did not include commit message in expected format. Response was: %s", result)
	}

	return validFiles, commitMessage, nil
}
