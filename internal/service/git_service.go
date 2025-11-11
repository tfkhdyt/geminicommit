package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/charmbracelet/huh/spinner"
	"github.com/fatih/color"
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
	if err := exec.Command("git", "add", "--all").Run(); err != nil {
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

func (g *GitService) GetAllChanges() ([]string, error) {
	// Get all changed files (including untracked, modified, deleted, etc.)
	cmd := exec.Command("git", "status", "--porcelain", "--untracked-files=all")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting all changes:", err)
		return nil, err
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return []string{}, nil
	}

	// Parse the git status output to get just the filenames
	lines := strings.Split(outputStr, "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			// The git status --porcelain format has status followed by space and filename
			// Examples: "M file", "?? file", "MM file", " A file"
			line = strings.TrimSpace(line)
			if len(line) > 2 { // At least "X " + filename
				// Find the first space after the status and extract filename
				spaceIndex := strings.Index(line, " ")
				if spaceIndex != -1 && spaceIndex < len(line)-1 {
					// Extract filename after the first space
					filename := strings.TrimSpace(line[spaceIndex+1:])
					if filename != "" {
						files = append(files, filename)
					}
				} else if len(line) > 2 {
					// Fallback: if no space found, take everything after the first 3 characters
					// This handles edge cases, though they shouldn't normally occur
					filename := strings.TrimSpace(line[2:])
					if filename != "" {
						files = append(files, filename)
					}
				}
			}
		}
	}

	return files, nil
}

// GetAllChangesWithStatus returns all changed files along with their git status
func (g *GitService) GetAllChangesWithStatus() ([]string, map[string]string, error) {
	// Get all changed files (including untracked, modified, deleted, etc.)
	cmd := exec.Command("git", "status", "--porcelain", "--untracked-files=all")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting all changes:", err)
		return nil, nil, err
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return []string{}, map[string]string{}, nil
	}

	// Parse the git status output to get filenames and their status
	lines := strings.Split(outputStr, "\n")
	var files []string
	fileStatus := make(map[string]string)

	for _, line := range lines {
		if line != "" {
			// The git status --porcelain format has status followed by space and filename
			// Examples: "M file", "?? file", "MM file", " A file"
			line = strings.TrimSpace(line)
			if len(line) > 2 { // At least "X " + filename
				// Extract status (first 2 characters)
				status := line[:2]
				// Find the first space after the status and extract filename
				spaceIndex := strings.Index(line, " ")
				if spaceIndex != -1 && spaceIndex < len(line)-1 {
					// Extract filename after the first space
					filename := strings.TrimSpace(line[spaceIndex+1:])
					if filename != "" {
						files = append(files, filename)
						fileStatus[filename] = status
					}
				} else if len(line) > 2 {
					// Fallback: if no space found, take everything after the first 3 characters
					filename := strings.TrimSpace(line[2:])
					if filename != "" {
						files = append(files, filename)
						fileStatus[filename] = status
					}
				}
			}
		}
	}

	return files, fileStatus, nil
}

// GetDiffWithUntracked generates a diff that includes both tracked and untracked files
func (g *GitService) GetDiffWithUntracked() (string, error) {
	var diffParts []string

	// Get diff for tracked files
	diffCmd := exec.Command("git", "diff", "--diff-algorithm=minimal")
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get tracked files diff: %v", err)
	}
	trackedDiff := strings.TrimSpace(string(diffOutput))
	if trackedDiff != "" {
		diffParts = append(diffParts, trackedDiff)
	}

	// Get untracked files and generate diff for each
	files, fileStatus, err := g.GetAllChangesWithStatus()
	if err != nil {
		return "", fmt.Errorf("failed to get file status: %v", err)
	}

	for _, file := range files {
		status := fileStatus[file]
		// "??" means untracked file
		if status == "??" {
			// Check if file exists and is readable
			if _, err := os.Stat(file); os.IsNotExist(err) {
				continue
			}

			// Generate diff for untracked file using git diff --no-index
			// This shows the file as a new file (all lines added)
			diffCmd := exec.Command("git", "diff", "--no-index", "--no-color", "/dev/null", file)
			diffOutput, err := diffCmd.Output()
			if err != nil {
				// If git diff fails (e.g., binary file), try to read and format as new file
				content, readErr := os.ReadFile(file)
				if readErr != nil {
					continue // Skip files we can't read
				}
				// Format as a simple diff showing new file
				lines := strings.Split(string(content), "\n")
				var diffLines []string
				diffLines = append(diffLines, fmt.Sprintf("diff --git a/dev/null b/%s", file))
				diffLines = append(diffLines, "new file mode 100644")
				diffLines = append(diffLines, "index 0000000..0000000")
				diffLines = append(diffLines, "--- /dev/null")
				diffLines = append(diffLines, fmt.Sprintf("+++ b/%s", file))
				diffLines = append(diffLines, fmt.Sprintf("@@ -0,0 +1,%d @@", len(lines)))
				for _, line := range lines {
					diffLines = append(diffLines, "+"+line)
				}
				untrackedDiff := strings.Join(diffLines, "\n")
				diffParts = append(diffParts, untrackedDiff)
			} else {
				untrackedDiff := strings.TrimSpace(string(diffOutput))
				if untrackedDiff != "" {
					diffParts = append(diffParts, untrackedDiff)
				}
			}
		}
	}

	return strings.Join(diffParts, "\n\n"), nil
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

