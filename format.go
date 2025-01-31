package errors

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

// Formatter interface for custom formatting
type Formatter interface {
	// FormatError formats a single error node
	FormatError(err *Error) string
}

// TextConfig configures the text formatter
type TextConfig struct {
	// Indent is the indentation string used for nested errors
	Indent string
}

// textFormatter is the default formatter that maintains the original error format
type textFormatter struct {
	config TextConfig
}

func (f *textFormatter) FormatError(err *Error) string {
	entries := err.ExtractEntries()

	// Build the output with proper indentation
	// Iterate in reverse to show root cause first
	var b strings.Builder
	for i := len(entries) - 1; i >= 0; i-- {
		if i < len(entries)-1 {
			b.WriteString("\n")
			b.WriteString(f.config.Indent)
		}
		b.WriteString(f.createStackEntry(&entries[i]))
	}

	return b.String()
}

func (f *textFormatter) createStackEntry(err *Error) string {
	var b strings.Builder

	hasDetails := err.FuncPath != "" && err.FilePath != NO_SOURCE && err.Line != 0
	if hasDetails {
		b.WriteString("at ")
	}

	if err.FuncPath != "" {
		b.WriteString(err.FuncPath)
		b.WriteString(" ")
	}

	if err.FilePath != NO_SOURCE {
		b.WriteString(err.FilePath)
		b.WriteString(":")
		b.WriteString(strconv.Itoa(err.Line))
	}

	if err.Message != nil {
		if hasDetails {
			b.WriteString(": ")
		} else {
			b.WriteString("caught error: ")
		}

		b.WriteString(err.MessageString)
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

// coloredFormatter formats errors with colors for terminal output
type coloredFormatter struct {
	config ColorConfig
}

func (f *coloredFormatter) FormatError(err *Error) string {
	entries := err.ExtractEntries()

	// Reverse to show root cause first
	for i := 0; i < len(entries)/2; i++ {
		j := len(entries) - 1 - i
		entries[i], entries[j] = entries[j], entries[i]
	}

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

func (f *coloredFormatter) createStackEntry(err *Error) string {
	var b strings.Builder

	hasDetails := err.FuncPath != "" && err.FilePath != NO_SOURCE && err.Line != 0
	if hasDetails {
		b.WriteString("at ")
	}

	if err.FuncPath != "" {
		b.WriteString(f.config.InnerColor.Sprint(err.FuncPath))
		b.WriteString(" ")
	}

	if err.FilePath != NO_SOURCE {
		b.WriteString(f.config.SourceColor.Sprint(err.FilePath))
		b.WriteString(":")
		b.WriteString(strconv.Itoa(err.Line))
	}

	if err.MessageString != "" {
		if hasDetails {
			b.WriteString(": ")
		} else {
			b.WriteString(f.config.InnerColor.Sprint("caught error: "))
		}
		b.WriteString(f.config.MessageColor.Sprint(err.MessageString))
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

// jsonFormatter formats errors as JSON objects
type jsonFormatter struct {
	config JSONConfig
}

func (f *jsonFormatter) FormatError(err *Error) string {
	entries := err.ExtractEntries()

	// Reverse to show root cause first
	for i := 0; i < len(entries)/2; i++ {
		j := len(entries) - 1 - i
		entries[i], entries[j] = entries[j], entries[i]
	}

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

// yamlFormatter formats errors as YAML documents
type yamlFormatter struct {
	config YAMLConfig
}

func (f *yamlFormatter) FormatError(err *Error) string {
	entries := err.ExtractEntries()

	// Reverse to show root cause first
	for i := 0; i < len(entries)/2; i++ {
		j := len(entries) - 1 - i
		entries[i], entries[j] = entries[j], entries[i]
	}

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

// TextFormatter returns a new text formatter instance with custom configuration
func TextFormatter(config TextConfig) Formatter {
	return &textFormatter{config: config}
}

// JSONFormatter returns a new JSON formatter instance with custom configuration
func JSONFormatter(config JSONConfig) Formatter {
	return &jsonFormatter{config: config}
}

// YAMLFormatter returns a new YAML formatter instance with custom configuration
func YAMLFormatter(config YAMLConfig) Formatter {
	return &yamlFormatter{config: config}
}

// ColoredFormatter returns a new colored formatter instance with custom configuration
func ColoredFormatter(config ColorConfig) Formatter {
	return &coloredFormatter{config: config}
}

// DefaultFormatter returns the default text formatter with default configuration
func DefaultFormatter() Formatter {
	return TextFormatter(DefaultTextConfig())
}
