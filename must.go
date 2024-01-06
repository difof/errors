package errors

import "fmt"

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

// MustResult2 panics on error. Use Recover to catch the panic.
func MustResult2[A, B any](a A, b B, err error) (A, B) {
	mayPanicf(err, "")
	return a, b
}

// MustResult2f returns a formatter function which panics with the given formatted error message.
func MustResult2f[A, B any](a A, b B, err error) func(format string, params ...any) (A, B) {
	return func(format string, params ...any) (A, B) {
		mayPanicf(err, format, params...)
		return a, b
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
