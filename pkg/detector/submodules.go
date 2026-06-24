package detector

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Danceiny/wtool/pkg/utils"
)

// SubmoduleInfo holds information about a submodule
type SubmoduleInfo struct {
	Path           string
	URL            string
	Branch         string
	IsInitialized  bool
	IsEmpty        bool
	ExpectedCommit string
	GitDir         string
}

// SubmoduleResult holds detection results
type SubmoduleResult struct {
	HasSubmodules bool
	Modules       []SubmoduleInfo
	NeedInit      bool
}

// DetectSubmodules detects git submodules in a project
func DetectSubmodules(projectRoot string) (*SubmoduleResult, error) {
	gitmodulesPath := filepath.Join(projectRoot, ".gitmodules")
	if !utils.FileExists(gitmodulesPath) {
		return &SubmoduleResult{HasSubmodules: false}, nil
	}

	modules, err := parseGitmodules(gitmodulesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .gitmodules: %w", err)
	}

	if len(modules) == 0 {
		return &SubmoduleResult{HasSubmodules: true, Modules: []SubmoduleInfo{}}, nil
	}

	result := &SubmoduleResult{
		HasSubmodules: true,
		Modules:       make([]SubmoduleInfo, 0, len(modules)),
	}

	for _, mod := range modules {
		modPath := filepath.Join(projectRoot, mod.Path)
		mod.GitDir = filepath.Join(modPath, ".git")
		mod.IsInitialized = utils.FileExists(mod.GitDir)
		mod.IsEmpty = !utils.DirExists(modPath) || isDirEmpty(modPath)
		mod.ExpectedCommit = getSubmoduleCommit(projectRoot, mod.Path)

		if !mod.IsInitialized {
			result.NeedInit = true
		}

		result.Modules = append(result.Modules, mod)
	}

	return result, nil
}

func parseGitmodules(path string) ([]SubmoduleInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var modules []SubmoduleInfo
	var current *SubmoduleInfo

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[submodule") {
			if current != nil {
				modules = append(modules, *current)
			}
			current = &SubmoduleInfo{}
			continue
		}

		if current == nil {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "path":
			current.Path = value
		case "url":
			current.URL = value
		case "branch":
			current.Branch = value
		}
	}

	if current != nil {
		modules = append(modules, *current)
	}

	return modules, scanner.Err()
}

func getSubmoduleCommit(projectRoot, path string) string {
	out, err := utils.ExecGit(projectRoot, "ls-files", "--stage", "--", path)
	if err != nil {
		return ""
	}
	
	fields := strings.Fields(out)
	if len(fields) >= 2 && fields[0] == "160000" {
		return fields[1]
	}
	return ""
}

func isDirEmpty(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return true
	}
	return len(entries) == 0
}
