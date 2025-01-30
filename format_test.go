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
	inner := fmt.Errorf("inner error")

	t.Run("TextFormatter", func(t *testing.T) {
		// Default config
		formatter := TextFormatter(DefaultTextConfig())
		got := formatter.FormatError(source, message, inner)
		want := "test.go:42: test error"
		if got != want {
			t.Errorf("TextFormatter.FormatError() = %q, want %q", got, want)
		}

		// Custom config
		customConfig := TextConfig{Indent: "    "}
		formatter = TextFormatter(customConfig)
		got = formatter.FormatError(source, message, inner)
		want = "test.go:42: test error"
		if got != want {
			t.Errorf("TextFormatter.FormatError() with custom config = %q, want %q", got, want)
		}
	})

	t.Run("JSONFormatter", func(t *testing.T) {
		// Default config
		formatter := JSONFormatter(DefaultJSONConfig())
		got := formatter.FormatError(source, message, inner)

		// Verify JSON structure
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(got), &data); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		if data["source"] != source {
			t.Errorf("JSON source = %v, want %v", data["source"], source)
		}
		if data["message"] != message.Error() {
			t.Errorf("JSON message = %v, want %v", data["message"], message.Error())
		}
		if data["inner"] != inner.Error() {
			t.Errorf("JSON inner = %v, want %v", data["inner"], inner.Error())
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
			"source: test.go:42",
			"  message: test error",
			"  inner: inner error",
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
		if !strings.Contains(got, "    message:") {
			t.Error("Custom YAML indent not applied")
		}
	})

	t.Run("ColoredFormatter", func(t *testing.T) {
		// Default config
		formatter := ColoredFormatter(DefaultColorConfig())
		got := formatter.FormatError(source, message, inner)
		if !strings.Contains(got, colorBlue) || !strings.Contains(got, colorRed) || !strings.Contains(got, colorYellow) {
			t.Error("Default colors not applied")
		}

		// Custom config
		customConfig := ColorConfig{
			SourceColor:  "\033[32m", // Green
			MessageColor: "\033[35m", // Magenta
			InnerColor:   "\033[36m", // Cyan
		}
		formatter = ColoredFormatter(customConfig)
		got = formatter.FormatError(source, message, inner)
		if !strings.Contains(got, customConfig.SourceColor) ||
			!strings.Contains(got, customConfig.MessageColor) ||
			!strings.Contains(got, customConfig.InnerColor) {
			t.Error("Custom colors not applied")
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
