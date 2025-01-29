package errors

import (
	"errors"
	"fmt"
)

// New creates a new error with stack trace information.
// The error message will include the source location where New was called.
//
// Example:
//
//	return errors.New("connection failed")
func New(msg string) error {
	return NewError(getCallerPath(0), errors.New(msg), nil)
}

// Newf creates a new formatted error with stack trace information.
// The error message will include the source location where Newf was called.
//
// Example:
//
//	return errors.Newf("failed to connect to %s: %v", host, err)
func Newf(format string, params ...any) error {
	return NewError(getCallerPath(0), errors.New(fmt.Sprintf(format, params...)), nil)
}

// NewSkip creates a new error, skipping the specified number of stack frames
// when recording the source location.
//
// Parameters:
//   - skip: number of stack frames to skip
//   - msg: error message
func NewSkip(skip int, msg string) error {
	return NewError(getCallerPath(skip), errors.New(msg), nil)
}

// NewSkipf creates a new formatted error, skipping the specified number
// of stack frames when recording the source location.
//
// Parameters:
//   - skip: number of stack frames to skip
//   - format: format string for the error message
//   - params: arguments for the format string
func NewSkipf(skip int, format string, params ...any) error {
	return NewError(getCallerPath(skip), errors.New(fmt.Sprintf(format, params...)), nil)
}
