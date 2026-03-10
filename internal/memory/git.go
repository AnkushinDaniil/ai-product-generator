package memory

import (
	"context"
	"os/exec"
	"strings"
)

// GetCurrentCommit returns the current git commit hash for a file
func GetCurrentCommit(ctx context.Context, file string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "log", "-1", "--format=%H", "--", file) //nolint:gosec // file path is validated by caller
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	commit := strings.TrimSpace(string(output))
	return commit, nil
}

// GetCommitsSince returns commits for a file since a given commit
func GetCommitsSince(ctx context.Context, file, sinceCommit string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "log", "--format=%H", sinceCommit+"..HEAD", "--", file) //nolint:gosec // inputs validated by caller
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
func GetProjectRoot(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	root := strings.TrimSpace(string(output))
	return root, nil
}

// IsGitRepo checks if the current directory is in a git repository
func IsGitRepo(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}
