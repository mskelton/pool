package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.PoolSize != 5 {
		t.Errorf("Expected default pool size 5, got %d", cfg.PoolSize)
	}

	if cfg.PoolPrefix != "pool-" {
		t.Errorf("Expected default pool prefix 'pool-', got %s", cfg.PoolPrefix)
	}

	if cfg.Editor != "code" {
		t.Errorf("Expected default editor 'code', got %s", cfg.Editor)
	}

	if !cfg.AutoRefill {
		t.Error("Expected auto refill to be true by default")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "Valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "Invalid pool size (too small)",
			config: &Config{
				PoolSize:      0,
				PoolPrefix:    "pool-",
				DefaultBranch: "main",
				Editor:        "code",
			},
			wantErr: true,
		},
		{
			name: "Invalid pool size (too large)",
			config: &Config{
				PoolSize:      100,
				PoolPrefix:    "pool-",
				DefaultBranch: "main",
				Editor:        "code",
			},
			wantErr: true,
		},
		{
			name: "Empty pool prefix",
			config: &Config{
				PoolSize:      5,
				PoolPrefix:    "",
				DefaultBranch: "main",
				Editor:        "code",
			},
			wantErr: true,
		},
		{
			name: "Empty editor",
			config: &Config{
				PoolSize:      5,
				PoolPrefix:    "pool-",
				DefaultBranch: "main",
				Editor:        "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigSaveLoad(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a config
	cfg := &Config{
		PoolSize:      8,
		PoolPrefix:    "wp-",
		DefaultBranch: "develop",
		Editor:        "vim",
		AutoRefill:    false,
		CleanupOnExit: true,
		Aliases: map[string]string{
			"feat": "feature",
			"fix":  "bugfix",
		},
	}

	// Save it
	configPath := filepath.Join(tmpDir, "test-config.json")
	if err := cfg.Save(configPath); err != nil {
		t.Fatal(err)
	}

	// Load it back
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}

	// Verify values
	if loaded.PoolSize != cfg.PoolSize {
		t.Errorf("Expected pool size %d, got %d", cfg.PoolSize, loaded.PoolSize)
	}

	if loaded.PoolPrefix != cfg.PoolPrefix {
		t.Errorf("Expected pool prefix %s, got %s", cfg.PoolPrefix, loaded.PoolPrefix)
	}

	if loaded.Editor != cfg.Editor {
		t.Errorf("Expected editor %s, got %s", cfg.Editor, loaded.Editor)
	}

	if loaded.AutoRefill != cfg.AutoRefill {
		t.Errorf("Expected auto refill %v, got %v", cfg.AutoRefill, loaded.AutoRefill)
	}

	if len(loaded.Aliases) != len(cfg.Aliases) {
		t.Errorf("Expected %d aliases, got %d", len(cfg.Aliases), len(loaded.Aliases))
	}
}

func TestConfigMerge(t *testing.T) {
	base := &Config{
		PoolSize:      5,
		PoolPrefix:    "pool-",
		DefaultBranch: "main",
		Editor:        "code",
		AutoRefill:    true,
		Aliases:       map[string]string{"f": "feature"},
	}

	override := &Config{
		PoolSize:   10,
		Editor:     "nvim",
		AutoRefill: false,
		Aliases:    map[string]string{"b": "bugfix"},
	}

	base.merge(override)

	// Check merged values
	if base.PoolSize != 10 {
		t.Errorf("Expected pool size 10, got %d", base.PoolSize)
	}

	if base.Editor != "nvim" {
		t.Errorf("Expected editor nvim, got %s", base.Editor)
	}

	if base.AutoRefill != false {
		t.Error("Expected auto refill to be false")
	}

	// Check that unset values weren't changed
	if base.PoolPrefix != "pool-" {
		t.Errorf("Expected pool prefix to remain 'pool-', got %s", base.PoolPrefix)
	}

	// Check aliases were merged
	if len(base.Aliases) != 2 {
		t.Errorf("Expected 2 aliases after merge, got %d", len(base.Aliases))
	}
}
