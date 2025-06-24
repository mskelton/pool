package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLoggerOutput(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tests := []struct {
		name     string
		logFunc  func(string, ...interface{})
		message  string
		contains string
	}{
		{
			name:     "Info message",
			logFunc:  Info,
			message:  "test info",
			contains: "[INFO] test info",
		},
		{
			name:     "Success message",
			logFunc:  Success,
			message:  "test success",
			contains: "[SUCCESS] test success",
		},
		{
			name:     "Warning message",
			logFunc:  Warning,
			message:  "test warning",
			contains: "[WARNING] test warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc(tt.message)
			
			// Restore stdout and read output
			w.Close()
			os.Stdout = oldStdout
			
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()
			
			if !strings.Contains(output, tt.contains) {
				t.Errorf("Expected output to contain %q, got %q", tt.contains, output)
			}
		})
		
		// Reset for next test
		r, w, _ = os.Pipe()
		os.Stdout = w
	}
	
	// Final cleanup
	w.Close()
	os.Stdout = oldStdout
}

func TestLoggerFormatting(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	Info("Hello %s, you have %d messages", "Alice", 5)
	
	w.Close()
	os.Stdout = oldStdout
	
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	
	expected := "Hello Alice, you have 5 messages"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain %q, got %q", expected, output)
	}
}