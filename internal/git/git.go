package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mskelton/pool/internal/errors"
)

type Repository struct {
	Path          string
	IsBare        bool
	DefaultBranch string
}

func NewRepository(path string) (*Repository, error) {
	repo := &Repository{Path: path}

	if err := repo.run("rev-parse", "--git-dir"); err != nil {
		return nil, errors.ErrNotGitRepo
	}

	output, err := repo.output("rev-parse", "--is-bare-repository")
	if err == nil {
		repo.IsBare = strings.TrimSpace(output) == "true"
	}

	output, err = repo.output("symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		repo.DefaultBranch = strings.TrimPrefix(strings.TrimSpace(output), "refs/remotes/origin/")
	} else {
		repo.DefaultBranch = "main"
	}

	return repo, nil
}

func (r *Repository) GetTopLevel() (string, error) {
	output, err := r.output("rev-parse", "--show-toplevel")
	if err != nil {
		output, err = r.output("rev-parse", "--git-dir")
		if err != nil {
			return "", err
		}
		return filepath.Dir(strings.TrimSpace(output)), nil
	}
	return strings.TrimSpace(output), nil
}

func (r *Repository) IsInWorktree() bool {
	gitDir, _ := r.output("rev-parse", "--git-dir")
	commonDir, _ := r.output("rev-parse", "--git-common-dir")
	return strings.TrimSpace(gitDir) != strings.TrimSpace(commonDir)
}

func (r *Repository) HasUncommittedChanges() bool {
	err := r.run("diff-index", "--quiet", "HEAD", "--")
	return err != nil
}

func (r *Repository) BranchExists(branch string) bool {
	err := r.run("show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", branch))
	return err == nil
}

func (r *Repository) RemoteBranchExists(branch string) bool {
	output, err := r.output("ls-remote", "--heads", "origin", branch)
	return err == nil && strings.Contains(output, branch)
}

func (r *Repository) GetCurrentBranch() (string, error) {
	output, err := r.output("branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (r *Repository) GetMergedBranches() ([]string, error) {
	output, err := r.output("branch", "-r", "--merged", fmt.Sprintf("origin/%s", r.DefaultBranch))
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	branches := make([]string, 0, len(lines))

	for _, line := range lines {
		branch := strings.TrimSpace(line)
		if !strings.Contains(branch, r.DefaultBranch) && strings.HasPrefix(branch, "origin/") {
			branches = append(branches, strings.TrimPrefix(branch, "origin/"))
		}
	}

	return branches, nil
}

func (r *Repository) FetchOrigin() error {
	return r.run("fetch", "origin")
}

func (r *Repository) FetchOriginPrune() error {
	return r.run("fetch", "origin", "--prune")
}

func (r *Repository) run(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Path

	if err := cmd.Run(); err != nil {
		return errors.NewGitError(strings.Join(args, " "), err, "")
	}
	return nil
}

func (r *Repository) output(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Path

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", errors.NewGitError(strings.Join(args, " "), err, stderr.String())
	}

	return stdout.String(), nil
}

func RunInDir(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	if err := cmd.Run(); err != nil {
		return errors.NewGitError(strings.Join(args, " "), err, "")
	}
	return nil
}
