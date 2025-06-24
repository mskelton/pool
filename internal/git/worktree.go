package git

import (
	"bufio"
	"strings"
)

type Worktree struct {
	Path   string
	Branch string
	Commit string
	Bare   bool
}

func (r *Repository) ListWorktrees() ([]Worktree, error) {
	output, err := r.output("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	var worktrees []Worktree
	var current Worktree

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "worktree ") {
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		} else if strings.HasPrefix(line, "branch refs/heads/") {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		} else if strings.HasPrefix(line, "HEAD ") {
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		} else if line == "bare" {
			current.Bare = true
		}
	}

	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

func (r *Repository) AddWorktree(path, branch string, opts ...string) error {
	args := []string{"worktree", "add"}
	args = append(args, opts...)
	args = append(args, path, "-b", branch, r.DefaultBranch)
	return r.run(args...)
}

func (r *Repository) AddWorktreeFromBranch(path, branch, source string) error {
	return r.run("worktree", "add", path, "-b", branch, source)
}

func (r *Repository) RemoveWorktree(path string) error {
	return r.run("worktree", "remove", path)
}

func (r *Repository) MoveWorktree(from, to string) error {
	return r.run("worktree", "move", from, to)
}

func (r *Repository) RepairWorktrees() error {
	return r.run("worktree", "repair")
}

func (r *Repository) PruneWorktrees() error {
	return r.run("worktree", "prune")
}

func CheckoutBranch(branch string, create bool) error {
	if create {
		return RunInDir(".", "checkout", "-B", branch)
	}
	return RunInDir(".", "checkout", branch)
}

func CheckoutNewBranch(branch, source string) error {
	return RunInDir(".", "checkout", "-B", branch, source)
}
