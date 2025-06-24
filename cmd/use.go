package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mskelton/pool/internal/errors"
	"github.com/mskelton/pool/internal/git"
	"github.com/mskelton/pool/internal/logger"
	"github.com/mskelton/pool/internal/pool"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:     "use <branch>",
	Aliases: []string{"swim"},
	Short:   "Move a worktree back to the pool",
	Long: `Move a worktree from the bare checkout back to the pool.
This is useful when you're done working on a branch and want to
return the worktree to the pool for future use.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := returnWorktreeToPool(args[0]); err != nil {
			logger.Error("Failed to return worktree to pool: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}

func returnWorktreeToPool(branchName string) error {
	repo, err := git.NewRepository(".")
	if err != nil {
		return err
	}

	topLevel, err := repo.GetTopLevel()
	if err != nil {
		return err
	}

	safeBranchName := strings.ReplaceAll(branchName, "/", "-")
	worktreePath := filepath.Join(topLevel, safeBranchName)

	logger.Info("Returning worktree for branch: %s", branchName)

	worktrees, err := repo.ListWorktrees()
	if err != nil {
		return err
	}

	var found bool
	for _, wt := range worktrees {
		if wt.Path == worktreePath {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("no worktree found for branch %s at %s", branchName, worktreePath)
	}

	manager, err := pool.NewManager(repo)
	if err != nil {
		return err
	}

	poolSlot, err := findOrCreatePoolSlot(manager, repo)
	if err != nil {
		return err
	}

	poolPath := filepath.Join(topLevel, pool.PoolDir, poolSlot)

	if _, err := os.Stat(poolPath); err == nil {
		logger.Info("Removing existing pool worktree at %s", poolPath)
		if err := repo.RemoveWorktree(poolPath); err != nil {
			return errors.Wrapf(err, "failed to remove existing pool worktree")
		}
	}

	logger.Info("Moving worktree to pool slot: %s", poolSlot)
	if err := repo.MoveWorktree(worktreePath, poolPath); err != nil {
		return errors.Wrapf(err, "failed to move worktree to pool")
	}

	repo.RepairWorktrees()

	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(poolPath); err != nil {
		return err
	}

	logger.Info("Resetting pool worktree to clean state")
	if err := git.RunInDir(poolPath, "checkout", "-B", poolSlot, repo.DefaultBranch); err != nil {
		return errors.Wrapf(err, "failed to reset pool worktree")
	}

	if err := manager.MarkAvailable(poolSlot); err != nil {
		return err
	}

	logger.Success("Worktree returned to pool successfully")
	return nil
}

func findOrCreatePoolSlot(manager *pool.Manager, repo *git.Repository) (string, error) {
	topLevel, err := repo.GetTopLevel()
	if err != nil {
		return "", err
	}

	total, _ := manager.GetStatus()

	for i := 1; i <= total+1; i++ {
		poolName := fmt.Sprintf("%s%d", pool.PoolPrefix, i)
		poolPath := filepath.Join(topLevel, pool.PoolDir, poolName)

		if _, err := os.Stat(poolPath); os.IsNotExist(err) {
			return poolName, nil
		}
	}

	return "", fmt.Errorf("unable to find available pool slot")
}