func (g *GitService) PushChanges(quiet *bool) error {
	cmd := exec.Command("git", "push")
	if !*quiet {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push changes. %v", err)
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

func (g *GitService) DetectIssueFromBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %v", err)
	}

	branchName := strings.TrimSpace(string(output))

	// Common patterns for issue detection in branch names
	patterns := []string{
		`(?i)([A-Z]+-\d+)`, // GEN-123, ELI-1220 (case-insensitive)
		`#(\d+)`,           // #123
		`(\d+)-`,           // 123-feature
		`-(\d+)-`,          // feature-123-description
		`issue-(\d+)`,      // issue-123
		`fix-(\d+)`,        // fix-123
		`feat-(\d+)`,       // feat-123
		`bug-(\d+)`,        // bug-123
	}

	for i, pattern := range patterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(branchName); len(matches) > 1 {
			result := matches[1]
			// Convert to uppercase for the first pattern (issue identifiers like GEN-123)
			if i == 0 {
				result = strings.ToUpper(result)
			}
			return result, nil
		}
	}

	return "", nil
}

func (g *GitService) CommitChangesWithOptions(message string, quiet *bool, noVerify *bool) error {
	args := []string{"commit", "-m", message}
	if *noVerify {
		args = append(args, "--no-verify")
	}

	cmd := exec.Command("git", args...)
	if !*quiet {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes. %v", err)
	}

	return nil
}

// DetectAndPrepareChanges handles staging, file detection, and preparation
func (g *GitService) DetectAndPrepareChanges(opts *CommitOptions) (*PreCommitData, error) {
	if *opts.StageAll {
		if err := g.StageAll(); err != nil {
			return nil, err
		}
	}

	filesChan := make(chan []string, 1)
	diffChan := make(chan string, 1)

	if err := spinner.New().
		Title("Detecting changes").
		Action(func() {
			var files []string
			var diff string
			var err error

			// If auto-select is enabled, get all changes (not just staged)
			if *opts.AutoSelect {
				// Get all changes in working directory (not just staged)
				allChanges, err := g.GetAllChanges()
				if err != nil {
					filesChan <- []string{}
					diffChan <- ""
					return
				}

				// Get full diff of all changes including untracked files
				diff, err = g.GetDiffWithUntracked()
				if err != nil {
					filesChan <- []string{}
					diffChan <- ""
					return
				}

				files = allChanges
			} else {
				// For normal flow, get only staged changes
				files, diff, err = g.DetectDiffChanges()
				if err != nil {
					filesChan <- []string{}
					diffChan <- ""
					return
				}
			}

			filesChan <- files
			diffChan <- diff
		}).
		Run(); err != nil {
		return nil, err
	}

	files, diff := <-filesChan, <-diffChan

	if len(files) == 0 {
		if *opts.AutoSelect {
			return nil, fmt.Errorf(
				"no changes found in working directory",
			)
		} else {
			return nil, fmt.Errorf(
				"no staged changes found. stage your changes manually, or automatically stage all changes with the `--all` flag, or use the `--auto` flag to let AI select changes",
			)
		}
	}

	relatedFiles := g.getRelatedFiles(files)

	// Auto-detect issue number from branch name if not provided
	issue := *opts.Issue
	if issue == "" {
		detectedIssue, err := g.DetectIssueFromBranch()
		if err == nil && detectedIssue != "" {
			issue = detectedIssue
			if !*opts.Quiet {
				color.New(color.FgCyan).Printf("Auto-detected issue: %s\n", detectedIssue)
			}
		}
	}

	return &PreCommitData{
		Files:        files,
		Diff:         diff,
		RelatedFiles: relatedFiles,
		Issue:        issue,
	}, nil
}

// getRelatedFiles discovers related files in the same directories
func (g *GitService) getRelatedFiles(files []string) map[string]string {
	relatedFiles := make(map[string]string)
	visitedDirs := make(map[string]bool)

	for _, file := range files {
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

	return relatedFiles
}

// ResetStaged resets the staged area, unstaging all files
func (g *GitService) ResetStaged() error {
	cmd := exec.Command("git", "reset")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reset staged files: %v", err)
	}
	return nil
}

