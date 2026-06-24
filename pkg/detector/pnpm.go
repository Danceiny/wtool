package detector

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Danceiny/wtool/pkg/utils"
	"gopkg.in/yaml.v3"
)

// PnpmResult holds pnpm workspace detection results
type PnpmResult struct {
	IsWorkspace       bool
	IsRoot            bool
	ManifestPath      string
	HasNodeModules    bool
	HasPnpmLock       bool
	PackageManager    string
	Workspaces        []string
	WorkspacePackages []string
}

// PackageJSON represents package.json structure
type PackageJSON struct {
	Name       string            `json:"name"`
	Workspaces *WorkspaceConfig    `json:"workspaces,omitempty"`
}

// WorkspaceConfig represents workspaces in package.json
type WorkspaceConfig struct {
	Packages []string `json:"packages,omitempty"`
}

// PnpmWorkspaceYAML represents pnpm-workspace.yaml structure
type PnpmWorkspaceYAML struct {
	Packages []string `yaml:"packages"`
}

// DetectPnpm detects pnpm workspace configuration
func DetectPnpm(projectRoot string) (*PnpmResult, error) {
	result := &PnpmResult{
		PackageManager: "none",
	}

	// Check for pnpm-workspace.yaml
	workspaceYaml := filepath.Join(projectRoot, "pnpm-workspace.yaml")
	if utils.FileExists(workspaceYaml) {
		result.IsWorkspace = true
		result.IsRoot = true
		result.ManifestPath = workspaceYaml
		result.PackageManager = "pnpm"

		data, err := os.ReadFile(workspaceYaml)
		if err == nil {
			var yamlCfg PnpmWorkspaceYAML
			if err := yaml.Unmarshal(data, &yamlCfg); err == nil {
				result.WorkspacePackages = yamlCfg.Packages
			}
		}
	}

	// Check for package.json workspaces
	packageJson := filepath.Join(projectRoot, "package.json")
	if utils.FileExists(packageJson) {
		data, err := os.ReadFile(packageJson)
		if err == nil {
			var pkg PackageJSON
			if err := json.Unmarshal(data, &pkg); err == nil {
				if pkg.Workspaces != nil && len(pkg.Workspaces.Packages) > 0 {
					result.IsWorkspace = true
					result.IsRoot = true
					result.WorkspacePackages = pkg.Workspaces.Packages
					if result.PackageManager == "none" {
						result.PackageManager = detectPackageManager(projectRoot)
					}
				}
			}
		}
	}

	// Check for lock files to determine package manager
	if result.PackageManager == "none" {
		result.PackageManager = detectPackageManager(projectRoot)
	}

	result.HasNodeModules = utils.DirExists(filepath.Join(projectRoot, "node_modules"))
	result.HasPnpmLock = utils.FileExists(filepath.Join(projectRoot, "pnpm-lock.yaml"))

	return result, nil
}

func detectPackageManager(projectRoot string) string {
	if utils.FileExists(filepath.Join(projectRoot, "pnpm-lock.yaml")) {
		return "pnpm"
	}
	if utils.FileExists(filepath.Join(projectRoot, "package-lock.json")) {
		return "npm"
	}
	if utils.FileExists(filepath.Join(projectRoot, "yarn.lock")) {
		return "yarn"
	}
	if utils.FileExists(filepath.Join(projectRoot, "bun.lockb")) {
		return "bun"
	}
	if utils.FileExists(filepath.Join(projectRoot, "package.json")) {
		return "npm" // default
	}
	return "none"
}

// DetectNodeModulesInSubdirs checks for node_modules in subdirectories
func DetectNodeModulesInSubdirs(projectRoot string, paths []string) map[string]bool {
	result := make(map[string]bool)
	for _, path := range paths {
		dir := filepath.Join(projectRoot, path)
		if utils.DirExists(dir) {
			result[path] = utils.DirExists(filepath.Join(dir, "node_modules"))
		}
	}
	return result
}
