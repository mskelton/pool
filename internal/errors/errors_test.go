package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestOperationError(t *testing.T) {
	baseErr := errors.New("base error")
	opErr := NewOperationError("create worktree", baseErr)

	if !strings.Contains(opErr.Error(), "create worktree") {
		t.Errorf("Expected error to contain operation name")
	}

	if !strings.Contains(opErr.Error(), "base error") {
		t.Errorf("Expected error to contain base error message")
	}

	// Test unwrap
	var unwrapped error
	if err, ok := opErr.(*OperationError); ok {
		unwrapped = err.Unwrap()
	}

	if unwrapped != baseErr {
		t.Errorf("Expected unwrapped error to be base error")
	}
}

func TestGitError(t *testing.T) {
	baseErr := errors.New("command failed")
	gitErr := NewGitError("checkout main", baseErr, "fatal: branch not found")

	errStr := gitErr.Error()

	if !strings.Contains(errStr, "git checkout main failed") {
		t.Errorf("Expected error to contain git command")
	}

	if !strings.Contains(errStr, "fatal: branch not found") {
		t.Errorf("Expected error to contain output")
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		msg      string
		expected string
	}{
		{
			name:     "With value",
			field:    "branch",
			value:    "main/feature",
			msg:      "contains invalid characters",
			expected: "validation failed for branch=main/feature: contains invalid characters",
		},
		{
			name:     "Without value",
			field:    "pool-size",
			value:    "",
			msg:      "must be positive",
			expected: "validation failed for pool-size: must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.value, tt.msg)
			if err.Error() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, err.Error())
			}
		})
	}
}

func TestWrap(t *testing.T) {
	// Test nil error
	if Wrap(nil, "context") != nil {
		t.Errorf("Expected Wrap(nil) to return nil")
	}

	// Test wrapping
	baseErr := errors.New("base error")
	wrapped := Wrap(baseErr, "failed to create")

	if !strings.Contains(wrapped.Error(), "failed to create") {
		t.Errorf("Expected wrapped error to contain context")
	}

	if !strings.Contains(wrapped.Error(), "base error") {
		t.Errorf("Expected wrapped error to contain base error")
	}
}

func TestWrapf(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrapf(baseErr, "failed to create %s for %d items", "worktree", 5)

	expected := "failed to create worktree for 5 items"
	if !strings.Contains(wrapped.Error(), expected) {
		t.Errorf("Expected wrapped error to contain %q", expected)
	}
}
