package errors

import "fmt"

// MustResult panics if the error is not nil, otherwise returns the result.
// This is useful for operations that shouldn't fail during normal execution.
// Use Recover to catch the panic if needed.
//
// Example:
//
//	user := errors.MustResult(db.GetUser(id))
func MustResult[T any](r T, err error) T {
	mayPanicf(err, "")
	return r
}

// MustResultf returns a formatter function that panics with a formatted message
// if the error is not nil, otherwise returns the result.
//
// Example:
//
//	user := errors.MustResultf(db.GetUser(id))("failed to get user %d", id)
func MustResultf[T any](r T, err error) func(format string, params ...any) T {
	return func(format string, params ...any) T {
		mayPanicf(err, format, params...)
		return r
	}
}

// MustResult2 panics if the error is not nil, otherwise returns both results.
// This is useful for operations that return two values and shouldn't fail
// during normal execution.
//
// Example:
//
//	key, value := errors.MustResult2(cache.Get("mykey"))
func MustResult2[A, B any](a A, b B, err error) (A, B) {
	mayPanicf(err, "")
	return a, b
}

// MustResult2f returns a formatter function that panics with a formatted message
// if the error is not nil, otherwise returns both results.
//
// Example:
//
//	key, value := errors.MustResult2f(cache.Get("mykey"))("failed to get cache key %s", "mykey")
func MustResult2f[A, B any](a A, b B, err error) func(format string, params ...any) (A, B) {
	return func(format string, params ...any) (A, B) {
		mayPanicf(err, format, params...)
		return a, b
	}
}

// Must panics if the error is not nil.
// This is useful for operations that shouldn't fail during normal execution.
//
// Example:
//
//	errors.Must(db.Connect())
func Must(err error) {
	mayPanicf(err, "")
}

// Mustf panics with a formatted message if the error is not nil.
//
// Example:
//
//	errors.Mustf(db.Connect())("failed to connect to database: %v", err)
func Mustf(err error) func(format string, params ...any) {
	return func(format string, params ...any) {
		mayPanicf(err, format, params...)
	}
}

// Ignore returns the first result and ignores the error.
// Use with caution as this function deliberately ignores error handling.
//
// Example:
//
//	value := errors.Ignore(strconv.Atoi("123"))
func Ignore[T any](r T, _ error) T {
	return r
}

// Assert panics with the given message if the condition is false.
// This is useful for checking invariants that must be true.
//
// Example:
//
//	errors.Assert(len(slice) > 0, "slice must not be empty")
func Assert(truth bool, message string) {
	if !truth {
		panic(message)
	}
}

// Assertf panics with a formatted message if the condition is false.
// This is useful for checking invariants that must be true.
//
// Example:
//
//	errors.Assertf(len(slice) > 0, "slice %s must not be empty", name)
func Assertf(truth bool, format string, params ...any) {
	if !truth {
		panic(fmt.Sprintf(format, params...))
	}
}
