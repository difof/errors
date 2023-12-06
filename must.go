package errors

import "fmt"

// Must panics on error. Use Recover to catch the panic.
func Must[T any](r T, err error) T {
	if err != nil {
		panic(WrapSkip(1, err))
	}

	return r
}

// Mustf returns a formatter function which panics with the given formatted error message.
func Mustf[T any](r T, err error) func(format string, params ...any) T {
	return func(format string, params ...any) T {
		if err == nil {
			return r
		}

		panic(WrapSkipf(1, err, format, params...))
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
