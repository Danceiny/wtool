package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the wtool configuration
type Config struct {
	Version string       `mapstructure:"version"`
	Project ProjectConfig `mapstructure:"project"`
	Submodules SubmoduleConfig `mapstructure:"submodules"`
	Dependencies DepsConfig `mapstructure:"dependencies"`
	Ports PortConfig `mapstructure:"ports"`
	Env EnvConfig `mapstructure:"env"`
	PathFixes PathFixConfig `mapstructure:"path_fixes"`
	AITools AIToolsConfig `mapstructure:"ai_tools"`
	Sync SyncConfig `mapstructure:"sync"`
	Logging LoggingConfig `mapstructure:"logging"`
	Hooks HooksConfig `mapstructure:"hooks"`
}

// ProjectConfig holds project metadata
type ProjectConfig struct {
	Name string `mapstructure:"name"`
	Type string `mapstructure:"type"`
	Root string `mapstructure:"root"`
}

// SubmoduleConfig holds submodule settings
type SubmoduleConfig struct {
	Enabled       bool     `mapstructure:"enabled"`
	Strategy      string   `mapstructure:"strategy"`
	SourceRoot    string   `mapstructure:"source_root"`
	Parallel      bool     `mapstructure:"parallel"`
	RetryOnFailure bool    `mapstructure:"retry_on_failure"`
	SkipPaths     []string `mapstructure:"skip_paths"`
}

// DepsConfig holds dependency management settings
type DepsConfig struct {
	Node NodeDepsConfig `mapstructure:"node"`
	Go   GoDepsConfig   `mapstructure:"go"`
}

// NodeDepsConfig holds Node.js dependency settings
type NodeDepsConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	PackageManager   string `mapstructure:"package_manager"`
	LinkFromSource   bool   `mapstructure:"link_from_source"`
	InstallStrategy  string `mapstructure:"install_strategy"`
}

// GoDepsConfig holds Go dependency settings
type GoDepsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	VendorMode bool   `mapstructure:"vendor_mode"`
}

// PortConfig holds port allocation settings
type PortConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	BaseOffset  int           `mapstructure:"base_offset"`
	Services    []ServicePort `mapstructure:"services"`
}

// ServicePort defines a service port mapping
type ServicePort struct {
	Name    string `mapstructure:"name"`
	Base    int    `mapstructure:"base"`
	EnvKey  string `mapstructure:"env_key"`
}

// EnvConfig holds environment variable generation settings
type EnvConfig struct {
	Enabled    bool              `mapstructure:"enabled"`
	OutputFile string            `mapstructure:"output_file"`
	Format     string            `mapstructure:"format"`
	Variables  map[string]string `mapstructure:"variables"`
}

// PathFixConfig holds path repair settings
type PathFixConfig struct {
	Enabled    bool                `mapstructure:"enabled"`
	Strategies map[string]PathStrategy `mapstructure:"strategies"`
}

// PathStrategy defines a path fixing strategy
type PathStrategy struct {
	Enabled  bool   `mapstructure:"enabled"`
	Paths    []string `mapstructure:"paths"`
	Strategy string   `mapstructure:"strategy"`
}

// AIToolsConfig holds AI tool integration settings
type AIToolsConfig struct {
	Claude AIToolConfig `mapstructure:"claude"`
	Cursor AIToolConfig `mapstructure:"cursor"`
	Codex  AIToolConfig `mapstructure:"codex"`
	Qoder  AIToolConfig `mapstructure:"qoder"`
	Trae   AIToolConfig `mapstructure:"trae"`
}

// AIToolConfig holds settings for a specific AI tool
type AIToolConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	SettingsPath   string `mapstructure:"settings_path"`
	HookIntegration bool  `mapstructure:"hook_integration"`
}

