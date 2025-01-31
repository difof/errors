package errors

import (
	goerrors "errors"
)

// Error is a lightweight drop-in replacement for standard errors package with stacktrace.
//
// It provides enhanced error handling capabilities by maintaining:
//   - FilePath: The file path where the error occurred (file:line)
//   - FuncPath: The function path where the error occurred (package.function)
//   - Line: The line number where the error occurred
//   - Message: The actual error
//   - Inner: The wrapped/underlying error for error chaining
type Error struct {
	FilePath string
	FuncPath string
	Line     int
	Message  error
	Inner    error
}

// NewError creates a new Error instance with the given source location, message, and inner error.
//
// Parameters:
//   - funcpath: typically the package.function where the error occurred
//   - filepath: typically the file:line where the error occurred
//   - line: the line number where the error occurred
//   - message: the error message to be displayed
//   - inner: optional underlying error to be wrapped
func NewError(funcpath string, filepath string, line int, message, inner error) *Error {
	return &Error{
		FilePath: filepath,
		FuncPath: funcpath,
		Line:     line,
		Message:  message,
		Inner:    inner,
	}
}

type ErrorEntry struct {
	Message  string `json:"message,omitempty" yaml:"message,omitempty"`
	FuncPath string `json:"func_path,omitempty" yaml:"func_path,omitempty"`
	FilePath string `json:"file_path,omitempty" yaml:"file_path,omitempty"`
	Line     int    `json:"line,omitempty" yaml:"line,omitempty"`
}

func (e *Error) ExtractEntries() []ErrorEntry {
	var entries []ErrorEntry

	e.Each(func(err error) bool {
		var e *Error
		if As(err, &e) {
			entry := ErrorEntry{
				FilePath: e.FilePath,
				FuncPath: e.FuncPath,
				Line:     e.Line,
			}

			if e.Message != nil {
				entry.Message = e.Message.Error()
			}

			entries = append(entries, entry)
		} else {
			entries = append(entries, ErrorEntry{
				FilePath: NO_SOURCE,
				Message:  err.Error(),
			})
		}

		return true
	})

	return entries
}

// Each iterates over all inner errors of Error chain.
//
// The iteration continues until either:
//
//   - The callback returns false
//   - There are no more inner errors to process
//
// Parameters:
//   - it: callback function that receives each error in the chain
//     Return false from the callback to stop iteration
func (e *Error) Each(it func(err error) bool) {
	if it == nil {
		return
	}

	for current := error(e); current != nil; current = goerrors.Unwrap(current) {
		if !it(current) {
			break
		}
	}
}

// Len returns the number of errors in the chain.
func (e *Error) Len() (count int) {
	current := error(e)
	for current != nil {
		count++
		if ec, ok := current.(*Error); ok {
			current = ec.Inner
		} else {
			break
		}
	}

	return
}

// ErrorMessageOf returns the innermost error message of an error.
//
// If the error is an *Error type, it traverses the chain to get the root message.
// For other error types, it returns the standard error message.
func ErrorMessageOf(err error) string {
	if err == nil {
		return ""
	}

	var e *Error
	if As(err, &e) {
		return e.ErrorMessage()
	}

	return err.Error()
}

// ErrorMessage returns the innermost error message without source location
// or stack trace information. This is useful when you only need the core
// error message without the additional context.
func (e *Error) ErrorMessage() (msg string) {
	e.Each(func(err error) bool {
		var e *Error
		if As(err, &e) {
			if e.Inner == nil {
				msg = e.Message.Error()
				return false
			}
		} else {
			msg = err.Error()
			return false
		}

		return true
	})

	return
}

// JSON returns a JSON formatted representation of the error
func (e *Error) JSON() string {
	return JSONFormatter(DefaultJSONConfig()).FormatError(e)
}

// YAML returns a YAML formatted representation of the error
func (e *Error) YAML() string {
	return YAMLFormatter(DefaultYAMLConfig()).FormatError(e)
}

// Colored returns a colored representation of the error for terminal output
func (e *Error) Colored() string {
	return ColoredFormatter(DefaultColorConfig()).FormatError(e)
}

// Error implements the error interface and returns the complete
// stack trace of this error as a newline-separated string.
func (e *Error) Error() string {
	return TextFormatter(DefaultTextConfig()).FormatError(e)
}

// Unwrap returns the inner error, implementing the interface
// required for errors.Is and errors.As compatibility.
func (e *Error) Unwrap() error { return e.Inner }

func (e *Error) Is(target error) bool {
	if target == nil {
		return e == nil
	}
	return goerrors.Is(e.Message, target) || goerrors.Is(e.Inner, target)
}

func (e *Error) As(target any) bool {
	return goerrors.As(e.Message, target) || goerrors.As(e.Inner, target)
}
