package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotGitRepo      = errors.New("not a git repository")
	ErrAlreadyBare     = errors.New("repository is already bare")
	ErrInWorktree      = errors.New("cannot perform operation from within a worktree")
	ErrUncommitted     = errors.New("uncommitted changes present")
	ErrNoPoolAvailable = errors.New("no available pool worktrees")
	ErrPoolNotFound    = errors.New("worktree pool not found")
	ErrWorktreeExists  = errors.New("worktree already exists")
	ErrBranchNotFound  = errors.New("branch not found")
)

type OperationError struct {
	Op  string
	Err error
}

func (e *OperationError) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *OperationError) Unwrap() error {
	return e.Err
}

func NewOperationError(op string, err error) error {
	return &OperationError{Op: op, Err: err}
}

type GitError struct {
	Command string
	Output  string
	Err     error
}

func (e *GitError) Error() string {
	if e.Output != "" {
		return fmt.Sprintf("git %s failed: %v\nOutput: %s", e.Command, e.Err, e.Output)
	}
	return fmt.Sprintf("git %s failed: %v", e.Command, e.Err)
}

func (e *GitError) Unwrap() error {
	return e.Err
}

func NewGitError(command string, err error, output string) error {
	return &GitError{
		Command: command,
		Output:  output,
		Err:     err,
	}
}

type ValidationError struct {
	Field string
	Value string
	Msg   string
}

func (e *ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("validation failed for %s=%s: %s", e.Field, e.Value, e.Msg)
	}
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Msg)
}

func NewValidationError(field, value, msg string) error {
	return &ValidationError{
		Field: field,
		Value: value,
		Msg:   msg,
	}
}

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}