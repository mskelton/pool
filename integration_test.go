package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mskelton/pool/internal/git"
	"github.com/mskelton/pool/internal/pool"
)

func TestPoolIntegration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pool-integration-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := initTestRepo(tmpDir); err != nil {
		t.Fatal(err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	t.Run("PoolInit", func(t *testing.T) {
		repo, err := git.NewRepository(".")
		if err != nil {
			t.Fatal(err)
		}

		manager, err := pool.NewManager(repo)
		if err != nil {
			t.Fatal(err)
		}

		if err := manager.Initialize(3); err != nil {
			t.Fatal(err)
		}

		poolDir := filepath.Join(tmpDir, ".worktree-pool")
		if _, err := os.Stat(poolDir); os.IsNotExist(err) {
			t.Error("Pool directory was not created")
		}

		entries, err := os.ReadDir(poolDir)
		if err != nil {
			t.Fatal(err)
		}

		if len(entries) != 4 {
			t.Errorf("Expected 4 entries in pool directory, got %d", len(entries))
		}

		totalWorktrees, availableWorktrees := manager.GetStatus()
		if totalWorktrees != 3 {
			t.Errorf("Expected 3 total worktrees, got %d", totalWorktrees)
		}
		if availableWorktrees != 3 {
			t.Errorf("Expected 3 available worktrees, got %d", availableWorktrees)
		}
	})

	t.Run("UsePoolWorktree", func(t *testing.T) {
		repo, err := git.NewRepository(".")
		if err != nil {
			t.Fatal(err)
		}

		manager, err := pool.NewManager(repo)
		if err != nil {
			t.Fatal(err)
		}

		poolPath, poolName, err := manager.GetAvailable()
		if err != nil {
			t.Fatal(err)
		}

		if err := manager.MarkInUse(poolName); err != nil {
			t.Fatal(err)
		}

		_, availableCount := manager.GetStatus()
		if availableCount != 2 {
			t.Errorf("Expected 2 available worktrees after marking one in use, got %d", availableCount)
		}

		if err := os.Chdir(poolPath); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command("git", "checkout", "-b", "feature-test")
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}

		testFile := filepath.Join(poolPath, "feature.txt")
		if err := os.WriteFile(testFile, []byte("feature"), 0644); err != nil {
			t.Fatal(err)
		}

		os.Chdir(tmpDir)
		newPath := filepath.Join(tmpDir, "feature-test")
		if err := repo.MoveWorktree(poolPath, newPath); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			t.Error("Worktree was not moved to new location")
		}

		if _, err := os.Stat(poolPath); !os.IsNotExist(err) {
			t.Error("Old worktree location still exists")
		}
	})

	t.Run("RefillPool", func(t *testing.T) {
		repo, err := git.NewRepository(".")
		if err != nil {
			t.Fatal(err)
		}

		manager, err := pool.NewManager(repo)
		if err != nil {
			t.Fatal(err)
		}

		if err := manager.Refill(3); err != nil {
			t.Fatal(err)
		}

		totalCount, _ := manager.GetStatus()
		if totalCount != 3 {
			t.Errorf("Expected 3 total worktrees after refill, got %d", totalCount)
		}
	})
}

func TestPoolCLI(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "pool-test", ".")
	if err := cmd.Run(); err != nil {
		t.Fatal("Failed to build pool binary:", err)
	}
	defer os.Remove("pool-test")

	tmpDir, err := os.MkdirTemp("", "pool-cli-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := initTestRepo(tmpDir); err != nil {
		t.Fatal(err)
	}

	poolBinary, err := filepath.Abs("pool-test")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("InitCommand", func(t *testing.T) {
		cmd := exec.Command(poolBinary, "init", "--pool-size", "2")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("pool init failed: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(string(output), "SUCCESS") {
			t.Errorf("Expected success message in output: %s", output)
		}
	})

	t.Run("StatusCommand", func(t *testing.T) {
		cmd := exec.Command(poolBinary, "status")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("pool status failed: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(string(output), "Pool size: 2") {
			t.Errorf("Expected pool size in output: %s", output)
		}

		if !strings.Contains(string(output), "Available: 2") {
			t.Errorf("Expected available count in output: %s", output)
		}
	})
}

func initTestRepo(dir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}

	testFile := filepath.Join(dir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repo"), 0644); err != nil {
		return err
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "branch", "-M", "main")
	cmd.Dir = dir
	cmd.Run()

	return nil
}
