package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mskelton/pool/internal/git"
	"github.com/mskelton/pool/internal/logger"
	"github.com/mskelton/pool/internal/pool"
)

func createWorktree(branchName string) error {
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

	logger.Info("Setting up worktree for branch: %s", branchName)

	worktrees, err := repo.ListWorktrees()
	if err != nil {
		return err
	}

	for _, wt := range worktrees {
		if wt.Path == worktreePath {
			logger.Warning("Worktree already exists at %s", worktreePath)
			return openInEditor(worktreePath)
		}
	}

	manager, err := pool.NewManager(repo)
	if err != nil {
		return err
	}

	poolPath, poolName, err := manager.GetAvailable()
	if err != nil {
		logger.Warning("No available worktrees in pool. Creating new worktree...")
		return createWorktreeDirect(repo, worktreePath, branchName)
	}

	logger.Info("Using pool worktree: %s", poolName)

	if err := manager.MarkInUse(poolName); err != nil {
		return err
	}

	if _, err := os.Stat(worktreePath); err == nil {
		logger.Warning("Directory already exists at %s", worktreePath)
		manager.MarkAvailable(poolName)
		return openInEditor(worktreePath)
	}

	if err := setupBranchInPool(repo, poolPath, branchName); err != nil {
		manager.MarkAvailable(poolName)
		return err
	}

	if err := repo.MoveWorktree(poolPath, worktreePath); err != nil {
		manager.MarkAvailable(poolName)
		return err
	}

	repo.RepairWorktrees()

	go refillPoolAsync(manager)

	logger.Success("Opening %s in VS Code...", worktreePath)
	if err := openInEditor(worktreePath); err != nil {
		return err
	}

	fmt.Println()
	logger.Info("Worktree ready at: %s", worktreePath)
	logger.Info("Branch: %s", branchName)

	return nil
}

func createWorktreeDirect(repo *git.Repository, worktreePath, branchName string) error {
	if repo.RemoteBranchExists(branchName) {
		logger.Info("Branch exists remotely, checking out...")
		return repo.AddWorktreeFromBranch(worktreePath, branchName, fmt.Sprintf("origin/%s", branchName))
	}

	logger.Info("Creating new branch...")
	return repo.AddWorktree(worktreePath, branchName)
}

func setupBranchInPool(repo *git.Repository, poolPath, branchName string) error {
	// Fetch latest changes
	if err := git.RunInDir(poolPath, "fetch", "origin"); err != nil {
		return err
	}

	if repo.RemoteBranchExists(branchName) {
		logger.Info("Checking out existing branch...")
		return repo.CheckoutNewBranch(branchName, fmt.Sprintf("origin/%s", branchName))
	}

	logger.Info("Creating new branch...")
	return repo.CheckoutNewBranch(branchName, repo.DefaultBranch)
}

func openInEditor(path string) error {
	editor := cfg.Editor
	if editor == "" {
		editor = "code"
	}

	var cmd *exec.Cmd
	switch editor {
	case "code", "subl", "atom":
		cmd = exec.Command(editor, ".")
	case "vim", "nvim", "emacs":
		cmd = exec.Command(editor)
	default:
		cmd = exec.Command(editor, ".")
	}

	cmd.Dir = path
	return cmd.Run()
}

func refillPoolAsync(manager *pool.Manager) {
	time.Sleep(2 * time.Second)

	size := poolSize
	if size == 0 {
		size = pool.DefaultPoolSize
	}

	if err := manager.Refill(size); err != nil {
		logger.Error("Failed to refill pool: %v", err)
	}
}
