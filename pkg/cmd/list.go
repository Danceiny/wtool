package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/Danceiny/wtool/pkg/utils"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all worktrees",
	RunE: func(cmd *cobra.Command, args []string) error {
		worktrees, err := utils.ListWorktrees(".")
		if err != nil {
			return err
		}

		fmt.Println(color.CyanString("Worktrees:"))
		fmt.Println()

		for i, wt := range worktrees {
			marker := " "
			if i == 0 {
				marker = color.GreenString("*")
			}
			
			branch := wt.Branch
			if branch == "" {
				branch = color.YellowString("(detached)")
			}
			
			fmt.Printf("%s [%d] %s\n", marker, i, filepath.Base(wt.Path))
			fmt.Printf("    Path:   %s\n", wt.Path)
			fmt.Printf("    Branch: %s\n", branch)
			fmt.Printf("    Commit: %s\n", wt.Commit[:7])
			if i == 0 {
				fmt.Println("    " + color.GreenString("(main worktree)"))
			}
			fmt.Println()
		}

		return nil
	},
}
