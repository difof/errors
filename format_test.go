package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/fatih/color"
)

// customFormatter for testing custom implementations
type customFormatter struct {
	config TextConfig // Add config to match other formatters
}

func (f *customFormatter) FormatError(filepath string, message error, inner error) string {
	return "CUSTOM:" + filepath
}

func (f *customFormatter) FormatStack(filepath string, message error) string {
	return "CUSTOM_STACK:" + filepath
}

// NewCustomFormatter creates a new custom formatter
func NewCustomFormatter() Formatter {
	return &customFormatter{config: DefaultTextConfig()}
}

func TestFormatters(t *testing.T) {
	// Test data
	filepath := "test.go:42"
	message := fmt.Errorf("test error")
	inner := NewError("inner.go:24", fmt.Errorf("inner error"), nil)

	t.Run("TextFormatter", func(t *testing.T) {
		// Default config
		formatter := TextFormatter(DefaultTextConfig())
		got := formatter.FormatError(filepath, message, inner)
		want := "at inner.go:24: inner error\n  at test.go:42: test error"
		if got != want {
			t.Errorf("TextFormatter.FormatError() = %q, want %q", got, want)
		}

		// Custom config
		customConfig := TextConfig{Indent: "    "}
		formatter = TextFormatter(customConfig)
		got = formatter.FormatError(filepath, message, inner)
		want = "at inner.go:24: inner error\n    at test.go:42: test error"
		if got != want {
			t.Errorf("TextFormatter.FormatError() with custom config = %q, want %q", got, want)
		}
	})

	t.Run("JSONFormatter", func(t *testing.T) {
		// Default config
		formatter := JSONFormatter(DefaultJSONConfig())
		got := formatter.FormatError(filepath, message, inner)

		// Verify JSON structure
		var stack []jsonError
		if err := json.Unmarshal([]byte(got), &stack); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		if len(stack) != 2 {
			t.Fatalf("Expected 2 errors in stack, got %d", len(stack))
		}

		if stack[0].FilePath != "inner.go:24" {
			t.Errorf("JSON filepath = %v, want inner.go:24", stack[0].FilePath)
		}
		if stack[0].Message != "inner error" {
			t.Errorf("JSON message = %v, want inner error", stack[0].Message)
		}
		if stack[1].FilePath != filepath {
			t.Errorf("JSON inner filepath = %v, want %v", stack[1].FilePath, filepath)
		}
		if stack[1].Message != message.Error() {
			t.Errorf("JSON inner message = %v, want %v", stack[1].Message, message.Error())
		}

		// Custom config
		customConfig := JSONConfig{
			Indent: "    ",
			Prefix: "  ",
		}
		formatter = JSONFormatter(customConfig)
		got = formatter.FormatError(filepath, message, inner)
		if !strings.Contains(got, "    ") {
			t.Error("Custom JSON indent not applied")
		}
	})

	t.Run("YAMLFormatter", func(t *testing.T) {
		// Default config
		formatter := YAMLFormatter(DefaultYAMLConfig())
		got := formatter.FormatError(filepath, message, inner)
		want := []string{
			"errors:",
			"  - filepath: inner.go:24",
			"    message: inner error",
			"  - filepath: test.go:42",
			"    message: test error",
		}
		for _, line := range want {
			if !strings.Contains(got, line) {
				t.Errorf("YAML output missing %q", line)
			}
		}

		// Custom config
		customConfig := YAMLConfig{Indent: "    "}
		formatter = YAMLFormatter(customConfig)
		got = formatter.FormatError(filepath, message, inner)
		if !strings.Contains(got, "    - filepath:") {
			t.Error("Custom YAML indent not applied")
		}
	})

	t.Run("ColoredFormatter", func(t *testing.T) {
		// Save color state and restore after test
		oldNoColor := color.NoColor
		defer func() { color.NoColor = oldNoColor }()
		color.NoColor = false

		// Default config
		formatter := ColoredFormatter(DefaultColorConfig())
		got := formatter.FormatError(filepath, message, inner)

		// Test presence of colored content
		if !strings.Contains(got, "inner.go:24") {
			t.Error("Missing filepath location in output")
		}
		if !strings.Contains(got, "inner error") {
			t.Error("Missing error message in output")
		}
		if !strings.Contains(got, "test.go:42") {
			t.Error("Missing inner filepath in output")
		}
		if !strings.Contains(got, "test error") {
			t.Error("Missing inner message in output")
		}

		// Custom config
		customConfig := ColorConfig{
			SourceColor:  color.New(color.FgGreen),
			MessageColor: color.New(color.FgMagenta),
			InnerColor:   color.New(color.FgCyan),
		}
		formatter = ColoredFormatter(customConfig)
		got = formatter.FormatError(filepath, message, inner)

		// Verify content is present in custom colored output
		if !strings.Contains(got, "inner.go:24") {
			t.Error("Missing filepath in custom colored output")
		}
		if !strings.Contains(got, "inner error") {
			t.Error("Missing message in custom colored output")
		}
	})

	t.Run("CustomFormatter", func(t *testing.T) {
		custom := &customFormatter{config: DefaultTextConfig()}
		got := custom.FormatError(filepath, message, inner)
		want := "CUSTOM:test.go:42"
		if got != want {
			t.Errorf("CustomFormatter = %q, want %q", got, want)
		}
	})

	t.Run("FormatStack", func(t *testing.T) {
		tests := []struct {
			name      string
			formatter Formatter
			want      string
		}{
			{
				name:      "Text",
				formatter: TextFormatter(DefaultTextConfig()),
				want:      "at test.go:42: test error",
			},
			{
				name:      "YAML",
				formatter: YAMLFormatter(DefaultYAMLConfig()),
				want:      "filepath: test.go:42",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.formatter.FormatStack(filepath, message)
				if !strings.Contains(got, tt.want) {
					t.Errorf("FormatStack() = %q, want to contain %q", got, tt.want)
				}
			})
		}
	})
}

func TestSetFormatter(t *testing.T) {
	original := GetFormatter()
	defer SetFormatter(original) // Reset after test

	custom := TextFormatter(TextConfig{Indent: "custom"})
	SetFormatter(custom)

	if GetFormatter() != custom {
		t.Error("SetFormatter did not update the global formatter")
	}

	SetFormatter(nil)
	if GetFormatter() == nil {
		t.Error("SetFormatter(nil) should set default formatter")
	}
}
