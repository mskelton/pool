package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mskelton/pool/internal/git"
	"github.com/mskelton/pool/internal/logger"
	"github.com/mskelton/pool/internal/pool"
	"github.com/spf13/cobra"
)

var (
	dryRun bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean [type]",
	Short: "Clean up worktrees",
	Long: `Clean up worktrees based on type:
  orphaned - Remove orphaned worktrees
  stale    - Remove worktrees for deleted branches
  merged   - Remove worktrees for merged branches
  pool     - Reset pool worktrees to clean state
  all      - Run all cleanup tasks`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runClean(args[0]); err != nil {
			logger.Error("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be cleaned")
}

func runClean(cleanType string) error {
	repo, err := git.NewRepository(".")
	if err != nil {
		return err
	}
	
	switch cleanType {
	case "orphaned":
		return cleanOrphaned(repo)
	case "stale":
		return cleanStale(repo)
	case "merged":
		return cleanMerged(repo)
	case "pool":
		return cleanPool(repo)
	case "all":
		logger.Info("Running all cleanup tasks...")
		if err := cleanOrphaned(repo); err != nil {
			logger.Error("Failed to clean orphaned: %v", err)
		}
		fmt.Println()
		if err := cleanStale(repo); err != nil {
			logger.Error("Failed to clean stale: %v", err)
		}
		fmt.Println()
		if err := cleanPool(repo); err != nil {
			logger.Error("Failed to clean pool: %v", err)
		}
		fmt.Println()
		if err := cleanMerged(repo); err != nil {
			logger.Error("Failed to clean merged: %v", err)
		}
		logger.Success("All cleanup tasks completed")
		return nil
	default:
		return fmt.Errorf("unknown clean type: %s", cleanType)
	}
}

func cleanOrphaned(repo *git.Repository) error {
	logger.Info("Cleaning orphaned worktrees...")
	
	if dryRun {
		logger.Warning("Dry run mode - no changes made")
		return nil
	}
	
	if err := repo.PruneWorktrees(); err != nil {
		return err
	}
	
	logger.Success("Orphaned worktrees cleaned")
	return nil
}

func cleanStale(repo *git.Repository) error {
	logger.Info("Finding stale branches in worktrees...")
	
	worktrees, err := repo.ListWorktrees()
	if err != nil {
		return err
	}
	
	for _, wt := range worktrees {
		if strings.Contains(wt.Path, pool.PoolDir) {
			continue
		}
		
		if wt.Branch != "" && !repo.RemoteBranchExists(wt.Branch) {
			logger.Warning("Branch '%s' not found on remote (worktree: %s)", wt.Branch, wt.Path)
			
			if !dryRun && confirm(fmt.Sprintf("Remove worktree for deleted branch '%s'?", wt.Branch)) {
				if err := repo.RemoveWorktree(wt.Path); err != nil {
					logger.Error("Failed to remove worktree: %v", err)
				} else {
					logger.Success("Removed worktree: %s", wt.Path)
				}
			}
		}
	}
	
	if dryRun {
		logger.Warning("Dry run mode - no changes made")
	}
	
	return nil
}

func cleanMerged(repo *git.Repository) error {
	logger.Info("Finding worktrees with merged branches...")
	
	if err := repo.FetchOriginPrune(); err != nil {
		return err
	}
	
	mergedBranches, err := repo.GetMergedBranches()
	if err != nil {
		return err
	}
	
	worktrees, err := repo.ListWorktrees()
	if err != nil {
		return err
	}
	
	for _, wt := range worktrees {
		if strings.Contains(wt.Path, pool.PoolDir) || wt.Bare {
			continue
		}
		
		for _, merged := range mergedBranches {
			if wt.Branch == merged {
				logger.Warning("Branch '%s' has been merged (worktree: %s)", wt.Branch, wt.Path)
				
				if !dryRun && confirm(fmt.Sprintf("Remove worktree for merged branch '%s'?", wt.Branch)) {
					if err := repo.RemoveWorktree(wt.Path); err != nil {
						logger.Error("Failed to remove worktree: %v", err)
					} else {
						logger.Success("Removed worktree and branch: %s", wt.Branch)
					}
				}
				break
			}
		}
	}
	
	if dryRun {
		logger.Warning("Dry run mode - no changes made")
	}
	
	return nil
}

func cleanPool(repo *git.Repository) error {
	logger.Info("Resetting pool worktrees...")
	
	manager, err := pool.NewManager(repo)
	if err != nil {
		return err
	}
	
	worktrees, err := repo.ListWorktrees()
	if err != nil {
		return err
	}
	
	for _, wt := range worktrees {
		if strings.Contains(wt.Path, pool.PoolDir) {
			poolName := filepath.Base(wt.Path)
			
			if status, ok := manager.Status.Worktrees[poolName]; ok && status == pool.StatusAvailable {
				oldDir, _ := os.Getwd()
				os.Chdir(wt.Path)
				
				if repo.HasUncommittedChanges() {
					logger.Warning("Pool worktree %s has uncommitted changes", poolName)
					
					if !dryRun && confirm(fmt.Sprintf("Reset pool worktree %s?", poolName)) {
						git.RunInDir(wt.Path, "reset", "--hard")
						git.RunInDir(wt.Path, "clean", "-fd")
						logger.Success("Reset pool worktree: %s", poolName)
					}
				}
				
				os.Chdir(oldDir)
			}
		}
	}
	
	if dryRun {
		logger.Warning("Dry run mode - no changes made")
	}
	
	return nil
}

func confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N] ", prompt)
	
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}