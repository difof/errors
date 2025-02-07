package errors

import "fmt"

// Catch returns a new error if the given error is not nil, otherwise returns nil.
// This function wraps the error with stack trace information.
// Useful for returning error or nil as last statement in functions.
//
// Example:
//
//	return errors.Catch(db.Query("SELECT * FROM users"))
func Catch(err error) error {
	if err != nil {
		return WrapSkip(2, err)
	}
	return nil
}

// Catchf is same as Catch except that it accepts a message to be included with the error.
//
// Parameters:
//   - err: the error to be wrapped
//   - msg: format string for the error message
//   - params: arguments for the format string
//
// Example:
//
//	return errors.Catchf(db.Query("SELECT * FROM users"), "failed to query users table: %v", err)
func Catchf(err error, msg string, params ...any) error {
	if err != nil {
		msg = fmt.Sprintf(msg, params...)
		return WrapSkipf(2, err, msg)
	}
	return nil
}

// IgnoreResult is used in CatchResult callback to ignore the result value.
// Returns a callback that always returns nil error regardless of the input value.
//
// Example:
//
//	return CatchResult(rows, err)(IgnoreResult[*sql.Rows](), "failed to query users: %v", err)
func IgnoreResult[R any]() func(R) error { return func(R) error { return nil } }

// CatchResult is used for functions returning a value and an error.
// It provides a way to handle the success case with a callback while automatically
// handling the error case.
//
// Parameters:
//   - result: the value returned by the original function
//   - err: the error returned by the original function
//
// Returns a function that takes a callback which will be called only if err is nil.
// The callback receives the result value and can return an error.
//
// Example:
//
//	return CatchResult(db.Query("SELECT * FROM users"))(func(rows *sql.Rows) error {
//	    defer rows.Close()
//	    // process rows
//	    return nil
//	})
func CatchResult[R any](result R, err error) func(callback func(R) error) error {
	if err != nil {
		return func(f func(result R) error) error {
			return WrapSkip(3, err)
		}
	}

	return func(f func(result R) error) (err error) {
		if err = f(result); err != nil {
			return WrapSkip(3, err)
		}

		return
	}
}

// CatchResultf is same as CatchResult except that it appends a format message to the error.
//
// Parameters:
//   - result: the value returned by the original function
//   - err: the error returned by the original function
//
// Returns a function that takes:
//   - callback: function to process the result if no error occurred
//   - format: format string for error message
//   - params: arguments for the format string
//
// Example:
//
//	return CatchResultf(db.Query("SELECT * FROM users"))(func(rows *sql.Rows) error {
//	    defer rows.Close()
//	    // process rows
//	    return nil
//	}, "failed to query users: %v", err)
func CatchResultf[R any](result R, err error) func(callback func(R) error, format string, params ...any) error {
	if err != nil {
		return func(f func(result R) error, format string, params ...any) error {
			return WrapSkipf(3, err, format, params...)
		}
	}

	return func(f func(result R) error, format string, params ...any) (err error) {
		if err = f(result); err != nil {
			return WrapSkipf(3, err, format, params...)
		}

		return
	}
}
