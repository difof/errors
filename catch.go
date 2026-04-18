package errors

// Catch wraps err with a package-owned callsite when err is not nil.
// It returns nil when err is nil.
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

// Catchf is like Catch but adds formatted context.
//
// Example:
//
//	return errors.Catchf(db.Query("SELECT * FROM users"), "query users for account %d", accountID)
func Catchf(err error, msg string, params ...any) error {
	if err != nil {
		return WrapSkipf(2, err, msg, params...)
	}
	return nil
}

// IgnoreResult returns a CatchResult-style callback that ignores the success value.
//
// Example:
//
//	return errors.CatchResultf(tx, err)(errors.IgnoreResult[*sql.Tx](), "begin tx")
func IgnoreResult[R any]() func(R) error { return func(R) error { return nil } }

// CatchResult converts a `(result, err)` pair into a callback-based flow.
// The callback runs only when err is nil. Any error returned by the callback is
// wrapped with a package-owned callsite.
//
// Example:
//
//	rows, err := db.Query("SELECT * FROM users")
//	return errors.CatchResult(rows, err)(func(rows *sql.Rows) error {
//		defer rows.Close()
//		return scanUsers(rows)
//	})
func CatchResult[R any](result R, err error) func(callback func(R) error) error {
	if err != nil {
		return func(f func(result R) error) error {
			return WrapSkip(2, err)
		}
	}

	return func(f func(result R) error) (err error) {
		if err = f(result); err != nil {
			return WrapSkip(2, err)
		}

		return
	}
}

// CatchResultf is like CatchResult but adds formatted context when it wraps an error.
//
// Example:
//
//	rows, err := db.Query("SELECT * FROM users WHERE id = ?", id)
//	return errors.CatchResultf(rows, err)(func(rows *sql.Rows) error {
//		defer rows.Close()
//		return scanUser(rows)
//	}, "query user %d", id)
func CatchResultf[R any](result R, err error) func(callback func(R) error, format string, params ...any) error {
	if err != nil {
		return func(f func(result R) error, format string, params ...any) error {
			return WrapSkipf(2, err, format, params...)
		}
	}

	return func(f func(result R) error, format string, params ...any) (err error) {
		if err = f(result); err != nil {
			return WrapSkipf(2, err, format, params...)
		}

		return
	}
}
