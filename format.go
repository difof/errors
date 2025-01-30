package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
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

func (f *textFormatter) FormatError(filepath string, message error, inner error) string {
	var stack []string

	// Add current error first
	if message != nil {
		stack = append(stack, fmt.Sprintf("at %s: %s", filepath, message.Error()))
	} else {
		stack = append(stack, fmt.Sprintf("at %s", filepath))
	}

	// Add inner errors recursively
	var current error = inner
	var e *Error
	for current != nil {
		if As(current, &e) {
			if e.Message != nil {
				stack = append(stack, fmt.Sprintf("at %s: %s", e.FilePath, e.Message.Error()))
			} else {
				stack = append(stack, fmt.Sprintf("at %s", e.FilePath))
			}
			current = e.Inner
		} else {
			stack = append(stack, fmt.Sprintf("at %s", current.Error()))
			break
		}
	}

	// Reverse the stack to show root cause first
	for i := 0; i < len(stack)/2; i++ {
		j := len(stack) - 1 - i
		stack[i], stack[j] = stack[j], stack[i]
	}

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

func (f *textFormatter) FormatStack(filepath string, message error) string {
	if message == nil {
		return fmt.Sprintf("at %s", filepath)
	}
	return fmt.Sprintf("at %s: %s", filepath, message.Error())
}

// JSONConfig configures the JSON formatter
type JSONConfig struct {
	// Indent is the indentation string used for pretty printing
	Indent string
	// Prefix is the prefix used for each line in pretty printed output
	Prefix string
}

type jsonError struct {
	FilePath string `json:"filepath,omitempty"`
	FuncPath string `json:"funcpath,omitempty"`
	Message  string `json:"message,omitempty"`
}

// jsonFormatter formats errors as JSON objects
type jsonFormatter struct {
	config JSONConfig
}

func (f *jsonFormatter) FormatError(filepath string, message error, inner error) string {
	var stack []jsonError

	// Add current error first
	entry := jsonError{}
	if GetErrorConfig().ShowFilePath {
		entry.FilePath = filepath
	}
	if message != nil {
		entry.Message = message.Error()
	}
	if e, ok := message.(*Error); ok && GetErrorConfig().ShowFuncName {
		entry.FuncPath = e.FuncPath
	}
	stack = append(stack, entry)

	// Add inner errors recursively
	var current error = inner
	var e *Error
	for current != nil {
		if As(current, &e) {
			entry := jsonError{}
			if GetErrorConfig().ShowFilePath {
				entry.FilePath = e.FilePath
			}
			if e.Message != nil {
				entry.Message = e.Message.Error()
			}
			if GetErrorConfig().ShowFuncName {
				entry.FuncPath = e.FuncPath
			}
			stack = append(stack, entry)
			current = e.Inner
		} else {
			stack = append(stack, jsonError{Message: current.Error()})
			break
		}
	}

	// Reverse the stack to show root cause first
	for i := 0; i < len(stack)/2; i++ {
		j := len(stack) - 1 - i
		stack[i], stack[j] = stack[j], stack[i]
	}

	data, _ := json.MarshalIndent(stack, f.config.Prefix, f.config.Indent)
	return string(data)
}

func (f *jsonFormatter) FormatStack(filepath string, message error) string {
	entry := jsonError{}
	if GetErrorConfig().ShowFilePath {
		entry.FilePath = filepath
	}
	if message != nil {
		entry.Message = message.Error()
	}
	if e, ok := message.(*Error); ok && GetErrorConfig().ShowFuncName {
		entry.FuncPath = e.FuncPath
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

func (f *yamlFormatter) FormatError(filepath string, message error, inner error) string {
	var stack []struct {
		FilePath string
		FuncPath string
		Message  string
	}

	// Add current error first
	entry := struct {
		FilePath string
		FuncPath string
		Message  string
	}{}
	if GetErrorConfig().ShowFilePath {
		entry.FilePath = filepath
	}
	if message != nil {
		entry.Message = message.Error()
	}
	if e, ok := message.(*Error); ok && GetErrorConfig().ShowFuncName {
		entry.FuncPath = e.FuncPath
	}
	stack = append(stack, entry)

	// Add inner errors recursively
	var current error = inner
	var e *Error
	for current != nil {
		if As(current, &e) {
			entry := struct {
				FilePath string
				FuncPath string
				Message  string
			}{}
			if GetErrorConfig().ShowFilePath {
				entry.FilePath = e.FilePath
			}
			if e.Message != nil {
				entry.Message = e.Message.Error()
			}
			if GetErrorConfig().ShowFuncName {
				entry.FuncPath = e.FuncPath
			}
			stack = append(stack, entry)
			current = e.Inner
		} else {
			stack = append(stack, struct {
				FilePath string
				FuncPath string
				Message  string
			}{Message: current.Error()})
			break
		}
	}

	// Reverse the stack to show root cause first
	for i := 0; i < len(stack)/2; i++ {
		j := len(stack) - 1 - i
		stack[i], stack[j] = stack[j], stack[i]
	}

	// Build YAML output
	var b strings.Builder
	b.WriteString("errors:\n")
	for _, err := range stack {
		b.WriteString(f.config.Indent)
		b.WriteString("- ")
		var hasField bool
		if GetErrorConfig().ShowFilePath && err.FilePath != "" {
			b.WriteString("filepath: ")
			b.WriteString(err.FilePath)
			hasField = true
		}
		if GetErrorConfig().ShowFuncName && err.FuncPath != "" {
			if hasField {
				b.WriteString("\n")
				b.WriteString(f.config.Indent)
				b.WriteString("  ")
			}
			b.WriteString("funcpath: ")
			b.WriteString(err.FuncPath)
			hasField = true
		}
		if err.Message != "" {
			if hasField {
				b.WriteString("\n")
				b.WriteString(f.config.Indent)
				b.WriteString("  ")
			}
			b.WriteString("message: ")
			b.WriteString(err.Message)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (f *yamlFormatter) FormatStack(filepath string, message error) string {
	var b strings.Builder
	var hasField bool

	if GetErrorConfig().ShowFilePath {
		b.WriteString("filepath: ")
		b.WriteString(filepath)
		hasField = true
	}

	if e, ok := message.(*Error); ok && GetErrorConfig().ShowFuncName && e.FuncPath != "" {
		if hasField {
			b.WriteString("\n")
		}
		b.WriteString("funcpath: ")
		b.WriteString(e.FuncPath)
		hasField = true
	}

	if message != nil {
		if hasField {
			b.WriteString("\n")
		}
		b.WriteString("message: ")
		b.WriteString(message.Error())
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

func (f *coloredFormatter) FormatError(filepath string, message error, inner error) string {
	var stack []struct {
		FilePath string
		FuncPath string
		Message  string
	}

	// Add current error first
	entry := struct {
		FilePath string
		FuncPath string
		Message  string
	}{}
	if GetErrorConfig().ShowFilePath {
		entry.FilePath = filepath
	}
	if message != nil {
		entry.Message = message.Error()
	}
	if e, ok := message.(*Error); ok && GetErrorConfig().ShowFuncName {
		entry.FuncPath = e.FuncPath
	}
	stack = append(stack, entry)

	// Add inner errors recursively
	var current error = inner
	var e *Error
	for current != nil {
		if As(current, &e) {
			entry := struct {
				FilePath string
				FuncPath string
				Message  string
			}{}
			if GetErrorConfig().ShowFilePath {
				entry.FilePath = e.FilePath
			}
			if e.Message != nil {
				entry.Message = e.Message.Error()
			}
			if GetErrorConfig().ShowFuncName {
				entry.FuncPath = e.FuncPath
			}
			stack = append(stack, entry)
			current = e.Inner
		} else {
			stack = append(stack, struct {
				FilePath string
				FuncPath string
				Message  string
			}{Message: current.Error()})
			break
		}
	}

	// Reverse the stack to show root cause first
	for i := 0; i < len(stack)/2; i++ {
		j := len(stack) - 1 - i
		stack[i], stack[j] = stack[j], stack[i]
	}

	// Build colored output
	var b strings.Builder
	for i, err := range stack {
		if i > 0 {
			b.WriteString("\n  ")
		}
		b.WriteString("at ")
		if GetErrorConfig().ShowFuncName && err.FuncPath != "" {
			b.WriteString(f.config.InnerColor.Sprint(err.FuncPath))
			b.WriteString(" ")
		}
		if GetErrorConfig().ShowFilePath && err.FilePath != "" {
			b.WriteString(f.config.SourceColor.Sprint(err.FilePath))
		}
		if err.Message != "" {
			if (GetErrorConfig().ShowFilePath && err.FilePath != "") || (GetErrorConfig().ShowFuncName && err.FuncPath != "") {
				b.WriteString(": ")
			}
			b.WriteString(f.config.MessageColor.Sprint(err.Message))
		}
	}

	return b.String()
}

func (f *coloredFormatter) FormatStack(filepath string, message error) string {
	var b strings.Builder
	b.WriteString("at ")

	if e, ok := message.(*Error); ok && GetErrorConfig().ShowFuncName && e.FuncPath != "" {
		b.WriteString(f.config.InnerColor.Sprint(e.FuncPath))
		b.WriteString(" ")
	}

	if GetErrorConfig().ShowFilePath {
		b.WriteString(f.config.SourceColor.Sprint(filepath))
	}

	if message != nil {
		if GetErrorConfig().ShowFilePath || (message.(*Error) != nil && GetErrorConfig().ShowFuncName) {
			b.WriteString(": ")
		}
		b.WriteString(f.config.MessageColor.Sprint(message.Error()))
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