// StageFiles stages specific files for commit
func (g *GitService) StageFiles(files []string) error {
	if len(files) == 0 {
		return nil
	}

	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage files %v: %v", files, err)
	}
	return nil
}

// ConfirmAction performs the actual commit and optional push
func (g *GitService) ConfirmAction(message string, quiet *bool, push *bool, dryRun *bool, noVerify *bool) error {
	if *dryRun {
		if !*quiet {
			color.New(color.FgYellow).Println("🔍 DRY RUN - No changes will be made")
			color.New(color.FgCyan).Printf("Would commit with message: %s\n", message)
			if *push {
				color.New(color.FgCyan).Println("Would push changes to remote repository")
			}
		}
		return nil
	}

	if err := g.CommitChangesWithOptions(message, quiet, noVerify); err != nil {
		return err
	}

	if !*quiet {
		color.New(color.FgGreen).Println("✔ Successfully committed!")
	}

	if *push {
		if err := g.PushChanges(quiet); err != nil {
			return err
		}

		if !*quiet {
			color.New(color.FgGreen).Println("✔ Successfully pushed!")
		}
	}

	return nil
}

func (g *GitService) GetDiff() (*PreCommitData, error) {
	// Get all remotes
	remotesOutput, err := exec.Command("git", "remote").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get remotes: %v", err)
	}
	remotes := strings.Fields(string(remotesOutput))
	if len(remotes) == 0 {
		return nil, fmt.Errorf("no git remotes configured")
	}

	// Prefer 'origin' if it exists
	remoteName := remotes[0]
	for _, r := range remotes {
		if r == "origin" {
			remoteName = r
			break
		}
	}

	// Fetch the remote to ensure it's up-to-date
	if err := exec.Command("git", "fetch", remoteName).Run(); err != nil {
		return nil, fmt.Errorf("failed to fetch remote '%s': %v", remoteName, err)
	}

	// Get remote details to find the HEAD branch
	defaultBranchOutput, err := exec.Command(
		"git",
		"remote",
		"show",
		remoteName,
	).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get details for remote '%s': %v", remoteName, err)
	}

	// Extract the HEAD branch name (e.g., 'main' or 'master')
	headBranchMatch := regexp.MustCompile(`HEAD branch: (.*)`).
		FindStringSubmatch(string(defaultBranchOutput))
	if len(headBranchMatch) < 2 {
		return nil, fmt.Errorf("could not determine HEAD branch for remote '%s'", remoteName)
	}
	headBranchName := headBranchMatch[1]

	// Diff against the remote's HEAD branch
	diff, err := exec.Command(
		"git",
		"diff",
		fmt.Sprintf("%s/%s", remoteName, headBranchName),
	).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get diff against '%s/%s': %v", remoteName, headBranchName, err)
	}

	return &PreCommitData{
		Diff:         string(diff),
		Files:        []string{},
		RelatedFiles: map[string]string{},
		Issue:        "",
	}, nil
}

func (g *GitService) CreatePullRequest(
	message string,
	quiet *bool,
	dryRun *bool,
	draft *bool,
) error {
	title, body, _ := strings.Cut(message, "\n")

	if *dryRun {
		if !*quiet {
			color.New(color.FgYellow).Println("🔍 DRY RUN - No changes will be made")
			color.New(color.FgCyan).
				Printf("Would create a pull request with title: %s\n", title)
		}
		return nil
	}

	// Get the current branch name
	branchName, err := g.GetCurrentBranchName()
	if err != nil {
		return err
	}

	// Get the remote name
	remoteName, err := g.GetRemoteName()
	if err != nil {
		return err
	}

	// Push the current branch to the remote
	if err := g.PushBranch(remoteName, branchName, quiet); err != nil {
		return err
	}

	args := []string{"pr", "create", "--title", title, "--body", body}
	if *draft {
		args = append(args, "--draft")
	}

	cmd := exec.Command("gh", args...)
	if !*quiet {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create pull request: %v", err)
	}

	if !*quiet {
		color.New(color.FgGreen).Println("✔ Successfully created a pull request!")
	}

	return nil
}

func (g *GitService) GetCurrentBranchName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch name: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (g *GitService) GetRemoteName() (string, error) {
	remotesOutput, err := exec.Command("git", "remote").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remotes: %v", err)
	}
	remotes := strings.Fields(string(remotesOutput))
	if len(remotes) == 0 {
		return "", fmt.Errorf("no git remotes configured")
	}

	// Prefer 'origin' if it exists
	remoteName := remotes[0]
	for _, r := range remotes {
		if r == "origin" {
			remoteName = r
			break
		}
	}
	return remoteName, nil
}

func (g *GitService) PushBranch(remoteName, branchName string, quiet *bool) error {
	cmd := exec.Command("git", "push", "-u", remoteName, branchName)
	if !*quiet {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push branch '%s' to remote '%s': %v", branchName, remoteName, err)
	}
	return nil
}
