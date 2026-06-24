package utils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitWorktree represents a git worktree
type GitWorktree struct {
	Path       string
	Commit     string
	Branch     string
	IsMain     bool
	IsBare     bool
}

// GetWorktreeInfo returns information about the current git worktree
func GetWorktreeInfo(dir string) (*GitWorktree, error) {
	// Check if it's a git repo
	if err := exec.Command("git", "-C", dir, "rev-parse", "--git-dir").Run(); err != nil {
		return nil, fmt.Errorf("not a git repository: %s", dir)
	}

	// Get worktree path
	path, err := execGit(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	path = strings.TrimSpace(path)

	// Get current commit
	commit, err := execGit(dir, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	commit = strings.TrimSpace(commit)

	// Get current branch
	branch, _ := execGit(dir, "symbolic-ref", "--short", "HEAD")
	branch = strings.TrimSpace(branch)

	// Check if main worktree
	worktrees, err := ListWorktrees(dir)
	if err != nil {
		return nil, err
	}

	isMain := false
	if len(worktrees) > 0 {
		mainPath, err := filepath.EvalSymlinks(worktrees[0].Path)
		if err != nil {
			mainPath = worktrees[0].Path
		}
		currentPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			currentPath = path
		}
		isMain = mainPath == currentPath
	}

	return &GitWorktree{
		Path:   path,
		Commit: commit,
		Branch: branch,
		IsMain: isMain,
	}, nil
}

// ListWorktrees returns all worktrees for a repo
func ListWorktrees(dir string) ([]GitWorktree, error) {
	out, err := execGit(dir, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	var worktrees []GitWorktree
	var current GitWorktree
	lines := strings.Split(out, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "worktree ") {
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = GitWorktree{Path: strings.TrimPrefix(line, "worktree ")}
		} else if strings.HasPrefix(line, "HEAD ") {
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		} else if line == "bare" {
			current.IsBare = true
		}
	}
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

// GetWorktreeIndex returns the index of a worktree in the list
func GetWorktreeIndex(dir string) (int, error) {
	worktrees, err := ListWorktrees(dir)
	if err != nil {
		return 0, err
	}

	path, err := execGit(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return 0, err
	}
	path = strings.TrimSpace(path)
	
	currentPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		currentPath = path
	}

	for i, wt := range worktrees {
		wtPath, err := filepath.EvalSymlinks(wt.Path)
		if err != nil {
			wtPath = wt.Path
		}
		if wtPath == currentPath {
			return i, nil
		}
	}
	return 0, nil
}

// GetPrimaryWorktree returns the main worktree path
func GetPrimaryWorktree(dir string) (string, error) {
	worktrees, err := ListWorktrees(dir)
	if err != nil {
		return "", err
	}
	if len(worktrees) == 0 {
		return "", fmt.Errorf("no worktrees found")
	}
	return worktrees[0].Path, nil
}

// execGit runs a git command and returns stdout
func execGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %v failed: %w (output: %s)", args, err, string(out))
	}
	return string(out), nil
}
