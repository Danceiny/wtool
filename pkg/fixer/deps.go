package fixer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Danceiny/wtool/pkg/utils"
)

// DepsFixer handles dependency linking/installation
type DepsFixer struct {
	WorktreePath string
	SourceRoot   string
	Strategy     string
}

// NewDepsFixer creates a new dependency fixer
func NewDepsFixer(worktreePath, sourceRoot, strategy string) *DepsFixer {
	if strategy == "" {
		strategy = "link-first"
	}
	return &DepsFixer{
		WorktreePath: worktreePath,
		SourceRoot:   sourceRoot,
		Strategy:     strategy,
	}
}

// Fix handles dependency setup
func (f *DepsFixer) Fix() error {
	// Fix main project
	if err := f.fixDir(f.WorktreePath); err != nil {
		return err
	}

	// Check submodules for package.json
	gitmodulesPath := filepath.Join(f.WorktreePath, ".gitmodules")
	if !utils.FileExists(gitmodulesPath) {
		return nil
	}

	// Parse submodules and check each
	// This is a simplified version - in production would parse .gitmodules properly
	return nil
}

func (f *DepsFixer) fixDir(dir string) error {
	if !utils.FileExists(filepath.Join(dir, "package.json")) {
		return nil
	}

	if utils.DirExists(filepath.Join(dir, "node_modules")) {
		return nil
	}

	// Try linking from source
	if f.Strategy == "link-first" && f.SourceRoot != "" && f.SourceRoot != f.WorktreePath {
		rel, err := filepath.Rel(f.WorktreePath, dir)
		if err != nil {
			rel = "."
		}
		sourceDir := filepath.Join(f.SourceRoot, rel)
		
		if f.canLinkNodeModules(dir, sourceDir) {
			return f.linkNodeModules(dir, sourceDir)
		}
	}

	// Try installing
	if f.Strategy == "install" || f.Strategy == "link-first" {
		return f.installNodeModules(dir)
	}

	return nil
}

func (f *DepsFixer) canLinkNodeModules(targetDir, sourceDir string) bool {
	if !utils.DirExists(filepath.Join(sourceDir, "node_modules")) {
		return false
	}

	// Check manifest consistency
	manifests := []string{"package.json", "pnpm-lock.yaml", "package-lock.json", "yarn.lock"}
	for _, manifest := range manifests {
		targetFile := filepath.Join(targetDir, manifest)
		sourceFile := filepath.Join(sourceDir, manifest)
		
		if utils.FileExists(targetFile) || utils.FileExists(sourceFile) {
			if !utils.FileExists(targetFile) || !utils.FileExists(sourceFile) {
				return false
			}
			
			// Compare files
			targetData, err1 := os.ReadFile(targetFile)
			sourceData, err2 := os.ReadFile(sourceFile)
			if err1 != nil || err2 != nil {
				return false
			}
			if string(targetData) != string(sourceData) {
				return false
			}
		}
	}

	return true
}

func (f *DepsFixer) linkNodeModules(targetDir, sourceDir string) error {
	nodeModulesSource := filepath.Join(sourceDir, "node_modules")
	nodeModulesTarget := filepath.Join(targetDir, "node_modules")

	if err := os.Symlink(nodeModulesSource, nodeModulesTarget); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

func (f *DepsFixer) installNodeModules(dir string) error {
	var cmd string
	var args []string

	if utils.FileExists(filepath.Join(dir, "pnpm-lock.yaml")) {
		cmd = "pnpm"
		args = []string{"install", "--frozen-lockfile"}
	} else if utils.FileExists(filepath.Join(dir, "yarn.lock")) {
		cmd = "yarn"
		args = []string{"install", "--frozen-lockfile"}
	} else if utils.FileExists(filepath.Join(dir, "package-lock.json")) {
		cmd = "npm"
		args = []string{"ci"}
	} else {
		cmd = "npm"
		args = []string{"install"}
	}

	if !utils.CommandExists(cmd) {
		return fmt.Errorf("command not found: %s", cmd)
	}

	_, err := utils.ExecCommand(dir, cmd, args...)
	return err
}
