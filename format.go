package errors

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

// Formatter formats an ErrorChain into a detailed representation that can
// include stack/source metadata.
type Formatter interface {
	// FormatError formats the full error chain.
	FormatError(err *ErrorChain) string
}

// TextConfig configures the text formatter
type TextConfig struct {
	// Indent is the indentation string used for nested errors
	Indent string
}

// textFormatter is the default detailed formatter.
type textFormatter struct {
	config TextConfig
}

func (f *textFormatter) FormatError(err *ErrorChain) string {
	entries := Collapse(err)

	// Build the output with proper indentation
	var b strings.Builder
	for i := 0; i < len(entries); i++ {
		if i > 0 {
			b.WriteString("\n")
			b.WriteString(f.config.Indent)
		}
		b.WriteString(f.createStackEntry(&entries[i]))
	}

	return b.String()
}

func (f *textFormatter) createStackEntry(err *ErrorEntry) string {
	var b strings.Builder

	hasDetails := err.FuncPath != "" && err.FilePath != "" && err.Line != 0

	if hasDetails {
		b.WriteString("at ")
	}

	if err.FuncPath != "" {
		b.WriteString(err.FuncPath)
		b.WriteString(" ")
	}

	if err.FilePath != "" {
		b.WriteString(err.FilePath)
	}

	if err.Line != 0 {
		b.WriteString(":")
		b.WriteString(strconv.Itoa(err.Line))
	}

	if err.Message != "" {
		if hasDetails {
			b.WriteString(": ")
		} else {
			b.WriteString("error: ")
		}

		b.WriteString(err.Message)
	}

	return b.String()
}

// ColorConfig configures the colored formatter
type ColorConfig struct {
	// SourceColor is the color for source locations
	SourceColor *color.Color
	// MessageColor is the color for error messages
	MessageColor *color.Color
	// InnerColor is the color for inner errors
	InnerColor *color.Color
}

// coloredFormatter formats detailed errors with colors for terminal output.
type coloredFormatter struct {
	config ColorConfig
}

func (f *coloredFormatter) FormatError(err *ErrorChain) string {
	entries := Collapse(err)

	// Build colored output
	var b strings.Builder
	for i, err := range entries {
		if i > 0 {
			b.WriteString("\n  ")
		}
		b.WriteString(f.createStackEntry(&err))
	}

	return b.String()
}

func (f *coloredFormatter) createStackEntry(err *ErrorEntry) string {
	var b strings.Builder

	hasDetails := err.FuncPath != "" && err.FilePath != "" && err.Line != 0
	if hasDetails {
		b.WriteString("at ")
	}

	if err.FuncPath != "" {
		b.WriteString(f.config.InnerColor.Sprint(err.FuncPath))
		b.WriteString(" ")
	}

	if err.FilePath != "" {
		b.WriteString(f.config.SourceColor.Sprint(err.FilePath))
	}

	if err.Line != 0 {
		b.WriteString(":")
		b.WriteString(strconv.Itoa(err.Line))
	}

	if err.Message != "" {
		if hasDetails {
			b.WriteString(": ")
		} else {
			b.WriteString(f.config.InnerColor.Sprint("error: "))
		}
		b.WriteString(f.config.MessageColor.Sprint(err.Message))
	}

	return b.String()
}

// JSONConfig configures the JSON formatter
type JSONConfig struct {
	// Indent is the indentation string used for pretty printing
	Indent string
	// Prefix is the prefix used for each line in pretty printed output
	Prefix string
}

// jsonFormatter formats detailed errors as JSON objects.
type jsonFormatter struct {
	config JSONConfig
}

func (f *jsonFormatter) FormatError(err *ErrorChain) string {
	entries := Collapse(err)

	var data []byte
	if f.config.Indent == "" {
		data, _ = json.Marshal(entries)
	} else {
		data, _ = json.MarshalIndent(entries, f.config.Prefix, f.config.Indent)
	}

	return string(data)
}

// YAMLConfig configures the YAML formatter
type YAMLConfig struct {
	// Indent is the indentation string used for nested levels
	Indent string
}

// yamlFormatter formats detailed errors as YAML documents.
type yamlFormatter struct {
	config YAMLConfig
}

func (f *yamlFormatter) FormatError(err *ErrorChain) string {
	entries := Collapse(err)

	data, _ := yaml.Marshal(entries)

	return string(data)
}

var (
	// defaultFormatter is the text formatter used by default
	defaultFormatter = TextFormatter(DefaultTextConfig())
	// currentFormatter holds the current global formatter
	currentFormatter = defaultFormatter
	// formatterMutex protects access to currentFormatter
	formatterMutex sync.RWMutex
)

func init() {
	currentFormatter = defaultFormatter
}

// SetFormatter sets a custom formatter for all new errors
func SetFormatter(f Formatter) {
	if f == nil {
		f = defaultFormatter
	}
	formatterMutex.Lock()
	currentFormatter = f
	formatterMutex.Unlock()
}

// GetFormatter returns the current global formatter
func GetFormatter() Formatter {
	formatterMutex.RLock()
	defer formatterMutex.RUnlock()
	return currentFormatter
}

// DefaultTextConfig returns the default configuration for text formatter
func DefaultTextConfig() TextConfig {
	return TextConfig{
		Indent: "  ",
	}
}

// DefaultJSONConfig returns the default configuration for JSON formatter
func DefaultJSONConfig() JSONConfig {
	return JSONConfig{
		Indent: "  ",
		Prefix: "",
	}
}

// DefaultYAMLConfig returns the default configuration for YAML formatter
func DefaultYAMLConfig() YAMLConfig {
	return YAMLConfig{
		Indent: "  ",
	}
}

// DefaultColorConfig returns the default configuration for colored formatter
func DefaultColorConfig() ColorConfig {
	return ColorConfig{
		SourceColor:  color.New(color.FgBlue),
		MessageColor: color.New(color.FgRed),
		InnerColor:   color.New(color.FgYellow),
	}
}

// TextFormatter returns a new detailed text formatter instance with custom configuration.
func TextFormatter(config TextConfig) Formatter {
	return &textFormatter{config: config}
}

// JSONFormatter returns a new detailed JSON formatter instance with custom configuration.
func JSONFormatter(config JSONConfig) Formatter {
	return &jsonFormatter{config: config}
}

// YAMLFormatter returns a new detailed YAML formatter instance with custom configuration.
func YAMLFormatter(config YAMLConfig) Formatter {
	return &yamlFormatter{config: config}
}

// ColoredFormatter returns a new detailed colored formatter instance with custom configuration.
func ColoredFormatter(config ColorConfig) Formatter {
	return &coloredFormatter{config: config}
}

// DefaultFormatter returns the default detailed text formatter with default configuration.
func DefaultFormatter() Formatter {
	return TextFormatter(DefaultTextConfig())
}