// SyncConfig holds sync settings
type SyncConfig struct {
	OnCheckout  bool `mapstructure:"on_checkout"`
	OnMerge     bool `mapstructure:"on_merge"`
	Submodules  bool `mapstructure:"submodules"`
	Paths       bool `mapstructure:"paths"`
	Deps        bool `mapstructure:"deps"`
	Interval    int  `mapstructure:"interval"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
}

// HooksConfig holds custom hook commands
type HooksConfig struct {
	PreInit  []string `mapstructure:"pre_init"`
	PostInit []string `mapstructure:"post_init"`
	PreSync  []string `mapstructure:"pre_sync"`
	PostSync []string `mapstructure:"post_sync"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Project: ProjectConfig{
			Type: "auto",
			Root: ".",
		},
		Submodules: SubmoduleConfig{
			Enabled:        true,
			Strategy:       "local-clone-first",
			Parallel:       true,
			RetryOnFailure: true,
		},
		Dependencies: DepsConfig{
			Node: NodeDepsConfig{
				Enabled:         true,
				PackageManager:  "auto",
				LinkFromSource:  true,
				InstallStrategy: "link-first",
			},
			Go: GoDepsConfig{
				Enabled:    true,
				VendorMode: false,
			},
		},
		Ports: PortConfig{
			Enabled:    true,
			BaseOffset: 0,
			Services: []ServicePort{
				{Name: "api", Base: 8080, EnvKey: "API_PORT"},
				{Name: "frontend", Base: 3000, EnvKey: "FE_PORT"},
				{Name: "devserver", Base: 9090, EnvKey: "DEVSERVER_PORT"},
			},
		},
		Env: EnvConfig{
			Enabled:    true,
			OutputFile: ".worktree.env",
			Format:     "shell",
			Variables: map[string]string{
				"WORKTREE_NAME":   "${WT_NAME}",
				"WORKTREE_OFFSET": "${WT_OFFSET}",
			},
		},
		PathFixes: PathFixConfig{
			Enabled: true,
			Strategies: map[string]PathStrategy{
				"codegraph": {
					Enabled:  true,
					Paths:    []string{".codegraph/codegraph.db"},
					Strategy: "rewrite_absolute_to_relative",
				},
				"graphify": {
					Enabled:  true,
					Paths:    []string{"graph.json"},
					Strategy: "strip_cwd_prefix",
				},
				"config_files": {
					Enabled: true,
					Paths: []string{
						".cursor/settings.json",
						".claude/settings.json",
						".vscode/settings.json",
					},
					Strategy: "rewrite_paths",
				},
			},
		},
		AITools: AIToolsConfig{
			Claude: AIToolConfig{Enabled: true, SettingsPath: ".claude/settings.json"},
			Cursor: AIToolConfig{Enabled: true, SettingsPath: ".cursor/worktrees.json"},
			Codex:  AIToolConfig{Enabled: true, SettingsPath: ".codex/config.toml"},
			Qoder:  AIToolConfig{Enabled: true, SettingsPath: ".qoder/settings.local.json"},
			Trae:   AIToolConfig{Enabled: true, SettingsPath: ".trae"},
		},
		Sync: SyncConfig{
			OnCheckout: true,
			OnMerge:    true,
			Submodules: true,
			Paths:      true,
			Deps:       true,
			Interval:   0,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			MaxSize:    10,
			MaxBackups: 3,
		},
		Hooks: HooksConfig{},
	}
}

// Load loads configuration from multiple sources
func Load(cfgPath string) (*Config, error) {
	cfg := DefaultConfig()

	v := viper.New()
	v.SetConfigType("yaml")

	// Set defaults
	setDefaults(v)

	// Environment variables
	v.SetEnvPrefix("WTOOL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Config file locations (in order of priority)
	configPaths := []string{}
	
	if cfgPath != "" {
		configPaths = append(configPaths, cfgPath)
	}
	
	// Current directory
	configPaths = append(configPaths, ".wtool.yaml")
	
	// Home directory
	home, err := os.UserHomeDir()
	if err == nil {
		configPaths = append(configPaths, 
			filepath.Join(home, ".config", "wtool", "config.yaml"),
			filepath.Join(home, ".wtool.yaml"),
		)
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			v.SetConfigFile(path)
			if err := v.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("failed to read config %s: %w", path, err)
			}
			break
		}
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("version", "1.0")
	v.SetDefault("project.type", "auto")
	v.SetDefault("project.root", ".")
	v.SetDefault("submodules.enabled", true)
	v.SetDefault("submodules.strategy", "local-clone-first")
	v.SetDefault("submodules.parallel", true)
	v.SetDefault("submodules.retry_on_failure", true)
	v.SetDefault("dependencies.node.enabled", true)
	v.SetDefault("dependencies.node.package_manager", "auto")
	v.SetDefault("dependencies.node.link_from_source", true)
	v.SetDefault("dependencies.go.enabled", true)
	v.SetDefault("ports.enabled", true)
	v.SetDefault("ports.base_offset", 0)
	v.SetDefault("env.enabled", true)
	v.SetDefault("env.output_file", ".worktree.env")
	v.SetDefault("env.format", "shell")
	v.SetDefault("path_fixes.enabled", true)
	v.SetDefault("sync.on_checkout", true)
	v.SetDefault("sync.on_merge", true)
	v.SetDefault("sync.submodules", true)
	v.SetDefault("sync.paths", true)
	v.SetDefault("sync.deps", true)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "text")
	v.SetDefault("logging.max_size", 10)
	v.SetDefault("logging.max_backups", 3)
}
