package usecase

import (
	"context"
	"fmt"
	"os"
	"sync"

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
	atomicCommit *bool,
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
		StageAll:      stageAll,
		UserContext:   userContext,
		Model:         model,
		NoConfirm:     noConfirm,
		Quiet:         quiet,
		Push:          push,
		DryRun:        dryRun,
		ShowDiff:      showDiff,
		MaxLength:     maxLength,
		AtomicCommits: atomicCommit,
		Language:      language,
		Issue:         issue,
		NoVerify:      noVerify,
	}

	// Detect and prepare changes
	data, err := r.gitService.DetectAndPrepareChanges(opts)
	if err != nil {
		return err
	}

	// Display detected files
	r.interactionService.DisplayDetectedFiles(data.Files, opts.Quiet)

	// Show diff if requested
	if *opts.ShowDiff && !*opts.Quiet {
		r.interactionService.DisplayDiff(data.Diff)
	}
	if *opts.AtomicCommits {
		changes, err := r.geminiService.AtomicMessage(client, ctx, data, opts)
		if err != nil {
			return err
		}
		for _, c := range changes {
			selectedAction, finalMessage, err := r.interactionService.HandleUserAction(c.CommitMessage, opts)
			if err != nil {
				return err
			}

			switch selectedAction {
			case service.ActionConfirm:
				if err := r.gitService.ConfirmAction(finalMessage, opts.Quiet, opts.Push, opts.DryRun, opts.NoVerify, c.FileIdentifiers); err != nil {
					return err
				}
				return nil
			case service.ActionRegenerate:
				continue
			case service.ActionEditContext:
				continue
			case service.ActionCancel:
				color.New(color.FgRed).Println("Commit cancelled")
				return nil
			}

		}
		return nil
	}

	// Main generation loop
	for {
		message, err := r.geminiService.GenerateCommitMessage(client, ctx, data, opts)
		if err != nil {
			return err
		}

		selectedAction, finalMessage, err := r.interactionService.HandleUserAction(message, opts)
		if err != nil {
			return err
		}

		switch selectedAction {
		case service.ActionConfirm:
			if err := r.gitService.ConfirmAction(finalMessage, opts.Quiet, opts.Push, opts.DryRun, opts.NoVerify, nil); err != nil {
				return err
			}
			return nil
		case service.ActionRegenerate:
			continue
		case service.ActionEditContext:
			continue
		case service.ActionCancel:
			color.New(color.FgRed).Println("Commit cancelled")
			return nil
		}
	}
}
