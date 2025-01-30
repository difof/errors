package errors

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestTextFormatter(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want []string // substrings that should be present in output
	}{
		{
			name: "single error",
			err:  NewError("pkg.func", "file.go", 42, errors.New("test error"), nil),
			want: []string{
				"at pkg.func file.go:42: test error",
			},
		},
		{
			name: "nested errors",
			err: NewError("pkg.func1", "file1.go", 42,
				errors.New("error1"),
				NewError("pkg.func2", "file2.go", 24, errors.New("error2"), nil),
			),
			want: []string{
				"at pkg.func1 file1.go:42: error1",
				"at pkg.func2 file2.go:24: error2",
			},
		},
		{
			name: "nil message",
			err:  NewError("pkg.func", "file.go", 42, nil, nil),
			want: []string{
				"at pkg.func file.go:42",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := TextFormatter(DefaultTextConfig())
			got := formatter.FormatError(tt.err)
			for _, want := range tt.want {
				assert.Contains(t, got, want)
			}
		})
	}

	// Test custom indent
	t.Run("custom indent", func(t *testing.T) {
		config := TextConfig{Indent: "----"}
		formatter := TextFormatter(config)
		err := NewError("pkg.func1", "file1.go", 42,
			errors.New("error1"),
			NewError("pkg.func2", "file2.go", 24, errors.New("error2"), nil),
		)
		got := formatter.FormatError(err)
		assert.Contains(t, got, "\n----at")
	})
}

func TestColoredFormatter(t *testing.T) {
	// Save color state and restore after test
	oldNoColor := color.NoColor
	defer func() { color.NoColor = oldNoColor }()
	color.NoColor = false

	tests := []struct {
		name string
		err  *Error
		want []string // substrings that should be present in output after stripping colors
	}{
		{
			name: "single error",
			err:  NewError("pkg.func", "file.go", 42, errors.New("test error"), nil),
			want: []string{
				"pkg.func",
				"file.go",
				"test error",
			},
		},
		{
			name: "nested errors",
			err: NewError("pkg.func1", "file1.go", 42,
				errors.New("error1"),
				NewError("pkg.func2", "file2.go", 24, errors.New("error2"), nil),
			),
			want: []string{
				"pkg.func1",
				"file1.go",
				"error1",
				"pkg.func2",
				"file2.go",
				"error2",
			},
		},
		{
			name: "nil message",
			err:  NewError("pkg.func", "file.go", 42, nil, nil),
			want: []string{
				"pkg.func",
				"file.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := ColoredFormatter(DefaultColorConfig())
			got := formatter.FormatError(tt.err)
			stripped := stripColors(got)
			for _, want := range tt.want {
				assert.Contains(t, stripped, want)
			}
		})
	}

	// Test custom colors
	t.Run("custom colors", func(t *testing.T) {
		config := ColorConfig{
			SourceColor:  color.New(color.FgGreen),
			MessageColor: color.New(color.FgRed),
			InnerColor:   color.New(color.FgBlue),
		}
		formatter := ColoredFormatter(config)
		err := NewError("pkg.func", "file.go", 42, errors.New("test error"), nil)
		got := formatter.FormatError(err)
		assert.Contains(t, got, "\x1b[32m") // Green color code
		assert.Contains(t, got, "\x1b[31m") // Red color code
	})
}

func TestJSONFormatter(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want []map[string]interface{}
	}{
		{
			name: "single error",
			err:  NewError("pkg.func", "file.go", 42, errors.New("test error"), nil),
			want: []map[string]interface{}{
				{
					"filepath": "file.go",
					"funcpath": "pkg.func",
					"line":     float64(42),
					"message":  "test error",
				},
			},
		},
		{
			name: "nil message",
			err:  NewError("pkg.func", "file.go", 42, nil, nil),
			want: []map[string]interface{}{
				{
					"filepath": "file.go",
					"funcpath": "pkg.func",
					"line":     float64(42),
				},
			},
		},
		{
			name: "nested errors",
			err: NewError("pkg.func1", "file1.go", 42,
				errors.New("error1"),
				NewError("pkg.func2", "file2.go", 24,
					errors.New("error2"),
					NewError("pkg.func3", "file3.go", 12, errors.New("error3"), nil),
				),
			),
			want: []map[string]interface{}{
				{
					"filepath": "file3.go",
					"funcpath": "pkg.func3",
					"line":     float64(12),
					"message":  "error3",
				},
				{
					"filepath": "file2.go",
					"funcpath": "pkg.func2",
					"line":     float64(24),
					"message":  "error2",
				},
				{
					"filepath": "file1.go",
					"funcpath": "pkg.func1",
					"line":     float64(42),
					"message":  "error1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := JSONFormatter(DefaultJSONConfig())
			got := formatter.FormatError(tt.err)

			var parsed []map[string]interface{}
			err := json.Unmarshal([]byte(got), &parsed)
			assert.NoError(t, err)
			assert.Equal(t, len(tt.want), len(parsed))

			// Compare fields
			for i := range tt.want {
				for key, want := range tt.want[i] {
					assert.Equal(t, want, parsed[i][key])
				}
			}
		})
	}

	// Test custom indent and prefix
	t.Run("custom format", func(t *testing.T) {
		config := JSONConfig{
			Indent: "  ",
			Prefix: "    ",
		}
		formatter := JSONFormatter(config)
		err := NewError("pkg.func1", "file1.go", 42,
			errors.New("error1"),
			NewError("pkg.func2", "file2.go", 24,
				errors.New("error2"),
				NewError("pkg.func3", "file3.go", 12, errors.New("error3"), nil),
			),
		)
		got := formatter.FormatError(err)
		assert.Contains(t, got, "    {")
		assert.Contains(t, got, "      \"")

		// Verify order of errors (should be reversed)
		lines := strings.Split(got, "\n")
		var foundLines []string
		for _, line := range lines {
			if strings.Contains(line, "funcpath") {
				foundLines = append(foundLines, line)
			}
		}
		assert.Equal(t, 3, len(foundLines))
		assert.Contains(t, foundLines[0], "pkg.func3")
		assert.Contains(t, foundLines[1], "pkg.func2")
		assert.Contains(t, foundLines[2], "pkg.func1")
	})
}

