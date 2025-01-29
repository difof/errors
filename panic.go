package errors

// mayPanicf is an internal helper function that panics with a wrapped error
// if the error is not nil. If a format string is provided, it is used to
// create a new error message that wraps the original error.
//
// Parameters:
//   - err: the error to check
//   - format: optional format string for the error message
//   - params: arguments for the format string
//
// This function skips 2 stack frames when recording the source location
// to ensure the correct source file and line number are captured.
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
