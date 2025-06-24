package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mskelton/pool/internal/git"
	"github.com/mskelton/pool/internal/logger"
	"github.com/mskelton/pool/internal/pool"
	"github.com/mskelton/pool/internal/progress"
	"github.com/spf13/cobra"
)

var (
	convertRepo bool
	bareURL     string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize worktree pool",
	Long: `Initialize a worktree pool in the current repository.

You can also convert an existing repository to a bare repository with worktrees,
or clone a new repository as bare with a pool.`,
	Run: func(cmd *cobra.Command, args []string) {
		if bareURL != "" {
			if err := cloneBare(bareURL); err != nil {
				logger.Error("%v", err)
				os.Exit(1)
			}
			return
		}

		if convertRepo {
			if err := convertToBare(); err != nil {
				logger.Error("%v", err)
				os.Exit(1)
			}
			return
		}

		if err := initializePool(); err != nil {
			logger.Error("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&convertRepo, "convert", false, "Convert current repo to bare with worktrees")
	initCmd.Flags().StringVar(&bareURL, "bare", "", "Clone repository as bare with pool")
}

func initializePool() error {
	repo, err := git.NewRepository(".")
	if err != nil {
		return err
	}

	if repo.IsInWorktree() {
		return fmt.Errorf("cannot initialize pool from within a worktree. Please run from the main repository")
	}

	manager, err := pool.NewManager(repo)
	if err != nil {
		return err
	}

	return manager.Initialize(poolSize)
}

func cloneBare(url string) error {
	repoName := filepath.Base(url)
	repoName = strings.TrimSuffix(repoName, ".git")

	logger.Info("Cloning %s as bare repository...", url)

	err := progress.WithProgress("Cloning repository", func() error {
		return git.CloneBare(url, repoName)
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	if err := os.Chdir(repoName); err != nil {
		return err
	}

	err = progress.WithProgress("Configuring repository", func() error {
		return git.ConfigureBareRepo(".")
	})
	if err != nil {
		return err
	}

	defaultBranch, err := git.GetDefaultBranch(".")
	if err != nil {
		return err
	}

	logger.Info("Creating main worktree...")
	repo, err := git.NewRepository(".")
	if err != nil {
		return err
	}

	if err := repo.AddWorktree("main", defaultBranch); err != nil {
		return fmt.Errorf("failed to create main worktree: %w", err)
	}

	manager, err := pool.NewManager(repo)
	if err != nil {
		return err
	}

	if err := manager.Initialize(poolSize); err != nil {
		return err
	}

	sourcePath := os.Getenv("POOL_BINARY_PATH")
	if sourcePath == "" {
		possiblePaths := []string{
			filepath.Join(os.Getenv("HOME"), "go/bin/pool"),
			"/usr/local/bin/pool",
			filepath.Join(os.Getenv("HOME"), "dev/dotfiles/pool/pool"),
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				sourcePath = path
				break
			}
		}
	}

	if sourcePath != "" {
		if err := copyFile(sourcePath, "pool"); err == nil {
			os.Chmod("pool", 0755)
			logger.Info("Copied pool binary to repository")
		}
	}

	logger.Success("Setup complete!")
	logger.Info("Usage: cd %s && ./pool <branch-name>", repoName)

	return nil
}

func convertToBare() error {
	repo, err := git.NewRepository(".")
	if err != nil {
		return err
	}

	if repo.IsBare {
		return fmt.Errorf("repository is already bare")
	}

	if repo.IsInWorktree() {
		return fmt.Errorf("cannot convert from within a worktree. Please run from the main repository")
	}

	if repo.HasUncommittedChanges() {
		return fmt.Errorf("you have uncommitted changes. Please commit or stash them first")
	}

	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return err
	}

	repoPath, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	repoName := filepath.Base(repoPath)
	parentDir := filepath.Dir(repoPath)

	logger.Info("Converting %s to bare repository with worktrees...", repoName)

	var bareDir string
	err = progress.WithProgress("Creating bare clone", func() error {
		var e error
		bareDir, e = git.ConvertToBare(repoPath)
		return e
	})
	if err != nil {
		return err
	}

	if err := os.Chdir(bareDir); err != nil {
		return err
	}

	err = progress.WithProgress("Configuring repository", func() error {
		return git.ConfigureBareRepo(".")
	})
	if err != nil {
		return err
	}

	logger.Info("Creating main worktree...")
	bareRepo, err := git.NewRepository(".")
	if err != nil {
		return err
	}

	mainWorktreePath := filepath.Join("..", repoName+"-main")
	if err := bareRepo.AddWorktreeFromBranch(mainWorktreePath, currentBranch, currentBranch); err != nil {
		return fmt.Errorf("failed to create main worktree: %w", err)
	}

	manager, err := pool.NewManager(bareRepo)
	if err != nil {
		return err
	}

	if err := manager.Initialize(poolSize); err != nil {
		return err
	}

	poolPath := filepath.Join(repoPath, "bin/pool")
	if _, err := os.Stat(poolPath); err == nil {
		if err := copyFile(poolPath, "pool"); err == nil {
			os.Chmod("pool", 0755)
		}
	}

	logger.Success("Conversion complete!")
	fmt.Println()
	logger.Info("Next steps:")
	fmt.Printf("  1. Verify everything works in: %s\n", mainWorktreePath)
	fmt.Printf("  2. Remove old repository: rm -rf %s\n", repoPath)
	fmt.Printf("  3. Rename directories:\n")
	fmt.Printf("     mv %s %s\n", bareDir, filepath.Join(parentDir, repoName))
	fmt.Printf("     mv %s %s\n", mainWorktreePath, filepath.Join(parentDir, repoName, "main"))
	fmt.Printf("  4. Use: cd %s && ./pool <branch-name>\n", filepath.Join(parentDir, repoName))

	return nil
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}

	return nil
}
