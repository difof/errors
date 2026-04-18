package errors

import "fmt"

// MustResult returns r when err is nil and panics with a wrapped error otherwise.
// It is commonly paired with Recover at a higher boundary.
//
// Example:
//
//	user := errors.MustResult(db.GetUser(id))
func MustResult[T any](r T, err error) T {
	mayPanicf(2, err, "")
	return r
}

// MustResultf is like MustResult but adds formatted context to the panic value.
//
// Example:
//
//	user := errors.MustResultf(db.GetUser(id))("failed to get user %d", id)
func MustResultf[T any](r T, err error) func(format string, params ...any) T {
	return func(format string, params ...any) T {
		mayPanicf(2, err, format, params...)
		return r
	}
}

// MustResult2 returns both values when err is nil and panics otherwise.
//
// Example:
//
//	key, value := errors.MustResult2(cache.Get("mykey"))
func MustResult2[A, B any](a A, b B, err error) (A, B) {
	mayPanicf(2, err, "")
	return a, b
}

// MustResult2f is like MustResult2 but adds formatted context to the panic value.
//
// Example:
//
//	key, value := errors.MustResult2f(cache.Get("mykey"))("failed to get cache key %s", "mykey")
func MustResult2f[A, B any](a A, b B, err error) func(format string, params ...any) (A, B) {
	return func(format string, params ...any) (A, B) {
		mayPanicf(2, err, format, params...)
		return a, b
	}
}

// Must panics with a wrapped error when err is not nil.
//
// Example:
//
//	errors.Must(db.Connect())
func Must(err error) {
	mayPanicf(2, err, "")
}

// Mustf is like Must but adds formatted context to the panic value.
//
// Example:
//
//	errors.Mustf(db.Connect())("failed to connect to database: %v", err)
func Mustf(err error) func(format string, params ...any) {
	return func(format string, params ...any) {
		mayPanicf(2, err, format, params...)
	}
}

// Ignore returns r and discards the accompanying error.
//
// Example:
//
//	value := errors.Ignore(strconv.Atoi("123"))
func Ignore[T any](r T, _ error) T {
	return r
}

// Assert panics with message when truth is false.
//
// Example:
//
//	errors.Assert(len(slice) > 0, "slice must not be empty")
func Assert(truth bool, message string) {
	if !truth {
		panic(message)
	}
}

// Assertf panics with a formatted message when truth is false.
//
// Example:
//
//	errors.Assertf(len(slice) > 0, "slice %s must not be empty", name)
func Assertf(truth bool, format string, params ...any) {
	if !truth {
		panic(fmt.Sprintf(format, params...))
	}
}
