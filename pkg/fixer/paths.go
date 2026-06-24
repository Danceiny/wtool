package fixer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Danceiny/wtool/pkg/detector"
)

// PathFixer handles path repairs
type PathFixer struct {
	WorktreePath string
}

// NewPathFixer creates a new path fixer
func NewPathFixer(worktreePath string) *PathFixer {
	return &PathFixer{WorktreePath: worktreePath}
}

// Fix repairs all detected path issues
func (f *PathFixer) Fix() error {
	result, err := detector.DetectPathIssues(f.WorktreePath)
	if err != nil {
		return err
	}

	if !result.HasIssues {
		return nil
	}

	for _, fix := range result.Fixes {
		if err := f.applyFix(fix); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fix paths in %s: %v\n", fix.File, err)
		}
	}

	return nil
}

func (f *PathFixer) applyFix(fix detector.PathFix) error {
	switch fix.Strategy {
	case "strip_cwd_prefix":
		return f.stripCwdPrefix(fix.File)
	case "rewrite_paths":
		return f.rewritePaths(fix.File)
	case "rewrite_absolute_to_relative":
		return f.stripCwdPrefix(fix.File)
	default:
		return fmt.Errorf("unknown path fix strategy: %s", fix.Strategy)
	}
}

func (f *PathFixer) stripCwdPrefix(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)
	// Replace worktree path prefix with empty string (making relative)
	content = strings.ReplaceAll(content, f.WorktreePath+"/", "")
	content = strings.ReplaceAll(content, f.WorktreePath, "")

	return os.WriteFile(filePath, []byte(content), 0644)
}

func (f *PathFixer) rewritePaths(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)
	content = strings.ReplaceAll(content, f.WorktreePath, "${PROJECT_ROOT}")
	content = strings.ReplaceAll(content, strings.ReplaceAll(f.WorktreePath, "/", "\\"), "${PROJECT_ROOT}")

	return os.WriteFile(filePath, []byte(content), 0644)
}

// FixGraphJSON specifically fixes graph.json files
func (f *PathFixer) FixGraphJSON(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var graph map[string]interface{}
	if err := json.Unmarshal(data, &graph); err != nil {
		return err
	}

	f.fixPathsInValue(graph, f.WorktreePath)

	out, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, out, 0644)
}

func (f *PathFixer) fixPathsInValue(v interface{}, cwd string) {
	switch val := v.(type) {
	case string:
		// Handled at parent level
	case map[string]interface{}:
		for k, v := range val {
			if str, ok := v.(string); ok {
				if strings.HasPrefix(str, cwd) {
					val[k] = strings.TrimPrefix(str, cwd+"/")
				}
			} else {
				f.fixPathsInValue(v, cwd)
			}
		}
	case []interface{}:
		for _, v := range val {
			f.fixPathsInValue(v, cwd)
		}
	}
}
