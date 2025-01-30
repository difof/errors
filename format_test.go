package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// customFormatter for testing custom implementations
type customFormatter struct {
	config TextConfig // Add config to match other formatters
}

func (f *customFormatter) FormatError(source string, message error, inner error) string {
	return "CUSTOM:" + source
}

func (f *customFormatter) FormatStack(source string, message error) string {
	return "CUSTOM_STACK:" + source
}

// NewCustomFormatter creates a new custom formatter
func NewCustomFormatter() Formatter {
	return &customFormatter{config: DefaultTextConfig()}
}

func TestFormatters(t *testing.T) {
	// Test data
	source := "test.go:42"
	message := fmt.Errorf("test error")
	inner := NewError("inner.go:24", fmt.Errorf("inner error"), nil)

	t.Run("TextFormatter", func(t *testing.T) {
		// Default config
		formatter := TextFormatter(DefaultTextConfig())
		got := formatter.FormatError(source, message, inner)
		want := "test.go:42: test error\n  inner.go:24: inner error"
		if got != want {
			t.Errorf("TextFormatter.FormatError() = %q, want %q", got, want)
		}

		// Custom config
		customConfig := TextConfig{Indent: "    "}
		formatter = TextFormatter(customConfig)
		got = formatter.FormatError(source, message, inner)
		want = "test.go:42: test error\n    inner.go:24: inner error"
		if got != want {
			t.Errorf("TextFormatter.FormatError() with custom config = %q, want %q", got, want)
		}
	})

	t.Run("JSONFormatter", func(t *testing.T) {
		// Default config
		formatter := JSONFormatter(DefaultJSONConfig())
		got := formatter.FormatError(source, message, inner)

		// Verify JSON structure
		var stack []jsonError
		if err := json.Unmarshal([]byte(got), &stack); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		if len(stack) != 2 {
			t.Fatalf("Expected 2 errors in stack, got %d", len(stack))
		}

		if stack[0].Source != source {
			t.Errorf("JSON source = %v, want %v", stack[0].Source, source)
		}
		if stack[0].Message != message.Error() {
			t.Errorf("JSON message = %v, want %v", stack[0].Message, message.Error())
		}
		if stack[1].Source != "inner.go:24" {
			t.Errorf("JSON inner source = %v, want inner.go:24", stack[1].Source)
		}
		if stack[1].Message != "inner error" {
			t.Errorf("JSON inner message = %v, want inner error", stack[1].Message)
		}

		// Custom config
		customConfig := JSONConfig{
			Indent: "    ",
			Prefix: "  ",
		}
		formatter = JSONFormatter(customConfig)
		got = formatter.FormatError(source, message, inner)
		if !strings.Contains(got, "    ") {
			t.Error("Custom JSON indent not applied")
		}
	})

	t.Run("YAMLFormatter", func(t *testing.T) {
		// Default config
		formatter := YAMLFormatter(DefaultYAMLConfig())
		got := formatter.FormatError(source, message, inner)
		want := []string{
			"errors:",
			"  - source: test.go:42",
			"    message: test error",
			"  - source: inner.go:24",
			"    message: inner error",
		}
		for _, line := range want {
			if !strings.Contains(got, line) {
				t.Errorf("YAML output missing %q", line)
			}
		}

		// Custom config
		customConfig := YAMLConfig{Indent: "    "}
		formatter = YAMLFormatter(customConfig)
		got = formatter.FormatError(source, message, inner)
		if !strings.Contains(got, "    - source:") {
			t.Error("Custom YAML indent not applied")
		}
	})

	t.Run("ColoredFormatter", func(t *testing.T) {
		// Default config
		formatter := ColoredFormatter(DefaultColorConfig())
		got := formatter.FormatError(source, message, inner)

		// Verify color codes and structure
		if !strings.Contains(got, colorBlue+source+colorReset) {
			t.Error("Source not properly colored")
		}
		if !strings.Contains(got, colorRed+message.Error()+colorReset) {
			t.Error("Message not properly colored")
		}
		if !strings.Contains(got, "  "+colorBlue+"inner.go:24"+colorReset) {
			t.Error("Inner source not properly colored and indented")
		}
		if !strings.Contains(got, colorRed+"inner error"+colorReset) {
			t.Error("Inner message not properly colored")
		}

		// Custom config
		customConfig := ColorConfig{
			SourceColor:  "\033[32m", // Green
			MessageColor: "\033[35m", // Magenta
			InnerColor:   "\033[36m", // Cyan
		}
		formatter = ColoredFormatter(customConfig)
		got = formatter.FormatError(source, message, inner)
		if !strings.Contains(got, customConfig.SourceColor+source+colorReset) {
			t.Error("Custom source color not applied")
		}
		if !strings.Contains(got, customConfig.MessageColor+message.Error()+colorReset) {
			t.Error("Custom message color not applied")
		}
	})

	t.Run("CustomFormatter", func(t *testing.T) {
		custom := &customFormatter{config: DefaultTextConfig()}
		got := custom.FormatError(source, message, inner)
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
				want:      "test.go:42: test error",
			},
			{
				name:      "YAML",
				formatter: YAMLFormatter(DefaultYAMLConfig()),
				want:      "source: test.go:42",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.formatter.FormatStack(source, message)
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
