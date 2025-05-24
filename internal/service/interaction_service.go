package service

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
)

// Action represents user actions
type Action string

const (
	ActionConfirm     Action = "CONFIRM"
	ActionRegenerate  Action = "REGENERATE"
	ActionEdit        Action = "EDIT"
	ActionEditContext Action = "EDIT_CONTEXT"
	ActionCancel      Action = "CANCEL"
)

// InteractionService manages user interactions and UI
type InteractionService struct{}

func NewInteractionService() *InteractionService {
	return &InteractionService{}
}

// HandleUserAction presents the user with action options and processes their choice
func (h *InteractionService) HandleUserAction(message string, opts *CommitOptions) (Action, string, error) {
	if *opts.NoConfirm {
		return ActionConfirm, message, nil
	}

	color.New(color.Bold).Printf("%s\n\n", message)

	var selectedAction Action
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[Action]().
				Title("Use this commit?").
				Options(
					huh.NewOption("Yes", ActionConfirm),
					huh.NewOption("Regenerate", ActionRegenerate),
					huh.NewOption("Edit", ActionEdit),
					huh.NewOption("Edit Context", ActionEditContext),
					huh.NewOption("Cancel", ActionCancel),
				).
				Value(&selectedAction),
		),
	).Run(); err != nil {
		return "", "", err
	}

	switch selectedAction {
	case ActionEdit:
		editedMessage, err := h.EditCommitMessage(message)
		if err != nil {
			return "", "", err
		}
		return ActionConfirm, editedMessage, nil
	case ActionEditContext:
		if err := h.EditContext(opts.UserContext); err != nil {
			return "", "", err
		}
		return ActionEditContext, message, nil
	default:
		return selectedAction, message, nil
	}
}

// EditCommitMessage allows the user to manually edit the commit message
func (h *InteractionService) EditCommitMessage(originalMessage string) (string, error) {
	message := originalMessage
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewText().Title("Edit commit message manually").CharLimit(1000).Value(&message),
		),
	).Run(); err != nil {
		return "", err
	}
	return message, nil
}

// EditContext allows the user to edit the user context
func (h *InteractionService) EditContext(userContext *string) error {
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewText().Title("Edit user context").CharLimit(1000).Value(userContext),
		),
	).Run(); err != nil {
		return err
	}
	return nil
}

// DisplayDetectedFiles shows the detected staged files to the user
func (h *InteractionService) DisplayDetectedFiles(files []string, quiet *bool) {
	if *quiet {
		return
	}

	underline := color.New(color.Underline)
	if len(files) == 1 {
		underline.Printf("Detected %d staged file:\n", len(files))
	} else {
		underline.Printf("Detected %d staged files:\n", len(files))
	}

	// List the files
	for idx, file := range files {
		color.New(color.Bold).Printf("     %d. %s\n", idx+1, file)
	}
}

// DisplayDiff shows the git diff to the user
func (h *InteractionService) DisplayDiff(diff string) {
	underline := color.New(color.Underline)
	underline.Println("\nDiff to be analyzed:")
	fmt.Println(diff)
	fmt.Println()
}
