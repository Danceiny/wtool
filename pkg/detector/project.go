package detector

import (
	"os"
	"path/filepath"

	"github.com/Danceiny/wtool/pkg/utils"
)

// ProjectType represents the detected project type
type ProjectType string

const (
	TypeMonorepo ProjectType = "monorepo"
	TypeSingle   ProjectType = "single"
	TypeLibrary  ProjectType = "library"
	TypeUnknown  ProjectType = "unknown"
)

// ProjectResult holds project detection results
type ProjectResult struct {
	Name         string
	Type         ProjectType
	Root         string
	HasGoMod     bool
	HasPackageJSON bool
	HasPnpmWorkspace bool
	HasGitmodules  bool
	Languages    []string
}

// DetectProject detects project type and structure
func DetectProject(projectRoot string) (*ProjectResult, error) {
	result := &ProjectResult{
		Root:      projectRoot,
		Type:      TypeUnknown,
		Languages: make([]string, 0),
	}

	// Check for Go
	if utils.FileExists(filepath.Join(projectRoot, "go.mod")) {
		result.HasGoMod = true
		result.Languages = append(result.Languages, "go")
	}

	// Check for Node.js
	if utils.FileExists(filepath.Join(projectRoot, "package.json")) {
		result.HasPackageJSON = true
		result.Languages = append(result.Languages, "node")
	}

	// Check for pnpm workspace
	if utils.FileExists(filepath.Join(projectRoot, "pnpm-workspace.yaml")) {
		result.HasPnpmWorkspace = true
	}

	// Check for submodules
	if utils.FileExists(filepath.Join(projectRoot, ".gitmodules")) {
		result.HasGitmodules = true
	}

	// Determine project type
	if result.HasPnpmWorkspace || result.HasGitmodules {
		result.Type = TypeMonorepo
	} else if result.HasGoMod || result.HasPackageJSON {
		result.Type = TypeSingle
	}

	// Try to get project name from go.mod or package.json
	if result.HasGoMod {
		result.Name = filepath.Base(projectRoot)
	} else if result.HasPackageJSON {
		result.Name = filepath.Base(projectRoot)
	}

	// Check for other languages
	if utils.FileExists(filepath.Join(projectRoot, "Cargo.toml")) {
		result.Languages = append(result.Languages, "rust")
	}
	if utils.FileExists(filepath.Join(projectRoot, "pyproject.toml")) || utils.FileExists(filepath.Join(projectRoot, "requirements.txt")) {
		result.Languages = append(result.Languages, "python")
	}
	if utils.FileExists(filepath.Join(projectRoot, "pom.xml")) || utils.FileExists(filepath.Join(projectRoot, "build.gradle")) {
		result.Languages = append(result.Languages, "java")
	}

	return result, nil
}

// GetProjectName returns a project name from directory or config
func GetProjectName(projectRoot string) string {
	// Try to read from go.mod
	goModPath := filepath.Join(projectRoot, "go.mod")
	if utils.FileExists(goModPath) {
		data, err := os.ReadFile(goModPath)
		if err == nil {
			// Simple parsing: first line is "module <name>"
			lines := string(data)
			if len(lines) > 7 && lines[:7] == "module " {
				end := 0
				for i := 7; i < len(lines); i++ {
					if lines[i] == '\n' || lines[i] == '\r' {
						end = i
						break
					}
				}
				if end > 7 {
					return lines[7:end]
				}
			}
		}
	}
	
	return filepath.Base(projectRoot)
}
