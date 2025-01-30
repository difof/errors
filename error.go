package errors

import (
	goerrors "errors"
	"strings"
)

// Error is a lightweight drop-in replacement for standard errors package with stacktrace.
//
// It provides enhanced error handling capabilities by maintaining:
//   - Source: The location where the error occurred (file:line)
//   - Message: The actual error message
//   - Inner: The wrapped/underlying error for error chaining
type Error struct {
	Source  string
	Message error
	Inner   error
}

// NewError creates a new Error instance with the given source location, message, and inner error.
//
// Parameters:
//   - source: typically the file:line where the error occurred
//   - message: the error message to be displayed
//   - inner: optional underlying error to be wrapped
func NewError(source string, message, inner error) *Error {
	return &Error{
		Source:  source,
		Message: message,
		Inner:   inner,
	}
}

// ErrorMessageOf returns the innermost error message of the error chain.
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

	var current error = e
	for current != nil {
		if !it(current) {
			break
		}

		var cast *Error
		if As(current, &cast) {
			current = cast.Unwrap()
		} else {
			current = nil
		}
	}
}

// StackTrace builds a complete stack trace of all inner errors.
//
// Returns a slice of strings where each entry represents one level
// in the error chain, starting from the innermost error.
//
// The returned stack trace is in reverse order (root cause first).
func (e *Error) StackTrace() (list []string) {
	list = make([]string, 0, 5)

	defer func() {
		// reverse to show root cause first
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
	}()

	e.Each(func(err error) bool {
		var e *Error
		if As(err, &e) {
			list = append(list, e.String())
		} else {
			list = append(list, err.Error())
		}
		return true
	})

	return
}

// String returns a formatted string of the current error level,
// combining the source location and error message if present.
// If there's no message, only the source is returned.
func (e *Error) String() string {
	return GetFormatter().FormatError(e.Source, e.Message, e.Inner)
}

// JSON returns a JSON formatted representation of the error
func (e *Error) JSON() string {
	return JSONFormatter(DefaultJSONConfig()).FormatError(e.Source, e.Message, e.Inner)
}

// YAML returns a YAML formatted representation of the error
func (e *Error) YAML() string {
	return YAMLFormatter(DefaultYAMLConfig()).FormatError(e.Source, e.Message, e.Inner)
}

// Colored returns a colored representation of the error for terminal output
func (e *Error) Colored() string {
	return ColoredFormatter(DefaultColorConfig()).FormatError(e.Source, e.Message, e.Inner)
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

// Error implements the error interface and returns the complete
// stack trace of this error as a newline-separated string.
func (e *Error) Error() string {
	return strings.Join(e.StackTrace(), "\n")
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
