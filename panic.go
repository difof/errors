package errors

// mayPanicf is an internal helper function that panics with a wrapped error
// if the error is not nil. If a format string is provided, it is used to
// create a new error message that wraps the original error.
//
// Parameters:
//   - skip: number of stack frames to skip
//   - err: the error to check
//   - format: optional format string for the error message
//   - params: arguments for the format string
func mayPanicf(skip int, err error, format string, params ...any) {
	if err == nil {
		return
	}

	if format == "" {
		panic(WrapSkip(skip+1, err))
	} else {
		panic(WrapSkipf(skip+1, err, format, params...))
	}
}
