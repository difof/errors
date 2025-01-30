package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Formatter interface for custom formatting
type Formatter interface {
	// FormatError formats a single error node
	FormatError(source string, message error, inner error) string
	// FormatStack formats a single node in a stack trace
	FormatStack(source string, message error) string
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

func (f *textFormatter) FormatError(source string, message error, inner error) string {
	var stack []string

	// Build the stack in reverse order (innermost first)
	var current error = inner
	var e *Error
	for current != nil {
		if As(current, &e) {
			stack = append([]string{fmt.Sprintf("%s: %s", e.Source, e.Message.Error())}, stack...)
			current = e.Inner
		} else {
			stack = append([]string{current.Error()}, stack...)
			break
		}
	}

	// Add the current error last
	stack = append(stack, fmt.Sprintf("%s: %s", source, message.Error()))

	// Build the output with proper indentation
	var b strings.Builder
	for i, err := range stack {
		if i > 0 {
			b.WriteString("\n")
			b.WriteString(f.config.Indent)
		}
		b.WriteString(err)
	}

	return b.String()
}

func (f *textFormatter) FormatStack(source string, message error) string {
	if message == nil {
		return source
	}
	return fmt.Sprintf("%s: %s", source, message.Error())
}

// JSONConfig configures the JSON formatter
type JSONConfig struct {
	// Indent is the indentation string used for pretty printing
	Indent string
	// Prefix is the prefix used for each line in pretty printed output
	Prefix string
}

type jsonError struct {
	Source  string `json:"source"`
	Message string `json:"message,omitempty"`
}

// jsonFormatter formats errors as JSON objects
type jsonFormatter struct {
	config JSONConfig
}

func (f *jsonFormatter) FormatError(source string, message error, inner error) string {
	var stack []jsonError

	// Build the stack in reverse order (innermost first)
	var current error = inner
	var e *Error
	for current != nil {
		if As(current, &e) {
			stack = append([]jsonError{{
				Source:  e.Source,
				Message: e.Message.Error(),
			}}, stack...)
			current = e.Inner
		} else {
			stack = append([]jsonError{{
				Message: current.Error(),
			}}, stack...)
			break
		}
	}

	// Add the current error last
	stack = append(stack, jsonError{
		Source:  source,
		Message: message.Error(),
	})

	data, _ := json.MarshalIndent(stack, f.config.Prefix, f.config.Indent)
	return string(data)
}

func (f *jsonFormatter) FormatStack(source string, message error) string {
	entry := jsonError{Source: source}
	if message != nil {
		entry.Message = message.Error()
	}
	data, _ := json.MarshalIndent(entry, f.config.Prefix, f.config.Indent)
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

func (f *yamlFormatter) FormatError(source string, message error, inner error) string {
	var stack []struct {
		Source  string
		Message string
	}

	// Build the stack in reverse order (innermost first)
	var current error = inner
	var e *Error
	for current != nil {
		if As(current, &e) {
			stack = append([]struct {
				Source  string
				Message string
			}{{
				Source:  e.Source,
				Message: e.Message.Error(),
			}}, stack...)
			current = e.Inner
		} else {
			stack = append([]struct {
				Source  string
				Message string
			}{{
				Message: current.Error(),
			}}, stack...)
			break
		}
	}

	// Add the current error last
	stack = append(stack, struct {
		Source  string
		Message string
	}{
		Source:  source,
		Message: message.Error(),
	})

	// Build YAML output
	var b strings.Builder
	b.WriteString("errors:\n")
	for _, err := range stack {
		b.WriteString(f.config.Indent)
		b.WriteString("- source: ")
		b.WriteString(err.Source)
		if err.Message != "" {
			b.WriteString("\n")
			b.WriteString(f.config.Indent)
			b.WriteString("  message: ")
			b.WriteString(err.Message)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (f *yamlFormatter) FormatStack(source string, message error) string {
	var b strings.Builder
	b.WriteString("source: ")
	b.WriteString(source)
	if message != nil {
		b.WriteString("\n")
		b.WriteString(f.config.Indent)
		b.WriteString("message: ")
		b.WriteString(message.Error())
	}
	return b.String()
}

// ANSI color codes
const (
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m"
)

// ColorConfig configures the colored formatter
type ColorConfig struct {
	// SourceColor is the ANSI color code for source locations
	SourceColor string
	// MessageColor is the ANSI color code for error messages
	MessageColor string
	// InnerColor is the ANSI color code for inner errors
	InnerColor string
}

// coloredFormatter formats errors with ANSI colors for terminal output
type coloredFormatter struct {
	config ColorConfig
}

func (f *coloredFormatter) FormatError(source string, message error, inner error) string {
	var stack []struct {
		Source  string
		Message string
	}

	// Build the stack in reverse order (innermost first)
	var current error = inner
	var e *Error
	for current != nil {
		if As(current, &e) {
			stack = append([]struct {
				Source  string
				Message string
			}{{
				Source:  e.Source,
				Message: e.Message.Error(),
			}}, stack...)
			current = e.Inner
		} else {
			stack = append([]struct {
				Source  string
				Message string
			}{{
				Message: current.Error(),
			}}, stack...)
			break
		}
	}

	// Add the current error last
	stack = append(stack, struct {
		Source  string
		Message string
	}{
		Source:  source,
		Message: message.Error(),
	})

	// Build colored output
	var b strings.Builder
	for i, err := range stack {
		if i > 0 {
			b.WriteString("\n")
			b.WriteString("  ")
		}
		b.WriteString(f.config.SourceColor)
		b.WriteString(err.Source)
		b.WriteString(colorReset)
		if err.Message != "" {
			b.WriteString(": ")
			b.WriteString(f.config.MessageColor)
			b.WriteString(err.Message)
			b.WriteString(colorReset)
		}
	}

	return b.String()
}

func (f *coloredFormatter) FormatStack(source string, message error) string {
	var b strings.Builder
	b.WriteString(f.config.SourceColor)
	b.WriteString(source)
	b.WriteString(colorReset)

	if message != nil {
		b.WriteString(": ")
		b.WriteString(f.config.MessageColor)
		b.WriteString(message.Error())
		b.WriteString(colorReset)
	}
	return b.String()
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
		SourceColor:  colorBlue,
		MessageColor: colorRed,
		InnerColor:   colorYellow,
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
