package errors

import goerrors "errors"

// As forwards to the standard library's errors.As.
func As(err error, target any) bool { return goerrors.As(err, target) }

// Is forwards to the standard library's errors.Is.
func Is(err, target error) bool { return goerrors.Is(err, target) }

// Unwrap forwards to the standard library's errors.Unwrap.
func Unwrap(err error) error { return goerrors.Unwrap(err) }
