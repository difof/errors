package errors

import (
	"errors"
	"fmt"
)

// Wrap wraps an existing error with stack trace information.
// The wrapped error will include the source location where Wrap was called.
//
// Example:
//
//	if err := doSomething(); err != nil {
//	    return errors.Wrap(err)
//	}
func Wrap(inner error) error {
	return NewError(getCallerPath(0), nil, inner)
}

// WrapResult wraps an existing error and returns both the result and the error.
// Useful for last function calls that return a result and an error.
//
// Example:
//
//	result, err := errors.WrapResult(doSomething()) // doSomething() returns (result, error)
func WrapResult[T any](r T, err error) (T, error) {
	if err == nil {
		return r, nil
	}

	return r, NewError(getCallerPath(0), nil, err)
}

// WrapResultf returns a formatter function that wraps an existing error and returns both the result and the error.
//
// The formatter function can be used to format the error message before returning it.
//
// Example:
//
//	result, err := errors.WrapResultf(doSomething())("failed to do something: %v", err)
func WrapResultf[T any](r T, err error) func(format string, params ...any) (T, error) {
	return func(format string, params ...any) (T, error) {
		if err == nil {
			return r, nil
		}

		return r, NewError(getCallerPath(0), nil, err)
	}
}

// Wrape creates a new error that wraps an existing error, adding both
// a new error message and the original error's information.
//
// Parameters:
//   - err: the new error message
//   - inner: the error to wrap
//
// Example:
//
//	return errors.Wrape(errors.New("validation failed"), err)
func Wrape(err error, inner error) error {
	return NewError(getCallerPath(0), err, inner)
}

// Wrapf creates a new formatted error that wraps an existing error.
// The new error message is formatted according to format and params.
//
// Parameters:
//   - inner: the error to wrap
//   - format: format string for the new error message
//   - params: arguments for the format string
//
// Example:
//
//	return errors.Wrapf(err, "failed to process user %s", username)
func Wrapf(inner error, format string, params ...any) error {
	return NewError(getCallerPath(0), errors.New(fmt.Sprintf(format, params...)), inner)
}

// WrapSkip wraps an existing error, skipping the specified number
// of stack frames when recording the source location.
//
// Parameters:
//   - skip: number of stack frames to skip
//   - inner: the error to wrap
func WrapSkip(skip int, inner error) error {
	return NewError(getCallerPath(skip), nil, inner)
}

// WrapSkipf creates a new formatted error wrapping an existing error,
// skipping the specified number of stack frames when recording the source location.
//
// Parameters:
//   - skip: number of stack frames to skip
//   - inner: the error to wrap
//   - format: format string for the new error message
//   - params: arguments for the format string
func WrapSkipf(skip int, inner error, format string, params ...any) error {
	return NewError(getCallerPath(skip), errors.New(fmt.Sprintf(format, params...)), inner)
}
