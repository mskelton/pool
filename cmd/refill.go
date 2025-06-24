package cmd

import (
	"os"

	"github.com/mskelton/pool/internal/git"
	"github.com/mskelton/pool/internal/logger"
	"github.com/mskelton/pool/internal/pool"
	"github.com/spf13/cobra"
)

var refillCmd = &cobra.Command{
	Use:   "refill",
	Short: "Refill the worktree pool",
	Run: func(cmd *cobra.Command, args []string) {
		if err := refillPool(); err != nil {
			logger.Error("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(refillCmd)
}

func refillPool() error {
	repo, err := git.NewRepository(".")
	if err != nil {
		return err
	}

	manager, err := pool.NewManager(repo)
	if err != nil {
		return err
	}

	return manager.Refill(poolSize)
}
