package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type GitService struct{}

var (
	instance *GitService
	once     sync.Once
)

func NewGitService() *GitService {
	once.Do(func() {
		instance = &GitService{}
	})

	return instance
}

func (g *GitService) VerifyGitInstallation() error {
	if err := exec.Command("git", "--version").Run(); err != nil {
		return fmt.Errorf("git is not installed. %v", err)
	}

	return nil
}

func (g *GitService) VerifyGitRepository() error {
	if err := exec.Command("git", "rev-parse", "--show-toplevel").Run(); err != nil {
		return fmt.Errorf(
			"the current directory must be a git repository. %v",
			err,
		)
	}

	return nil
}

func (g *GitService) StageAll() error {
	if err := exec.Command("git", "add", "-u").Run(); err != nil {
		return fmt.Errorf("failed to update tracked files. %v", err)
	}

	return nil
}

func (g *GitService) DetectDiffChanges() ([]string, string, error) {
	files, err := exec.Command("git", "diff", "--cached", "--diff-algorithm=minimal", "--name-only").
		Output()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, "", err
	}
	filesStr := strings.TrimSpace(string(files))

	if filesStr == "" {
		return nil, "", fmt.Errorf("nothing to be analyze")
	}

	diff, err := exec.Command("git", "diff", "--cached", "--diff-algorithm=minimal").
		Output()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, "", err
	}

	return strings.Split(filesStr, "\n"), string(diff), nil
}

func (g *GitService) CommitChanges(message string, quiet *bool) error {
	cmd := exec.Command("git", "commit", "-m", message)
	if !*quiet {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes. %v", err)
	}

	return nil
}

func (g *GitService) GetLastCommitMessages(count int) ([]string, error) {
	// Command to get only commit messages
	cmd := exec.Command("git", "log",
		"--pretty=format:%s", // %s formats only the commit message
		"-n", fmt.Sprintf("%d", count))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("git log error: %v: %s", err, stderr.String())
	}

	// Split output into lines and filter empty ones
	messages := strings.Split(out.String(), "\n")
	result := make([]string, 0, len(messages))
	for _, msg := range messages {
		if msg != "" {
			result = append(result, msg)
		}
	}

	return result, nil
}
