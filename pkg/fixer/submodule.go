package fixer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Danceiny/wtool/pkg/detector"
	"github.com/Danceiny/wtool/pkg/utils"
)

// SubmoduleFixer handles submodule initialization
type SubmoduleFixer struct {
	WorktreePath string
	SourceRoot   string
	Strategy     string
}

// NewSubmoduleFixer creates a new submodule fixer
func NewSubmoduleFixer(worktreePath, sourceRoot, strategy string) *SubmoduleFixer {
	if strategy == "" {
		strategy = "local-clone-first"
	}
	return &SubmoduleFixer{
		WorktreePath: worktreePath,
		SourceRoot:   sourceRoot,
		Strategy:     strategy,
	}
}

// Fix initializes all submodules
func (f *SubmoduleFixer) Fix() error {
	result, err := detector.DetectSubmodules(f.WorktreePath)
	if err != nil {
		return err
	}

	if !result.HasSubmodules {
		return nil
	}

	if !result.NeedInit {
		return nil
	}

	for _, mod := range result.Modules {
		if mod.IsInitialized {
			continue
		}

		if err := f.initSubmodule(mod); err != nil {
			// Log warning but continue with other submodules
			fmt.Fprintf(os.Stderr, "Warning: failed to initialize submodule %s: %v\n", mod.Path, err)
		}
	}

	return nil
}

func (f *SubmoduleFixer) initSubmodule(mod detector.SubmoduleInfo) error {
	targetDir := filepath.Join(f.WorktreePath, mod.Path)

	// Check if directory exists but is non-empty and not a submodule
	if utils.DirExists(targetDir) && !utils.FileExists(filepath.Join(targetDir, ".git")) {
		entries, err := os.ReadDir(targetDir)
		if err == nil && len(entries) > 0 {
			return fmt.Errorf("directory %s is non-empty but not a git submodule", mod.Path)
		}
	}

	// Try local clone first
	if f.Strategy == "local-clone-first" && f.SourceRoot != "" && f.SourceRoot != f.WorktreePath {
		if err := f.checkoutFromSource(mod); err == nil {
			return nil
		}
	}

	// Fallback to git submodule update
	_, err := utils.ExecGit(f.WorktreePath, "submodule", "update", "--init", "--recursive", "--", mod.Path)
	if err != nil {
		return fmt.Errorf("git submodule update failed: %w", err)
	}

	return nil
}

func (f *SubmoduleFixer) checkoutFromSource(mod detector.SubmoduleInfo) error {
	sourceDir := filepath.Join(f.SourceRoot, mod.Path)
	targetDir := filepath.Join(f.WorktreePath, mod.Path)

	// Validate source directory
	if !utils.DirExists(sourceDir) {
		return fmt.Errorf("source directory does not exist: %s", sourceDir)
	}

	// Check if source is a valid git repo
	if _, err := utils.ExecGit(sourceDir, "rev-parse", "--is-inside-work-tree"); err != nil {
		return fmt.Errorf("source is not a git repo: %s", sourceDir)
	}

	// Check if expected commit exists in source
	if mod.ExpectedCommit != "" {
		if _, err := utils.ExecGit(sourceDir, "cat-file", "-e", mod.ExpectedCommit+"^{commit}"); err != nil {
			return fmt.Errorf("expected commit not found in source: %s", mod.ExpectedCommit)
		}
	}

	// Check if target is empty
	if utils.DirExists(targetDir) {
		entries, err := os.ReadDir(targetDir)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return fmt.Errorf("target directory is not empty: %s", targetDir)
		}
	}

	// Remove target if it exists (and is empty)
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("failed to remove target: %w", err)
	}

	// Clone from source
	if _, err := utils.ExecCommand("", "git", "clone", "--quiet", "--no-checkout", sourceDir, targetDir); err != nil {
		return fmt.Errorf("failed to clone from source: %w", err)
	}

	// Checkout expected commit
	if mod.ExpectedCommit != "" {
		if _, err := utils.ExecGit(targetDir, "checkout", "--quiet", "--force", mod.ExpectedCommit); err != nil {
			return fmt.Errorf("failed to checkout commit: %w", err)
		}
	}

	return nil
}
