package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func VerifyGitInstallation() error {
	if err := exec.Command("git", "-v").Run(); err != nil {
		return fmt.Errorf("git is not installed. %v", err)
	}

	return nil
}

func VerifyGitRepository() error {
	if err := exec.Command("git", "rev-parse", "--show-toplevel").Run(); err != nil {
		return fmt.Errorf(
			"the current directory must be a Git repository. %v",
			err,
		)
	}

	return nil
}

func StageAll() error {
	if err := exec.Command("git", "add", "-u").Run(); err != nil {
		return fmt.Errorf("failed to update tracked files. %v", err)
	}

	return nil
}

func DetectDiffChanges(
	filesChan chan<- []string,
	diffChan chan<- string,
) func() {
	return func() {
		files, err := exec.Command("git", "diff", "--cached", "--diff-algorithm=minimal", "--name-only").
			Output()
		cobra.CheckErr(err)
		filesStr := strings.TrimSpace(string(files))

		if filesStr == "" {
			filesChan <- []string{}
			diffChan <- ""
			return
		}

		diff, err := exec.Command("git", "diff", "--cached", "--diff-algorithm=minimal").
			Output()
		cobra.CheckErr(err)

		filesChan <- strings.Split(filesStr, "\n")
		diffChan <- string(diff)
	}
}

func CommitChanges(message string) error {
	output, err := exec.Command("git", "commit", "-m", message).Output()
	if err != nil {
		return fmt.Errorf("failed to commit changes. %v", err)
	}

	fmt.Println(string(output))

	return nil
}
