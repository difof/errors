package errors

import "fmt"

func mayPanicf(err error, format string, params ...any) {
	if err == nil {
		return
	}

	if format == "" {
		panic(WrapSkip(2, err))
	} else {
		panic(WrapSkipf(2, err, format, params...))
	}
}

// MustResult panics on error. Use Recover to catch the panic.
func MustResult[T any](r T, err error) T {
	mayPanicf(err, "")
	return r
}

// MustResultf returns a formatter function which panics with the given formatted error message.
func MustResultf[T any](r T, err error) func(format string, params ...any) T {
	return func(format string, params ...any) T {
		mayPanicf(err, format, params...)
		return r
	}
}

// Must panics with the given error.
func Must(err error) {
	mayPanicf(err, "")
}

// Mustf panics with the given error.
func Mustf(err error) func(format string, params ...any) {
	return func(format string, params ...any) {
		mayPanicf(err, format, params...)
	}
}

func Ignore[T any](r T, _ error) T {
	return r
}

func Assert(truth bool, message string) {
	if !truth {
		panic(message)
	}
}

func Assertf(truth bool, format string, params ...any) {
	if !truth {
		panic(fmt.Sprintf(format, params...))
	}
}
