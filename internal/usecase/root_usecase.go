package usecase

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/charmbracelet/huh/spinner"
	"github.com/fatih/color"
	"google.golang.org/genai"

	"github.com/tfkhdyt/geminicommit/internal/service"
)

type RootUsecase struct {
	gitService         *service.GitService
	geminiService      *service.GeminiService
	interactionService *service.InteractionService
}

var (
	rootUsecaseInstance *RootUsecase
	rootUsecaseOnce     sync.Once
)

func NewRootUsecase() *RootUsecase {
	rootUsecaseOnce.Do(func() {
		gitService := service.NewGitService()
		geminiService := service.NewGeminiService()
		interactionService := service.NewInteractionService()

		rootUsecaseInstance = &RootUsecase{
			gitService:         gitService,
			geminiService:      geminiService,
			interactionService: interactionService,
		}
	})

	return rootUsecaseInstance
}

func (r *RootUsecase) initializeGeminiClient(ctx context.Context, apiKey string, customBaseUrl *string) (*genai.Client, error) {
	baseUrl := ""
	if customBaseUrl != nil {
		baseUrl = *customBaseUrl
	}
	client, err := genai.NewClient(
		ctx,
		&genai.ClientConfig{
			APIKey:  apiKey,
			Backend: genai.BackendGeminiAPI,
			HTTPOptions: genai.HTTPOptions{
				BaseURL: baseUrl,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting gemini client: %v", err)
	}
	return client, nil
}

func (r *RootUsecase) RootCommand(
	ctx context.Context,
	apiKey string,
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
	noVerify *bool,
	customBaseUrl *string,
) error {
	// Initialize Gemini client
	client, err := r.initializeGeminiClient(ctx, apiKey, customBaseUrl)
	if err != nil {
		fmt.Printf("Error getting gemini client: %v", err)
		os.Exit(1)
	}

	// Perform git verifications
	if err := r.gitService.VerifyGitInstallation(); err != nil {
		return err
	}

	if err := r.gitService.VerifyGitRepository(); err != nil {
		return err
	}

	// Prepare commit options
	opts := &service.CommitOptions{
		StageAll:    stageAll,
		AutoSelect:  autoSelect,
		UserContext: userContext,
		Model:       model,
		NoConfirm:   noConfirm,
		Quiet:       quiet,
		Push:        push,
		DryRun:      dryRun,
		ShowDiff:    showDiff,
		MaxLength:   maxLength,
		Language:    language,
		Issue:       issue,
		NoVerify:    noVerify,
	}

	// Detect and prepare changes
	data, err := r.gitService.DetectAndPrepareChanges(opts)
	if err != nil {
		return err
	}

	// Display detected files (skip this in auto mode since AI will select a subset later)
	if !*opts.AutoSelect {
		r.interactionService.DisplayDetectedFiles(data.Files, opts.Quiet)
	}

	// Show diff if requested
	if *opts.ShowDiff && !*opts.Quiet {
		r.interactionService.DisplayDiff(data.Diff)
	}

	// Check if auto-select flag is set and handle accordingly
	var initialCommitMessage string
	if *opts.AutoSelect {
		// Auto flow: Select files with AI and generate commit message in one request
		autoResult, err := r.handleAutoFlow(client, ctx, data, opts)
		if err != nil {
			return err
		}
		data = autoResult.Data // Update data with confirmed files
		initialCommitMessage = autoResult.CommitMessage

		// In auto mode, we need to stage only the selected files for the commit
		// First, unstage everything
		if err := r.gitService.ResetStaged(); err != nil {
			return fmt.Errorf("failed to reset staged files: %v", err)
		}

		// Then stage only the selected files
		if err := r.gitService.StageFiles(data.Files); err != nil {
			return fmt.Errorf("failed to stage selected files: %v", err)
		}
	}

	// Main generation loop
	message := initialCommitMessage
	for {
		// If we don't have a message yet (non-auto mode) or user wants to regenerate, generate one
		if message == "" {
			var err error
			message, err = r.geminiService.GenerateCommitMessage(client, ctx, data, opts)
			if err != nil {
				return err
			}
		}

		selectedAction, finalMessage, err := r.interactionService.HandleUserAction(message, opts)
		if err != nil {
			return err
		}

		switch selectedAction {
		case service.ActionConfirm:
			if err := r.gitService.ConfirmAction(finalMessage, opts.Quiet, opts.Push, opts.DryRun, opts.NoVerify); err != nil {
				return err
			}
			return nil
		case service.ActionRegenerate:
			message = "" // Clear message to regenerate
			continue
		case service.ActionEditContext:
			message = "" // Clear message to regenerate with new context
			continue
		case service.ActionCancel:
			color.New(color.FgRed).Println("Commit cancelled")
			return nil
		}
	}
}

// AutoFlowResult contains both selected files and generated commit message
type AutoFlowResult struct {
	Data          *service.PreCommitData
	CommitMessage string
}

// handleAutoFlow implements the complete auto flow as per the flowchart
func (r *RootUsecase) handleAutoFlow(
	client *genai.Client,
	ctx context.Context,
	data *service.PreCommitData,
	opts *service.CommitOptions,
) (*AutoFlowResult, error) {
	// Step 1: Detect all changes in working directory (already done in calling function)
	// Step 2: Send diff to AI for file selection AND commit message generation
	var selectedFiles []string
	var commitMessage string
	var err error

	if !*opts.Quiet {
		err = spinner.New().
			Title(fmt.Sprintf("AI is analyzing your changes. (Model: %s)", *opts.Model)).
			Action(func() {
				selectedFiles, commitMessage, err = r.geminiService.SelectFilesAndGenerateCommit(
					client,
					ctx,
					data.Diff,
					opts.UserContext,
					&data.RelatedFiles,
					opts.Model,
					opts.MaxLength,
					opts.Language,
					opts.Issue,
				)
			}).
			Run()
		if err != nil {
			return nil, err
		}
	} else {
		selectedFiles, commitMessage, err = r.geminiService.SelectFilesAndGenerateCommit(
			client,
			ctx,
			data.Diff,
			opts.UserContext,
			&data.RelatedFiles,
			opts.Model,
			opts.MaxLength,
			opts.Language,
			opts.Issue,
		)
		if err != nil {
			return nil, err
		}
	}

	// Step 3: Show selected files to user and get their choice
	action, confirmedFiles, err := r.interactionService.ConfirmAutoSelectedFiles(selectedFiles)
	if err != nil {
		return nil, err
	}

	// Step 4: Handle user choice
	switch action {
	case service.ActionCancel:
		return nil, fmt.Errorf("operation cancelled")
	case service.ActionEdit:
		// Open file list editor
		editedFiles, err := r.interactionService.EditFileList(selectedFiles)
		if err != nil {
			return nil, err
		}
		// Update data with edited files
		newData := *data
		newData.Files = editedFiles
		return &AutoFlowResult{
			Data:          &newData,
			CommitMessage: commitMessage,
		}, nil
	case service.ActionConfirm, service.ActionAutoSelect:
		// Proceed with selected files
		newData := *data
		newData.Files = confirmedFiles
		return &AutoFlowResult{
			Data:          &newData,
			CommitMessage: commitMessage,
		}, nil
	default:
		return nil, fmt.Errorf("unknown action: %v", action)
	}
}
