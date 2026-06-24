package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/Danceiny/wtool/pkg/detector"
	"github.com/Danceiny/wtool/pkg/utils"
)

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose worktree environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		worktreePath, _ := filepath.Abs(".")
		
		fmt.Println(color.CyanString("=== wtool Doctor ==="))
		fmt.Println()

		// Git check
		fmt.Println("Git:")
		if _, err := exec.LookPath("git"); err != nil {
			fmt.Println("  " + color.RedString("✗") + " git not found in PATH")
		} else {
			fmt.Println("  " + color.GreenString("✓") + " git available")
		}

		// Worktree check
		wtInfo, err := utils.GetWorktreeInfo(worktreePath)
		if err != nil {
			fmt.Println("  " + color.RedString("✗") + " Not in a git worktree")
			return nil
		}
		fmt.Printf("  "+color.GreenString("✓")+" Worktree: %s\n", wtInfo.Path)
		fmt.Printf("  "+color.GreenString("✓")+" Branch: %s\n", wtInfo.Branch)
		fmt.Printf("  "+color.GreenString("✓")+" Main worktree: %v\n", wtInfo.IsMain)

		// Project detection
		fmt.Println("\nProject:")
		proj, err := detector.DetectProject(worktreePath)
		if err != nil {
			fmt.Printf("  "+color.RedString("✗")+" %v\n", err)
		} else {
			fmt.Printf("  "+color.GreenString("✓")+" Type: %s\n", proj.Type)
			fmt.Printf("  "+color.GreenString("✓")+" Languages: %v\n", proj.Languages)
		}

		// Submodules
		fmt.Println("\nSubmodules:")
		subResult, err := detector.DetectSubmodules(worktreePath)
		if err != nil {
			fmt.Printf("  "+color.RedString("✗")+" %v\n", err)
		} else if !subResult.HasSubmodules {
			fmt.Println("  " + color.YellowString("-") + " No submodules")
		} else {
			fmt.Printf("  "+color.GreenString("✓")+" Found %d submodule(s)\n", len(subResult.Modules))
			for _, mod := range subResult.Modules {
				status := color.GreenString("✓")
				if !mod.IsInitialized {
					status = color.RedString("✗")
				}
				fmt.Printf("    %s %s\n", status, mod.Path)
			}
		}

		// pnpm
		fmt.Println("\nDependencies:")
		pnpmResult, err := detector.DetectPnpm(worktreePath)
		if err != nil {
			fmt.Printf("  "+color.RedString("✗")+" %v\n", err)
		} else {
			if pnpmResult.PackageManager != "none" {
				fmt.Printf("  "+color.GreenString("✓")+" Package manager: %s\n", pnpmResult.PackageManager)
			} else {
				fmt.Println("  " + color.YellowString("-") + " No Node.js project detected")
			}
		}

		// AI Tools
		fmt.Println("\nAI Tools:")
		aiResult := detector.DetectAITools(worktreePath)
		if !aiResult.AnyDetected {
			fmt.Println("  " + color.YellowString("-") + " No AI tool configs detected")
		} else {
			for _, tool := range aiResult.Tools {
				fmt.Printf("  "+color.GreenString("✓")+" %s: %s\n", tool.Name, tool.ConfigPath)
			}
		}
		if aiResult.AgentsMd {
			fmt.Println("  " + color.GreenString("✓") + " AGENTS.md found")
		}

		// Path issues
		fmt.Println("\nPath Issues:")
		pathResult, err := detector.DetectPathIssues(worktreePath)
		if err != nil {
			fmt.Printf("  "+color.RedString("✗")+" %v\n", err)
		} else if !pathResult.HasIssues {
			fmt.Println("  " + color.GreenString("✓") + " No path issues detected")
		} else {
			fmt.Printf("  "+color.YellowString("!")+" Found %d path issue(s)\n", len(pathResult.Fixes))
			for _, fix := range pathResult.Fixes {
				fmt.Printf("    %s %s: %s\n", color.YellowString("!"), fix.Type, fix.File)
			}
		}

		fmt.Println("\n" + color.CyanString("=== Diagnosis Complete ==="))
		return nil
	},
}
