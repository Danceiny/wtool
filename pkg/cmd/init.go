package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/Danceiny/wtool/pkg/config"
	"github.com/Danceiny/wtool/pkg/detector"
	"github.com/Danceiny/wtool/pkg/fixer"
	"github.com/Danceiny/wtool/pkg/utils"
)

var (
	skipSubmodules bool
	skipDeps       bool
	skipPaths      bool
	skipPorts      bool
	sourceRoot     string
)

func init() {
	initCmd.Flags().BoolVar(&skipSubmodules, "skip-submodules", false, "Skip submodule initialization")
	initCmd.Flags().BoolVar(&skipDeps, "skip-deps", false, "Skip dependency installation")
	initCmd.Flags().BoolVar(&skipPaths, "skip-paths", false, "Skip path fixes")
	initCmd.Flags().BoolVar(&skipPorts, "skip-ports", false, "Skip port allocation")
	initCmd.Flags().StringVar(&sourceRoot, "source", "", "Source worktree for local cloning/linking")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize current worktree",
	Long:  `Auto-detect and fix worktree configuration issues.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		worktreePath := "."
		if len(args) > 0 {
			worktreePath = args[0]
		}

		absPath, err := filepath.Abs(worktreePath)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}
		worktreePath = absPath

		// Validate git repo
		wtInfo, err := utils.GetWorktreeInfo(worktreePath)
		if err != nil {
			return fmt.Errorf("not a valid git worktree: %w", err)
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Determine source root
		srcRoot := sourceRoot
		if srcRoot == "" {
			srcRoot = cfg.Submodules.SourceRoot
		}
		if srcRoot == "" || srcRoot == "${PRIMARY_WORKTREE}" {
			primary, err := utils.GetPrimaryWorktree(worktreePath)
			if err == nil {
				srcRoot = primary
			}
		}

		fmt.Printf("%s Initializing worktree: %s\n", color.GreenString("[wtool]"), filepath.Base(worktreePath))
		if dryRun {
			fmt.Println(color.YellowString("[wtool] DRY RUN - no changes will be made"))
		}

		// Step 1: Detect project
		fmt.Println(color.BlueString("[wtool] Step 1/5: Detecting project structure..."))
		proj, err := detector.DetectProject(worktreePath)
		if err != nil {
			return err
		}
		fmt.Printf("  Project: %s (%s)\n", proj.Name, proj.Type)
		fmt.Printf("  Languages: %v\n", proj.Languages)

		// Step 2: Submodules
		if !skipSubmodules && cfg.Submodules.Enabled {
			fmt.Println(color.BlueString("[wtool] Step 2/5: Initializing submodules..."))
			subResult, err := detector.DetectSubmodules(worktreePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
			} else if subResult.HasSubmodules {
				fmt.Printf("  Found %d submodule(s)\n", len(subResult.Modules))
				if subResult.NeedInit {
					if !dryRun {
						fixer := fixer.NewSubmoduleFixer(worktreePath, srcRoot, cfg.Submodules.Strategy)
						if err := fixer.Fix(); err != nil {
							fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
						}
					} else {
						fmt.Println("  Would initialize submodules")
					}
				} else {
					fmt.Println("  Submodules already initialized")
				}
			} else {
				fmt.Println("  No submodules found")
			}
		}

		// Step 3: Dependencies
		if !skipDeps && cfg.Dependencies.Node.Enabled {
			fmt.Println(color.BlueString("[wtool] Step 3/5: Setting up dependencies..."))
			if !dryRun {
				depsFixer := fixer.NewDepsFixer(worktreePath, srcRoot, cfg.Dependencies.Node.InstallStrategy)
				if err := depsFixer.Fix(); err != nil {
					fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
				}
			} else {
				fmt.Println("  Would setup dependencies")
			}
		}

		// Step 4: Path fixes
		if !skipPaths && cfg.PathFixes.Enabled {
			fmt.Println(color.BlueString("[wtool] Step 4/5: Checking path issues..."))
			pathResult, err := detector.DetectPathIssues(worktreePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
			} else if pathResult.HasIssues {
				fmt.Printf("  Found %d path issue(s)\n", len(pathResult.Fixes))
				if !dryRun {
					pathFixer := fixer.NewPathFixer(worktreePath)
					if err := pathFixer.Fix(); err != nil {
						fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
					}
				} else {
					fmt.Println("  Would fix paths")
				}
			} else {
				fmt.Println("  No path issues found")
			}
		}

		// Step 5: Port allocation
		if !skipPorts && cfg.Ports.Enabled {
			fmt.Println(color.BlueString("[wtool] Step 5/5: Allocating ports..."))
			if !dryRun {
				var services []fixer.ServicePort
				for _, s := range cfg.Ports.Services {
					services = append(services, fixer.ServicePort{
						Name:   s.Name,
						Base:   s.Base,
						EnvKey: s.EnvKey,
					})
				}
				portFixer := fixer.NewPortFixer(worktreePath, cfg.Ports.BaseOffset, services)
				if err := portFixer.Fix(); err != nil {
					fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
				} else {
					fmt.Println("  Generated .worktree.env")
				}
			} else {
				fmt.Println("  Would allocate ports")
			}
		}

		fmt.Printf("%s Worktree initialization complete!\n", color.GreenString("[wtool]"))
		if !wtInfo.IsMain {
			fmt.Println("\nQuick Start:")
			fmt.Printf("  cd %s\n", worktreePath)
			fmt.Println("  source .worktree.env")
		}

		return nil
	},
}
