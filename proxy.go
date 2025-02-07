package errors

import goerrors "errors"

// As is a wrapper around Go's standard errors.As function,
// so user doesn't need to import std errors package.
func As(err error, target any) bool { return goerrors.As(err, target) }

// Is is a wrapper around Go's standard errors.Is function,
// so user doesn't need to import std errors package.
func Is(err, target error) bool { return goerrors.Is(err, target) }

// Unwrap is a wrapper around Go's standard errors.Unwrap function,
// so user doesn't need to import std errors package.
func Unwrap(err error) error { return goerrors.Unwrap(err) }
