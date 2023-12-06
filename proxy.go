package errors

import goerrors "errors"

// As wrapper around go's standard errors.As
func As(err error, target any) bool { return goerrors.As(err, target) }

// Is wrapper around go's standard errors.Is
func Is(err, target error) bool { return goerrors.Is(err, target) }

// Unwrap wrapper around go's standard errors.Unwrap
func Unwrap(err error) error { return goerrors.Unwrap(err) }
