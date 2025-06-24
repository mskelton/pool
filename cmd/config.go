package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mskelton/pool/internal/config"
	"github.com/mskelton/pool/internal/logger"
	"github.com/spf13/cobra"
)

var (
	global bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage pool configuration",
	Long:  `Display or modify pool configuration settings.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			logger.Error("Failed to format config: %v", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		var configPath string

		if global {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				logger.Error("Failed to get home directory: %v", err)
				os.Exit(1)
			}
			configPath = filepath.Join(homeDir, config.GlobalConfigFileName)
		} else {
			configPath = config.ConfigFileName
		}

		if _, err := os.Stat(configPath); err == nil {
			logger.Warning("Configuration file already exists at %s", configPath)
			return
		}

		defaultCfg := config.DefaultConfig()

		if err := defaultCfg.Save(configPath); err != nil {
			logger.Error("Failed to save config: %v", err)
			os.Exit(1)
		}

		logger.Success("Created configuration file at %s", configPath)
		logger.Info("Edit this file to customize your pool settings")
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		switch key {
		case "pool_size", "pool-size":
			var size int
			if _, err := fmt.Sscanf(value, "%d", &size); err != nil {
				logger.Error("Invalid pool size: %s", value)
				os.Exit(1)
			}
			cfg.PoolSize = size

		case "editor":
			cfg.Editor = value

		case "default_branch", "default-branch":
			cfg.DefaultBranch = value

		case "pool_prefix", "pool-prefix":
			cfg.PoolPrefix = value

		case "auto_refill", "auto-refill":
			cfg.AutoRefill = value == "true" || value == "yes" || value == "1"

		case "cleanup_on_exit", "cleanup-on-exit":
			cfg.CleanupOnExit = value == "true" || value == "yes" || value == "1"

		default:
			logger.Error("Unknown configuration key: %s", key)
			logger.Info("Valid keys: pool_size, editor, default_branch, pool_prefix, auto_refill, cleanup_on_exit")
			os.Exit(1)
		}

		var configPath string
		if global {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				logger.Error("Failed to get home directory: %v", err)
				os.Exit(1)
			}
			configPath = filepath.Join(homeDir, config.GlobalConfigFileName)
		} else {
			configPath = config.ConfigFileName
		}

		if err := cfg.Save(configPath); err != nil {
			logger.Error("Failed to save config: %v", err)
			os.Exit(1)
		}

		logger.Success("Set %s = %s", key, value)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSetCmd)

	configCmd.PersistentFlags().BoolVarP(&global, "global", "g", false, "Use global configuration")
}
