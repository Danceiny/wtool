package hooks

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Danceiny/wtool/pkg/utils"
)

const postCheckoutHook = `#!/bin/sh
# wtool post-checkout hook
# Auto-initialize worktree after checkout

if command -v wtool >/dev/null 2>&1; then
    wtool init --quiet
fi
`

const postMergeHook = `#!/bin/sh
# wtool post-merge hook
# Auto-sync worktree after merge

if command -v wtool >/dev/null 2>&1; then
    wtool sync --quiet
fi
`

// InstallHooks installs git hooks to the specified directory
func InstallHooks(hooksDir string) error {
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	hooks := map[string]string{
		"post-checkout": postCheckoutHook,
		"post-merge":    postMergeHook,
	}

	for name, content := range hooks {
		hookPath := filepath.Join(hooksDir, name)
		
		// Backup existing hook
		if utils.FileExists(hookPath) {
			backupPath := hookPath + ".wtool-backup"
			if !utils.FileExists(backupPath) {
				data, err := os.ReadFile(hookPath)
				if err != nil {
					return err
				}
				if err := os.WriteFile(backupPath, data, 0755); err != nil {
					return err
				}
			}
			
			// Append to existing hook if it doesn't already contain wtool
			existing, err := os.ReadFile(hookPath)
			if err != nil {
				return err
			}
			if !utils.Contains(string(existing), "wtool") {
				content = string(existing) + "\n" + content
			} else {
				continue // Already has wtool
			}
		}

		if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
			return fmt.Errorf("failed to write hook %s: %w", name, err)
		}
	}

	return nil
}

// GetGitDir returns the git directory for a repo
func GetGitDir(repoPath string) (string, error) {
	out, err := utils.ExecGit(repoPath, "rev-parse", "--git-dir")
	if err != nil {
		return "", err
	}
	
	gitDir := out
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(repoPath, gitDir)
	}
	
	return gitDir, nil
}
