package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CloneBare(url, name string) error {
	cmd := exec.Command("git", "clone", "--bare", url, name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ConfigureBareRepo(repoPath string) error {
	cmd := exec.Command("git", "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure fetch refs: %w", err)
	}

	cmd = exec.Command("git", "fetch", "origin")
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func GetDefaultBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "main", nil
	}

	branch := strings.TrimSpace(string(output))
	return strings.TrimPrefix(branch, "refs/remotes/origin/"), nil
}

func ConvertToBare(originalPath string) (string, error) {
	repoName := filepath.Base(originalPath)
	parentDir := filepath.Dir(originalPath)
	bareDir := filepath.Join(parentDir, repoName+".git")

	if _, err := os.Stat(bareDir); err == nil {
		return "", fmt.Errorf("directory %s already exists", bareDir)
	}

	cmd := exec.Command("git", "clone", "--bare", originalPath, bareDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create bare clone: %w", err)
	}

	return bareDir, nil
}
