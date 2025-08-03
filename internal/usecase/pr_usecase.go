/*
Copyright © 2025 Christina Sørensen <ces@fem.gg>
*/
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

type PRUsecase struct {
	gitService         *service.GitService
	geminiService      *service.GeminiService
	interactionService *service.InteractionService
}

var (
	prUsecaseInstance *PRUsecase
	prUsecaseOnce     sync.Once
)

func NewPRUsecase() *PRUsecase {
	prUsecaseOnce.Do(func() {
		gitService := service.NewGitService()
		geminiService := service.NewGeminiService()
		interactionService := service.NewInteractionService()

		prUsecaseInstance = &PRUsecase{
			gitService:         gitService,
			geminiService:      geminiService,
			interactionService: interactionService,
		}
	})

	return prUsecaseInstance
}

func (p *PRUsecase) initializeGeminiClient(
	ctx context.Context,
	apiKey string,
	customBaseUrl *string,
) (*genai.Client, error) {
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

func (p *PRUsecase) PRCommand(
	ctx context.Context,
	apiKey string,
	model *string,
	noConfirm *bool,
	quiet *bool,
	dryRun *bool,
	showDiff *bool,
	maxLength *int,
	language *string,
	userContext *string,
	draft *bool,
	customBaseUrl *string,
) error {
	client, err := p.initializeGeminiClient(ctx, apiKey, customBaseUrl)
	if err != nil {
		fmt.Printf("Error getting gemini client: %v", err)
		os.Exit(1)
	}

	if err := p.gitService.VerifyGitInstallation(); err != nil {
		return err
	}

	if err := p.gitService.VerifyGitRepository(); err != nil {
		return err
	}

	opts := &service.CommitOptions{
		Model:       model,
		NoConfirm:   noConfirm,
		Quiet:       quiet,
		DryRun:      dryRun,
		ShowDiff:    showDiff,
		MaxLength:   maxLength,
		Language:    language,
		UserContext: userContext,
	}

	data, err := p.gitService.GetDiff()
	if err != nil {
		return err
	}

	if *opts.ShowDiff && !*opts.Quiet {
		p.interactionService.DisplayDiff(data.Diff)
	}

	for {
		message, err := p.geminiService.GenerateCommitMessage(client, ctx, data, opts)
		if err != nil {
			return err
		}

		selectedAction, finalMessage, err := p.interactionService.HandleUserAction(
			message,
			opts,
		)
		if err != nil {
			return err
		}

		switch selectedAction {
		case service.ActionConfirm:
			if err := p.gitService.CreatePullRequest(
				finalMessage,
				opts.Quiet,
				opts.DryRun,
				draft,
			); err != nil {
				return err
			}
			return nil
		case service.ActionRegenerate:
			continue
		case service.ActionEditContext:
			continue
		case service.ActionCancel:
			color.New(color.FgRed).Println("Pull request cancelled")
			return nil
		}
	}
}
