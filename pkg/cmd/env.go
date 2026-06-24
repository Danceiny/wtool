package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/Danceiny/wtool/pkg/utils"
)

var envFormat string

func init() {
	envCmd.Flags().StringVarP(&envFormat, "format", "f", "shell", "Output format: shell, json, env")
	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Output environment variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		worktreePath, _ := filepath.Abs(".")
		
		wtInfo, err := utils.GetWorktreeInfo(worktreePath)
		if err != nil {
			return err
		}

		wtIndex, _ := utils.GetWorktreeIndex(worktreePath)
		wtName := filepath.Base(worktreePath)

		env := map[string]string{
			"WORKTREE_NAME":   wtName,
			"WORKTREE_PATH":   worktreePath,
			"WORKTREE_INDEX":  fmt.Sprintf("%d", wtIndex),
			"WORKTREE_BRANCH": wtInfo.Branch,
			"WORKTREE_COMMIT": wtInfo.Commit,
			"WORKTREE_IS_MAIN": fmt.Sprintf("%v", wtInfo.IsMain),
		}

		switch envFormat {
		case "json":
			data, err := json.MarshalIndent(env, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		case "env", "shell":
			for k, v := range env {
				fmt.Printf("export %s=%s\n", k, v)
			}
		default:
			return fmt.Errorf("unknown format: %s", envFormat)
		}

		return nil
	},
}
