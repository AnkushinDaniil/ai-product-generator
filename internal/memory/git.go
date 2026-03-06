package memory

import (
	"os/exec"
	"strings"
)

// GetCurrentCommit returns the current git commit hash for a file
func GetCurrentCommit(file string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%H", "--", file)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	commit := strings.TrimSpace(string(output))
	return commit, nil
}

// GetCommitsSince returns commits for a file since a given commit
func GetCommitsSince(file, sinceCommit string) ([]string, error) {
	cmd := exec.Command("git", "log", "--format=%H", sinceCommit+"..HEAD", "--", file)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	commitText := strings.TrimSpace(string(output))
	if commitText == "" {
		return []string{}, nil
	}

	commits := strings.Split(commitText, "\n")
	return commits, nil
}

// GetProjectRoot returns the git repository root directory
func GetProjectRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	root := strings.TrimSpace(string(output))
	return root, nil
}

// IsGitRepo checks if the current directory is in a git repository
func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}
