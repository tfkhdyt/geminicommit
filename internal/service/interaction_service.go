package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"google.golang.org/genai"
)

// Action represents user actions
type Action string

const (
	ActionConfirm     Action = "CONFIRM"
	ActionRegenerate  Action = "REGENERATE"
	ActionEdit        Action = "EDIT"
	ActionEditContext Action = "EDIT_CONTEXT"
	ActionCancel      Action = "CANCEL"
	ActionAutoSelect  Action = "AUTO_SELECT"
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

// ConfirmAutoSelectedFiles prompts the user to confirm, edit, or cancel AI-selected files
func (h *InteractionService) ConfirmAutoSelectedFiles(files []string) (Action, []string, error) {
	var choice string
	options := []string{"Yes", "Edit", "Cancel"}

	// Format the file list for display
	fileList := ""
	for _, f := range files {
		// Escape markdown special characters in file paths that could break UI
		escapedFile := strings.ReplaceAll(f, "_", "\\_")
		fileList += fmt.Sprintf("- %s\n", escapedFile)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title("AI Selected Files").Description(fileList),
			huh.NewSelect[string]().
				Title("Proceed with these files?").
				Options(huh.NewOptions(options...)...).
				Value(&choice),
		),
	)

	if err := form.Run(); err != nil {
		return ActionCancel, nil, err
	}

	switch choice {
	case "Cancel":
		return ActionCancel, nil, nil
	case "Edit":
		editedFiles, err := h.EditFileList(files)
		if err != nil {
			return ActionCancel, nil, err
		}
		return ActionAutoSelect, editedFiles, nil
	default:
		return ActionConfirm, files, nil
	}
}

// EditFileList allows the user to select files from the list
func (h *InteractionService) EditFileList(files []string) ([]string, error) {
	// // Create options from the files, with all initially selected
	// options := make([]huh.Option[string], len(files))
	// for i, file := range files {
	// 	options[i] = huh.NewOption(file, file).Selected(true)
	// }

	// Variable to store selected files
	var selectedFiles []string

	// Create the multi-select form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select files to include in commit").
				Options(huh.NewOptions(files...)...).
				Value(&selectedFiles).
				Height(15),
		),
	)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	if len(selectedFiles) == 0 {
		// Show appropriate message when no files are selected
		fmt.Println("No files selected. Operation cancelled.")
		return nil, errors.New("no files selected")
	}

	return selectedFiles, nil
}

// AutoFlow orchestrates the complete auto flow using the huh library
func (h *InteractionService) AutoFlow(geminiClient *genai.Client, ctx context.Context, data *PreCommitData, opts *CommitOptions) (*PreCommitData, error) {
	// Verify Git installation and repository (this is already done in the usecase)
	// Detect all changes in working directory (this is already done in the usecase)

	// Send diff to AI for file selection
	selectedFiles, err := h.SelectFilesUsingAI(geminiClient, ctx, data.Diff, opts.UserContext, opts.Model)
	if err != nil {
		return nil, err
	}

	// Show selected files to user
	action, confirmedFiles, err := h.ConfirmAutoSelectedFiles(selectedFiles)
	if err != nil || action == ActionCancel {
		return nil, fmt.Errorf("operation cancelled")
	}

	if action == ActionAutoSelect || action == ActionConfirm {
		data.Files = confirmedFiles
	} else if action == ActionEdit {
		// If user wants to edit, open file list editor
		editedFiles, err := h.EditFileList(selectedFiles)
		if err != nil {
			return nil, err
		}
		data.Files = editedFiles
	}

	return data, nil
}

// SelectFilesUsingAI is a wrapper method to call the AI file selection from the interaction service
func (h *InteractionService) SelectFilesUsingAI(geminiClient *genai.Client, ctx context.Context, diff string, userContext *string, modelName *string) ([]string, error) {
	// Use the gemini service to select files using AI
	geminiService := NewGeminiService()
	return geminiService.SelectFilesUsingAI(geminiClient, ctx, diff, userContext, modelName)
}
