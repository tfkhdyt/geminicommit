package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/fatih/color"

	"github.com/tfkhdyt/geminicommit/internal/service"
)

type action string

const (
	confirm     action = "CONFIRM"
	regenerate  action = "REGENERATE"
	edit        action = "EDIT"
	editcontext action = "EDIT_CONTEXT"
	cancel      action = "CANCEL"
)

type RootUsecase struct {
	gitService    *service.GitService
	geminiService *service.GeminiService
}

func NewRootUsecase(
	gitService *service.GitService,
	geminiService *service.GeminiService,
) *RootUsecase {
	return &RootUsecase{gitService, geminiService}
}

func (r *RootUsecase) RootCommand(stageAll *bool, userContext *string, model *string) error {
	if err := r.gitService.VerifyGitInstallation(); err != nil {
		return err
	}

	if err := r.gitService.VerifyGitRepository(); err != nil {
		return err
	}

	if *stageAll {
		if err := r.gitService.StageAll(); err != nil {
			return err
		}
	}

	filesChan := make(chan []string, 1)
	diffChan := make(chan string, 1)

	if err := spinner.New().
		Title("Detecting staged files").
		Action(func() {
			files, diff, err := r.gitService.DetectDiffChanges()
			if err != nil {
				filesChan <- []string{}
				diffChan <- ""
				return
			}

			filesChan <- files
			diffChan <- diff
		}).
		Run(); err != nil {
		return err
	}

	underline := color.New(color.Underline)
	files, diff := <-filesChan, <-diffChan

	if len(files) == 0 {
		return fmt.Errorf(
			"no staged changes found. stage your changes manually, or automatically stage all changes with the `--all` flag",
		)
	} else if len(files) == 1 {
		underline.Printf("Detected %d staged file:\n", len(files))
	} else {
		underline.Printf("Detected %d staged files:\n", len(files))
	}
	relatedFiles := make(map[string]string)
	visitedDirs := make(map[string]bool)
	for idx, file := range files {
		color.New(color.Bold).Printf("     %d. %s\n", idx+1, file)
		dir := filepath.Dir(file)
		if !visitedDirs[dir] {
			lsEntry, err := os.ReadDir(dir)
			if err == nil {
				var ls []string
				for _, entry := range lsEntry {
					ls = append(ls, entry.Name())
				}
				relatedFiles[dir] = strings.Join(ls, ", ")
				visitedDirs[dir] = true
			}
		}
	}

generate:
	for {
		fmt.Println("Model:", *model)
		messageChan := make(chan string, 1)
		if err := spinner.New().
			Title(fmt.Sprintf("AI is analyzing your changes. (Model: %s)", *model)).
			Action(func() {
				message, err := r.geminiService.AnalyzeChanges(context.Background(), diff, userContext, &relatedFiles, model)
				if err != nil {
					messageChan <- ""
					return
				}

				messageChan <- message
			}).
			Run(); err != nil {
			return err
		}

		message := <-messageChan
		fmt.Print("\n")
		underline.Println("Changes analyzed!")

		message = strings.TrimSpace(message)

		if message == "" {
			return fmt.Errorf("no commit messages were generated. try again")
		}

		color.New(color.Bold).Printf("%s\n\n", message)

		var selectedAction action
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[action]().
					Title("Use this commit?").
					Options(
						huh.NewOption("Yes", confirm),
						huh.NewOption("Regenerate", regenerate),
						huh.NewOption("Edit", edit),
						huh.NewOption("Edit Context", editcontext),
						huh.NewOption("Cancel", cancel),
					).
					Value(&selectedAction),
			),
		).Run(); err != nil {
			return err
		}

		switch selectedAction {
		case confirm:
			if err := r.gitService.CommitChanges(message); err != nil {
				return err
			}
			color.New(color.FgGreen).Println("✔ Successfully committed!")
			break generate
		case regenerate:
			continue
		case edit:
			if err := huh.NewForm(
				huh.NewGroup(
					huh.NewText().Title("Edit commit message manually").CharLimit(1000).Value(&message),
				),
			).Run(); err != nil {
				return err
			}

			if err := r.gitService.CommitChanges(message); err != nil {
				return err
			}
			color.New(color.FgGreen).Println("✔ Successfully committed!")
			break generate
		case editcontext:
			if err := huh.NewForm(
				huh.NewGroup(
					huh.NewText().Title("Edit user context").CharLimit(1000).Value(userContext),
				),
			).Run(); err != nil {
				return err
			}
			continue
		case cancel:
			color.New(color.FgRed).Println("Commit cancelled")
			break generate
		}
	}

	return nil
}
