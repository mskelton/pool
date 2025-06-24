package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRepository(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	if err := initTestRepo(tmpDir); err != nil {
		t.Fatal(err)
	}
	
	repo, err := NewRepository(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	
	if repo.Path != tmpDir {
		t.Errorf("Expected path %s, got %s", tmpDir, repo.Path)
	}
	
	if repo.IsBare {
		t.Error("Expected non-bare repository")
	}
	
	topLevel, err := repo.GetTopLevel()
	if err != nil {
		t.Fatal(err)
	}
	
	resolvedTmpDir, _ := filepath.EvalSymlinks(tmpDir)
	resolvedTopLevel, _ := filepath.EvalSymlinks(topLevel)
	
	if resolvedTopLevel != resolvedTmpDir {
		t.Errorf("Expected top level %s, got %s", resolvedTmpDir, resolvedTopLevel)
	}
	
	if repo.IsInWorktree() {
		t.Error("Expected not to be in worktree")
	}
	
	if repo.HasUncommittedChanges() {
		t.Error("Expected no uncommitted changes")
	}
	
	testFile := filepath.Join(tmpDir, "test2.txt")
	if err := os.WriteFile(testFile, []byte("test2"), 0644); err != nil {
		t.Fatal(err)
	}
	
	cmd := exec.Command("git", "add", "test2.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	
	if !repo.HasUncommittedChanges() {
		t.Error("Expected uncommitted changes")
	}
}

func TestWorktrees(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "worktree-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	if err := initTestRepo(tmpDir); err != nil {
		t.Fatal(err)
	}
	
	repo, err := NewRepository(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	
	worktrees, err := repo.ListWorktrees()
	if err != nil {
		t.Fatal(err)
	}
	
	if len(worktrees) != 1 {
		t.Fatalf("Expected 1 worktree, got %d", len(worktrees))
	}
	
	worktreePath := filepath.Join(tmpDir, "feature-branch")
	if err := repo.AddWorktree(worktreePath, "feature"); err != nil {
		t.Fatal(err)
	}
	
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		t.Error("Worktree directory was not created")
	}
	
	worktrees, err = repo.ListWorktrees()
	if err != nil {
		t.Fatal(err)
	}
	
	if len(worktrees) != 2 {
		t.Fatalf("Expected 2 worktrees, got %d", len(worktrees))
	}
	
	var found bool
	for _, wt := range worktrees {
		if wt.Branch == "feature" {
			found = true
			resolvedWorktreePath, _ := filepath.EvalSymlinks(worktreePath)
			resolvedWtPath, _ := filepath.EvalSymlinks(wt.Path)
			if resolvedWtPath != resolvedWorktreePath {
				t.Errorf("Expected worktree path %s, got %s", resolvedWorktreePath, resolvedWtPath)
			}
		}
	}
	
	if !found {
		t.Error("New worktree not found in list")
	}
	
	if err := repo.RemoveWorktree(worktreePath); err != nil {
		t.Fatal(err)
	}
	
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("Worktree directory was not removed")
	}
}

func TestBranches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "branch-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	if err := initTestRepo(tmpDir); err != nil {
		t.Fatal(err)
	}
	
	repo, err := NewRepository(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	
	if !repo.BranchExists("main") {
		t.Error("Expected main branch to exist")
	}
	
	if repo.BranchExists("nonexistent") {
		t.Error("Expected nonexistent branch to not exist")
	}
	
	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		t.Fatal(err)
	}
	
	if currentBranch != "main" {
		t.Errorf("Expected current branch to be main, got %s", currentBranch)
	}
	
	cmd := exec.Command("git", "checkout", "-b", "test-branch")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	
	if !repo.BranchExists("test-branch") {
		t.Error("Expected test-branch to exist")
	}
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
	
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
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