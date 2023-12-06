package errors

import "fmt"

func Must[T any](r T, err error) T {
	if err != nil {
		panic(err)
	}

	return r
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