func TestYAMLFormatter(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want []map[string]interface{}
	}{
		{
			name: "single error",
			err:  NewError("pkg.func", "file.go", 42, errors.New("test error"), nil),
			want: []map[string]interface{}{
				{
					"filepath": "file.go",
					"funcpath": "pkg.func",
					"line":     42,
					"message":  "test error",
				},
			},
		},
		{
			name: "nil message",
			err:  NewError("pkg.func", "file.go", 42, nil, nil),
			want: []map[string]interface{}{
				{
					"filepath": "file.go",
					"funcpath": "pkg.func",
					"line":     42,
				},
			},
		},
		{
			name: "nested errors",
			err: NewError("pkg.func1", "file1.go", 42,
				errors.New("error1"),
				NewError("pkg.func2", "file2.go", 24,
					errors.New("error2"),
					NewError("pkg.func3", "file3.go", 12, errors.New("error3"), nil),
				),
			),
			want: []map[string]interface{}{
				{
					"filepath": "file3.go",
					"funcpath": "pkg.func3",
					"line":     12,
					"message":  "error3",
				},
				{
					"filepath": "file2.go",
					"funcpath": "pkg.func2",
					"line":     24,
					"message":  "error2",
				},
				{
					"filepath": "file1.go",
					"funcpath": "pkg.func1",
					"line":     42,
					"message":  "error1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := YAMLFormatter(DefaultYAMLConfig())
			got := formatter.FormatError(tt.err)

			var parsed []map[string]interface{}
			err := yaml.Unmarshal([]byte(got), &parsed)
			assert.NoError(t, err)
			assert.Equal(t, len(tt.want), len(parsed))

			// Compare fields
			for i := range tt.want {
				for key, want := range tt.want[i] {
					assert.Equal(t, want, parsed[i][key])
				}
			}
		})
	}

	// Test custom indent
	t.Run("custom indent", func(t *testing.T) {
		config := YAMLConfig{
			Indent: "    ",
		}
		formatter := YAMLFormatter(config)
		err := NewError("pkg.func1", "file1.go", 42,
			errors.New("error1"),
			NewError("pkg.func2", "file2.go", 24,
				errors.New("error2"),
				NewError("pkg.func3", "file3.go", 12, errors.New("error3"), nil),
			),
		)
		got := formatter.FormatError(err)
		lines := strings.Split(got, "\n")
		assert.True(t, len(lines) > 0)
		assert.Contains(t, lines[0], "filepath:")

		// Verify order of errors (should be reversed)
		var funcpaths []string
		for _, line := range lines {
			if strings.Contains(line, "funcpath:") {
				funcpaths = append(funcpaths, line)
			}
		}
		assert.Equal(t, 3, len(funcpaths))
		assert.Contains(t, funcpaths[0], "pkg.func3")
		assert.Contains(t, funcpaths[1], "pkg.func2")
		assert.Contains(t, funcpaths[2], "pkg.func1")
	})
}

func TestDefaultConfigs(t *testing.T) {
	t.Run("TextConfig", func(t *testing.T) {
		config := DefaultTextConfig()
		assert.Equal(t, "  ", config.Indent)
	})

	t.Run("JSONConfig", func(t *testing.T) {
		config := DefaultJSONConfig()
		assert.Equal(t, "  ", config.Indent)
		assert.Equal(t, "", config.Prefix)
	})

	t.Run("YAMLConfig", func(t *testing.T) {
		config := DefaultYAMLConfig()
		assert.Equal(t, "  ", config.Indent)
	})

	t.Run("ColorConfig", func(t *testing.T) {
		config := DefaultColorConfig()
		assert.NotNil(t, config.SourceColor)
		assert.NotNil(t, config.MessageColor)
		assert.NotNil(t, config.InnerColor)
	})
}

func TestSetGetFormatter(t *testing.T) {
	// Save default formatter
	defaultF := GetFormatter()
	defer SetFormatter(defaultF)

	// Test setting custom formatter
	customF := TextFormatter(TextConfig{Indent: "    "})
	SetFormatter(customF)
	assert.Equal(t, customF, GetFormatter())

	// Test setting nil formatter (should revert to default)
	SetFormatter(nil)
	assert.Equal(t, defaultFormatter, GetFormatter())

	// Test DefaultFormatter function
	assert.Equal(t, TextFormatter(DefaultTextConfig()), DefaultFormatter())
}

// stripColors removes ANSI color codes from a string
func stripColors(s string) string {
	var b strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
