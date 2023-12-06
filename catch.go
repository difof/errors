package errors

import "fmt"

// Catch returns a new error if the given error is not nil, otherwise returns nil.
//
// Useful for returning error or nil as last statement.
func Catch(err error) error {
	if err != nil {
		return WrapSkip(1, err)
	}
	return nil
}

// Catchf is same as Catch except that it accepts a message
func Catchf(err error, msg string, params ...any) error {
	if err != nil {
		msg = fmt.Sprintf(msg, params...)
		return WrapSkipf(1, err, msg)
	}

	return nil
}

// IgnoreResult is used in CatchResult callback to ignore the result
func IgnoreResult[R any]() func(R) error { return func(R) error { return nil } }

// CatchResult is used for two return values functions returning an error.
//
// You should call the returned function,
// callback will be called if error is nil, otherwise it returns the error.
// Also returns the error returned by the callback.
//
// This function is a shortcut for when you either return an error or handle a result as the last statement.
func CatchResult[R any](result R, err error) func(callback func(R) error) error {
	if err != nil {
		return func(f func(result R) error) error {
			return err
		}
	}

	return func(f func(result R) error) (err error) {
		if err = f(result); err != nil {
			return WrapSkip(1, f(result))
		}

		return
	}
}

// CatchResultf is same as CatchResult except that it appends a format message to the error.
func CatchResultf[R any](result R, err error) func(callback func(R) error, format string, params ...any) error {
	if err != nil {
		return func(f func(result R) error, format string, params ...any) error {
			return WrapSkipf(1, err, format, params...)
		}
	}

	return func(f func(result R) error, format string, params ...any) (err error) {
		if err = f(result); err != nil {
			return WrapSkipf(1, f(result), format, params...)
		}

		return
	}
}
