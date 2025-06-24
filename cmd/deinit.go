package cmd

import (
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
	force bool
)

var deinitCmd = &cobra.Command{
	Use:   "deinit",
	Short: "Remove worktree pool from the current repository",
	Long: `Remove the worktree pool infrastructure from the current repository.

This command will:
- Remove all pool worktrees
- Delete the .worktree-pool directory and its contents
- Clean up pool-related configuration

Note: This will NOT convert a bare repository back to a normal repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDeinit(); err != nil {
			logger.Error("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(deinitCmd)
	deinitCmd.Flags().BoolVar(&force, "force", false, "Force removal without confirmation")
}

func runDeinit() error {
	repo, err := git.NewRepository(".")
	if err != nil {
		return err
	}

	if repo.IsInWorktree() {
		return fmt.Errorf("cannot deinitialize pool from within a worktree. Please run from the main repository")
	}

	topLevel, err := repo.GetTopLevel()
	if err != nil {
		return err
	}

	poolPath := filepath.Join(topLevel, pool.PoolDir)
	
	if _, err := os.Stat(poolPath); os.IsNotExist(err) {
		return fmt.Errorf("no worktree pool found in this repository")
	}

	logger.Info("Removing worktree pool from repository...")

	if !force {
		fmt.Println()
		logger.Warning("This will remove all pool worktrees and configuration.")
		if !confirm("Are you sure you want to continue?") {
			logger.Info("Deinit cancelled")
			return nil
		}
	}

	worktrees, err := repo.ListWorktrees()
	if err != nil {
		return err
	}

	removedCount := 0
	for _, wt := range worktrees {
		if strings.Contains(wt.Path, pool.PoolDir) {
			poolName := filepath.Base(wt.Path)
			logger.Info("Removing pool worktree: %s", poolName)
			
			if err := repo.RemoveWorktree(wt.Path); err != nil {
				logger.Error("Failed to remove worktree %s: %v", poolName, err)
				if !force {
					return err
				}
			} else {
				removedCount++
			}
		}
	}

	logger.Info("Removing pool directory: %s", poolPath)
	if err := os.RemoveAll(poolPath); err != nil {
		logger.Error("Failed to remove pool directory: %v", err)
		if !force {
			return err
		}
	}

	logger.Success("Successfully removed worktree pool (%d worktrees removed)", removedCount)
	
	if repo.IsBare {
		fmt.Println()
		logger.Info("Note: This repository is still a bare repository.")
		logger.Info("To work with it, you'll need to use regular git worktree commands.")
	}

	return nil
}