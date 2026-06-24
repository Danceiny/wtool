package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/Danceiny/wtool/pkg/config"
	"github.com/Danceiny/wtool/pkg/fixer"
	"github.com/Danceiny/wtool/pkg/utils"
)

var (
	syncSubmodules bool
	syncPaths      bool
	syncDeps       bool
)

func init() {
	syncCmd.Flags().BoolVar(&syncSubmodules, "submodules", true, "Sync submodules")
	syncCmd.Flags().BoolVar(&syncPaths, "paths", true, "Fix paths")
	syncCmd.Flags().BoolVar(&syncDeps, "deps", true, "Sync dependencies")
	rootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync worktree state",
	Long:  `Synchronize worktree with main repository (submodules, paths, deps).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		worktreePath, _ := filepath.Abs(".")
		
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		srcRoot := cfg.Submodules.SourceRoot
		if srcRoot == "" || srcRoot == "${PRIMARY_WORKTREE}" {
			primary, err := utils.GetPrimaryWorktree(worktreePath)
			if err == nil {
				srcRoot = primary
			}
		}

		fmt.Printf("%s Syncing worktree...\n", color.GreenString("[wtool]"))

		if syncSubmodules && cfg.Submodules.Enabled {
			fmt.Println(color.BlueString("[wtool] Syncing submodules..."))
			fixer := fixer.NewSubmoduleFixer(worktreePath, srcRoot, cfg.Submodules.Strategy)
			if err := fixer.Fix(); err != nil {
				fmt.Printf("  Warning: %v\n", err)
			}
		}

		if syncPaths && cfg.PathFixes.Enabled {
			fmt.Println(color.BlueString("[wtool] Fixing paths..."))
			pathFixer := fixer.NewPathFixer(worktreePath)
			if err := pathFixer.Fix(); err != nil {
				fmt.Printf("  Warning: %v\n", err)
			}
		}

		if syncDeps && cfg.Dependencies.Node.Enabled {
			fmt.Println(color.BlueString("[wtool] Syncing dependencies..."))
			depsFixer := fixer.NewDepsFixer(worktreePath, srcRoot, cfg.Dependencies.Node.InstallStrategy)
			if err := depsFixer.Fix(); err != nil {
				fmt.Printf("  Warning: %v\n", err)
			}
		}

		fmt.Printf("%s Sync complete!\n", color.GreenString("[wtool]"))
		return nil
	},
}
