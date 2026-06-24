package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Danceiny/wtool/pkg/utils"
)

// PathFix represents a detected path fix needed
type PathFix struct {
	Type        string
	File        string
	Strategy    string
	Description string
	HasAbsolute bool
}

// PathFixResult holds all path fixes needed
type PathFixResult struct {
	Fixes      []PathFix
	HasIssues  bool
}

// DetectPathIssues scans for absolute path issues in worktree
func DetectPathIssues(projectRoot string) (*PathFixResult, error) {
	result := &PathFixResult{
		Fixes: make([]PathFix, 0),
	}

	// Check codegraph
	codegraphDb := filepath.Join(projectRoot, ".codegraph", "codegraph.db")
	if utils.FileExists(codegraphDb) {
		result.Fixes = append(result.Fixes, PathFix{
			Type:        "codegraph",
			File:        codegraphDb,
			Strategy:    "rewrite_absolute_to_relative",
			Description: "CodeGraph SQLite DB stores absolute paths",
			HasAbsolute: true,
		})
		result.HasIssues = true
	}

	// Check graph.json
	graphJson := filepath.Join(projectRoot, "graph.json")
	if utils.FileExists(graphJson) {
		hasAbs, err := checkAbsolutePathsInJSON(graphJson, projectRoot)
		if err == nil && hasAbs {
			result.Fixes = append(result.Fixes, PathFix{
				Type:        "graphify",
				File:        graphJson,
				Strategy:    "strip_cwd_prefix",
				Description: "Graphify stores absolute source_file paths",
				HasAbsolute: true,
			})
			result.HasIssues = true
		}
	}

	// Check config files for absolute paths
	configCandidates := []string{
		".cursor/settings.json",
		".claude/settings.json",
		".vscode/settings.json",
		"tsconfig.json",
	}

	for _, candidate := range configCandidates {
		path := filepath.Join(projectRoot, candidate)
		if !utils.FileExists(path) {
			continue
		}
		
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		
		if containsAbsolutePaths(string(content), projectRoot) {
			result.Fixes = append(result.Fixes, PathFix{
				Type:        "config",
				File:        path,
				Strategy:    "rewrite_paths",
				Description: fmt.Sprintf("Config file contains absolute paths: %s", candidate),
				HasAbsolute: true,
			})
			result.HasIssues = true
		}
	}

	return result, nil
}

func checkAbsolutePathsInJSON(filePath, projectRoot string) (bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}
	
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return false, err
	}
	
	return containsAbsolutePathsInValue(obj, projectRoot), nil
}

func containsAbsolutePathsInValue(v interface{}, projectRoot string) bool {
	switch val := v.(type) {
	case string:
		return strings.HasPrefix(val, projectRoot) || strings.HasPrefix(val, "/home/") || strings.HasPrefix(val, "/Users/")
	case map[string]interface{}:
		for _, v := range val {
			if containsAbsolutePathsInValue(v, projectRoot) {
				return true
			}
		}
	case []interface{}:
		for _, v := range val {
			if containsAbsolutePathsInValue(v, projectRoot) {
				return true
			}
		}
	}
	return false
}

func containsAbsolutePaths(content, projectRoot string) bool {
	// Check for common absolute path patterns
	patterns := []string{
		projectRoot,
		"/home/",
		"/Users/",
		"C:\\\\",
	}
	
	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	return false
}
