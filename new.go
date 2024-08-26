package errors

import (
	"errors"
	"fmt"
)

// New creates a new error
func New(msg string) error {
	return NewError(getCallerPath(0), errors.New(msg), nil)
}

// Newf creates a new formatted error
func Newf(format string, params ...any) error {
	return NewError(getCallerPath(0), errors.New(fmt.Sprintf(format, params...)), nil)
}

// Wrap wraps the error
func Wrap(inner error) error {
	return NewError(getCallerPath(0), nil, inner)
}

// Warpe creates a new error with inner error
func Warpe(err error, inner error) error {
	return NewError(getCallerPath(0), err, inner)
}

// Wrapf creates a new formatted error with inner error
func Wrapf(inner error, format string, params ...any) error {
	return NewError(getCallerPath(0), errors.New(fmt.Sprintf(format, params...)), inner)
}

// NewSkip creates a new error with custom stack skip
func NewSkip(skip int, msg string) error {
	return NewError(getCallerPath(skip), errors.New(msg), nil)
}

// NewSkipf creates a new formatted error with custom stack skip
func NewSkipf(skip int, format string, params ...any) error {
	return NewError(getCallerPath(skip), errors.New(fmt.Sprintf(format, params...)), nil)
}

// WrapSkip wraps the error with custom stack skip
func WrapSkip(skip int, inner error) error {
	return NewError(getCallerPath(skip), nil, inner)
}

// WrapSkipf creates a new formatted error with inner error and custom stack skip
func WrapSkipf(skip int, inner error, format string, params ...any) error {
	return NewError(getCallerPath(skip), errors.New(fmt.Sprintf(format, params...)), inner)
}
