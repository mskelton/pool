package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mskelton/pool/internal/errors"
)

const (
	ConfigFileName = ".poolrc.json"
	GlobalConfigFileName = ".poolrc"
)

type Config struct {
	PoolSize       int               `json:"pool_size,omitempty"`
	PoolPrefix     string            `json:"pool_prefix,omitempty"`
	DefaultBranch  string            `json:"default_branch,omitempty"`
	Editor         string            `json:"editor,omitempty"`
	AutoRefill     bool              `json:"auto_refill,omitempty"`
	CleanupOnExit  bool              `json:"cleanup_on_exit,omitempty"`
	Aliases        map[string]string `json:"aliases,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		PoolSize:      5,
		PoolPrefix:    "pool-",
		DefaultBranch: "main",
		Editor:        "code",
		AutoRefill:    true,
		CleanupOnExit: false,
		Aliases:       make(map[string]string),
	}
}

func Load() (*Config, error) {
	config := DefaultConfig()
	
	if err := config.loadFromEnv(); err != nil {
		return nil, err
	}
	
	if homeDir, err := os.UserHomeDir(); err == nil {
		globalConfigPath := filepath.Join(homeDir, GlobalConfigFileName)
		if err := config.loadFromFile(globalConfigPath); err != nil && !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "failed to load global config")
		}
	}
	
	if localPath, err := findLocalConfig(); err == nil {
		if err := config.loadFromFile(localPath); err != nil {
			return nil, errors.Wrap(err, "failed to load local config")
		}
	}
	
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	return config, nil
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.Wrap(err, "failed to write config file")
	}
	
	return nil
}

func (c *Config) Validate() error {
	if c.PoolSize < 1 {
		return errors.NewValidationError("pool_size", "", "must be at least 1")
	}
	
	if c.PoolSize > 50 {
		return errors.NewValidationError("pool_size", "", "must be at most 50")
	}
	
	if c.PoolPrefix == "" {
		return errors.NewValidationError("pool_prefix", "", "cannot be empty")
	}
	
	if c.DefaultBranch == "" {
		return errors.NewValidationError("default_branch", "", "cannot be empty")
	}
	
	if c.Editor == "" {
		return errors.NewValidationError("editor", "", "cannot be empty")
	}
	
	return nil
}

func (c *Config) loadFromEnv() error {
	if poolSize := os.Getenv("WORKTREE_POOL_SIZE"); poolSize != "" {
		var size int
		if _, err := fmt.Sscanf(poolSize, "%d", &size); err == nil {
			c.PoolSize = size
		}
	}
	
	if editor := os.Getenv("EDITOR"); editor != "" {
		c.Editor = editor
	}
	
	if editor := os.Getenv("POOL_EDITOR"); editor != "" {
		c.Editor = editor
	}
	
	return nil
}

func (c *Config) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	var fileConfig Config
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return errors.Wrap(err, "invalid JSON in config file")
	}
	
	c.merge(&fileConfig)
	
	return nil
}

func (c *Config) merge(other *Config) {
	if other.PoolSize > 0 {
		c.PoolSize = other.PoolSize
	}
	
	if other.PoolPrefix != "" {
		c.PoolPrefix = other.PoolPrefix
	}
	
	if other.DefaultBranch != "" {
		c.DefaultBranch = other.DefaultBranch
	}
	
	if other.Editor != "" {
		c.Editor = other.Editor
	}
	
	if other.AutoRefill != c.AutoRefill {
		c.AutoRefill = other.AutoRefill
	}
	
	if other.CleanupOnExit != c.CleanupOnExit {
		c.CleanupOnExit = other.CleanupOnExit
	}
	
	if other.Aliases != nil {
		if c.Aliases == nil {
			c.Aliases = make(map[string]string)
		}
		for k, v := range other.Aliases {
			c.Aliases[k] = v
		}
	}
}

func findLocalConfig() (string, error) {
	if _, err := os.Stat(ConfigFileName); err == nil {
		return ConfigFileName, nil
	}
	
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	gitRoot := strings.TrimSpace(string(output))
	configPath := filepath.Join(gitRoot, ConfigFileName)
	
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}
	
	return "", os.ErrNotExist
}