package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/mskelton/pool/internal/git"
	"github.com/mskelton/pool/internal/logger"
	"github.com/mskelton/pool/internal/pool"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show pool and worktree status",
	Run: func(cmd *cobra.Command, args []string) {
		if err := showStatus(); err != nil {
			logger.Error("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func showStatus() error {
	repo, err := git.NewRepository(".")
	if err != nil {
		return err
	}
	
	manager, err := pool.NewManager(repo)
	if err != nil {
		return err
	}
	
	logger.Info("Worktree pool status:")
	fmt.Println()
	
	total, available := manager.GetStatus()
	
	worktrees, err := repo.ListWorktrees()
	if err != nil {
		return err
	}
	
	for _, wt := range worktrees {
		if strings.Contains(wt.Path, pool.PoolDir) {
			name := filepath.Base(wt.Path)
			if status, ok := manager.Status.Worktrees[name]; ok {
				if status == pool.StatusAvailable {
					fmt.Printf("  %s %s - available\n", color.GreenString("●"), name)
				} else {
					fmt.Printf("  %s %s - in use\n", color.RedString("●"), name)
				}
			}
		}
	}
	
	fmt.Println()
	fmt.Printf("Pool size: %d\n", total)
	fmt.Printf("Available: %d\n", available)
	fmt.Println()
	
	logger.Info("Active worktrees:")
	for _, wt := range worktrees {
		if !strings.Contains(wt.Path, pool.PoolDir) && !wt.Bare {
			fmt.Printf("  %s (%s)\n", wt.Path, wt.Branch)
		}
	}
	
	return nil
}