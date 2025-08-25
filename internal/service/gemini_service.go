package service

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/tfkhdyt/geminicommit/internal/model"
	"github.com/tfkhdyt/geminicommit/internal/stringtools"
	"regexp"
	"strings"
	"sync"

	"github.com/charmbracelet/huh/spinner"
	"github.com/fatih/color"
	"google.golang.org/genai"
)

//go:embed system_prompt.md
var systemPrompt string

var withinBracketRegex = regexp.MustCompile(`\[(.*?)]`)

type GeminiService struct {
	systemPrompt string
}

// CommitOptions contains options for commit generation
type CommitOptions struct {
	StageAll      *bool
	UserContext   *string
	Model         *string
	NoConfirm     *bool
	Quiet         *bool
	Push          *bool
	DryRun        *bool
	ShowDiff      *bool
	AtomicCommits *bool
	MaxLength     *int
	Language      *string
	Issue         *string
	NoVerify      *bool
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

func (g *GeminiService) AtomicMessage(
	client *genai.Client,
	ctx context.Context,
	data *PreCommitData,
	opts *CommitOptions,
) ([]model.CommitChange, error) {
	messageChan := make(chan []model.CommitChange, 1)

	if !*opts.Quiet {
		if err := spinner.New().
			Title(fmt.Sprintf("AI is analyzing your changes. (Model: %s)", *opts.Model)).
			Action(func() {
				g.analyzeAtomicToChannel(client, ctx, data, opts, messageChan)
			}).
			Run(); err != nil {
			return nil, err
		}
	} else {
		g.analyzeAtomicToChannel(client, ctx, data, opts, messageChan)
	}
	changes := <-messageChan
	if !*opts.Quiet {
		underline := color.New(color.Underline)
		underline.Println("\nChanges analyzed!")
	}

	if len(changes) == 0 {
		return nil, fmt.Errorf("no commit messages were generated. try again")
	}

	return changes, nil
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

func (g *GeminiService) analyzeAtomicToChannel(
	client *genai.Client,
	ctx context.Context,
	data *PreCommitData,
	opts *CommitOptions,
	changesChan chan []model.CommitChange,
) {
	changes, err := g.AnalyzeAtomicChange(
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
		changesChan <- []model.CommitChange{{Error: err}}
	} else {
		changesChan <- changes
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

func (g *GeminiService) AtomicChangePrompt(
	context *string,
	diff string,
	files []string,
	maxLength *int,
	language *string,
	issue *string,
	// lastCommits []string,
) (string, error) {
	prompt := ""
	if *context != "" {
		prompt += fmt.Sprintf("Use the following context to understand intent: %s", *context)
	}

	prompt += fmt.Sprintf(
		`%s
how could I split this into meaningful atomic commits  
In json format Only 
[{
commitMessage: "the commit message for the change",
Reason: "reason for atomic change"
fileIdentifiers: ["the file names for the change as array"]
}]
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

func (g *GeminiService) AnalyzeAtomicChange(
	geminiClient *genai.Client,
	ctx context.Context,
	diff string,
	userContext *string,
	relatedFiles *map[string]string,
	modelName *string,
	maxLength *int,
	language *string,
	issue *string,
) ([]model.CommitChange, error) {
	// format relatedFiles to be dir : files
	relatedFilesArray := make([]string, 0, len(*relatedFiles))
	for dir, ls := range *relatedFiles {
		relatedFilesArray = append(relatedFilesArray, fmt.Sprintf("%s/%s", dir, ls))
	}

	atomicPrompt, err := g.AtomicChangePrompt(userContext, diff, relatedFilesArray, maxLength, language, issue)
	if err != nil {
		return nil, err
	}

	// Update system prompt to include language and length requirements
	enhancedSystemPrompt := g.systemPrompt

	temp := float32(0.2)
	resp, err := geminiClient.Models.GenerateContent(ctx, *modelName, genai.Text(atomicPrompt), &genai.GenerateContentConfig{
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
		return nil, err
	}
	var changes []model.CommitChange
	text := resp.Candidates[0].Content.Parts[0].Text
	text = stringtools.Subslice(text, "[", "]")
	err = json.Unmarshal([]byte(text), &changes)
	return changes, err
}
