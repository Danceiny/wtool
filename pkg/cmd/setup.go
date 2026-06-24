package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/Danceiny/wtool/pkg/hooks"
)

var (
	setupGlobal bool
	setupTool   string
)

func init() {
	setupCmd.Flags().BoolVar(&setupGlobal, "global", false, "Install hooks globally")
	setupCmd.Flags().StringVar(&setupTool, "tool", "", "Setup specific tool (claude, cursor, codex, qoder, trae)")
	
	setupCmd.AddCommand(setupHookCmd)
	setupCmd.AddCommand(setupConfigCmd)
	
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup wtool integrations",
	Long:  `Configure git hooks and AI tool integrations.`,
}

var setupHookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Install git hooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		if setupGlobal {
			// Install to global git hooks template
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			templateDir := filepath.Join(home, ".config", "git", "templates", "hooks")
			if err := hooks.InstallHooks(templateDir); err != nil {
				return err
			}
			fmt.Printf("%s Installed hooks to global template: %s\n", color.GreenString("[wtool]"), templateDir)
		} else {
			// Install to current repo
			repoPath, _ := filepath.Abs(".")
			gitDir, err := hooks.GetGitDir(repoPath)
			if err != nil {
				return err
			}
			hooksDir := filepath.Join(gitDir, "hooks")
			if err := hooks.InstallHooks(hooksDir); err != nil {
				return err
			}
			fmt.Printf("%s Installed hooks to repo: %s\n", color.GreenString("[wtool]"), hooksDir)
		}
		return nil
	},
}

var setupConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate AI tool configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(color.CyanString("AI Tool Configuration Setup"))
		fmt.Println()
		fmt.Println("Add the following to your AI tool configuration:")
		fmt.Println()
		fmt.Println("Claude (.claude/settings.json):")
		fmt.Println(`  "hooks": {
    "SessionStart": [
      "wtool init --quiet"
    ]
  }`)
		fmt.Println()
		fmt.Println("Cursor (.cursor/worktrees.json):")
		fmt.Println(`  {
    "setup-worktree": ["wtool init"]
  }`)
		fmt.Println()
		fmt.Println("Codex (AGENTS.md):")
		fmt.Println("  Run `wtool init` before starting work in a new worktree.")
		fmt.Println()
		fmt.Println("Qoder (.qoder/settings.local.json):")
		fmt.Println("  Add wtool init to your pre-task hooks.")
		return nil
	},
}
