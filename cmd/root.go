package cmd

import (
	"fmt"
	"os"

	"github.com/mskelton/pool/internal/config"
	"github.com/spf13/cobra"
)

var (
	poolSize int
	cfg      *config.Config
	rootCmd  = &cobra.Command{
		Use:   "pool",
		Short: "Fast worktree management with pre-seeded pool",
		Long: `pool provides instant worktree creation by maintaining a pool of
pre-created worktrees that can be claimed and renamed on demand.`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			
			if err := createWorktree(args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}
	
	rootCmd.PersistentFlags().IntVar(&poolSize, "pool-size", cfg.PoolSize, "Number of pre-seeded worktrees")
	
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("pool-size") {
			cfg.PoolSize = poolSize
		} else {
			poolSize = cfg.PoolSize
		}
	}
}