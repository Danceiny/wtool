package detector

import (
	"path/filepath"

	"github.com/Danceiny/wtool/pkg/utils"
)

// AIToolResult holds detection results for a single AI tool
type AIToolResult struct {
	Name        string
	Detected    bool
	ConfigPath  string
	ConfigType  string
	HasHooks    bool
	HookPaths   []string
}

// AIToolsResult holds detection results for all AI tools
type AIToolsResult struct {
	Tools      []AIToolResult
	AgentsMd   bool
	AnyDetected bool
}

// DetectAITools scans for AI tool configurations
func DetectAITools(projectRoot string) *AIToolsResult {
	result := &AIToolsResult{
		Tools: make([]AIToolResult, 0),
	}

	tools := []struct {
		name       string
		configPath string
		configType string
		hookPaths  []string
	}{
		{
			name:       "claude",
			configPath: ".claude/settings.json",
			configType: "json",
			hookPaths:  []string{".claude/helpers/hook-handler.cjs"},
		},
		{
			name:       "cursor",
			configPath: ".cursor/worktrees.json",
			configType: "json",
			hookPaths:  []string{},
		},
		{
			name:       "codex",
			configPath: ".codex/config.toml",
			configType: "toml",
			hookPaths:  []string{".codex/agents"},
		},
		{
			name:       "qoder",
			configPath: ".qoder/settings.local.json",
			configType: "json",
			hookPaths:  []string{},
		},
		{
			name:       "trae",
			configPath: ".trae",
			configType: "dir",
			hookPaths:  []string{},
		},
	}

	for _, tool := range tools {
		fullPath := filepath.Join(projectRoot, tool.configPath)
		detected := utils.FileExists(fullPath) || utils.DirExists(fullPath)
		
		var hooks []string
		for _, hookPath := range tool.hookPaths {
			fullHookPath := filepath.Join(projectRoot, hookPath)
			if utils.FileExists(fullHookPath) || utils.DirExists(fullHookPath) {
				hooks = append(hooks, hookPath)
			}
		}

		if detected {
			result.AnyDetected = true
			result.Tools = append(result.Tools, AIToolResult{
				Name:       tool.name,
				Detected:   true,
				ConfigPath: tool.configPath,
				ConfigType: tool.configType,
				HasHooks:   len(hooks) > 0,
				HookPaths:  hooks,
			})
		}
	}

	result.AgentsMd = utils.FileExists(filepath.Join(projectRoot, "AGENTS.md"))
	return result
}
