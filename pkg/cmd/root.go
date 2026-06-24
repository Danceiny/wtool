package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/Danceiny/wtool/pkg/utils"
)

var (
	cfgFile string
	verbose bool
	quiet   bool
	noColor bool
	dryRun  bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "wtool",
	Short: "wtool - Universal Git Worktree Automation Tool",
	Long: `wtool automatically configures git worktrees for AI coding tools.

It detects and fixes:
  - Git submodules (auto-initialization)
  - Codegraph/Graphify absolute paths
  - pnpm workspace dependencies
  - AI tool configurations (Claude, Cursor, Codex, Qoder, Trae)
  - Port allocation for multiple worktrees`,
	Version: fmt.Sprintf("%s (built %s)", utils.Version, utils.BuildTime),
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output (debug level)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode (error level only)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "simulate execution without making changes")

	if os.Getenv("WTOOL_NO_COLOR") == "1" || os.Getenv("NO_COLOR") == "true" {
		noColor = true
	}
}
