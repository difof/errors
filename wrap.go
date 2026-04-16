package errors

// Wrap wraps an existing error with stack trace information.
// The wrapped error will include the source location where Wrap was called.
//
// Example:
//
//	if err := doSomething(); err != nil {
//	    return errors.Wrap(err)
//	}
func Wrap(err error) error {
	// there are two scenarios:
	// 1. err is an ErrorChain
	// 2. err is a standard error
	//
	// if err is an ErrorChain, simply set it as next error
	// if err is a standard error:
	// - unwrap loop until we get to the root error
	// - create no-pc error chains at each unwrap step
	return WrapSkip(2, err)
}

// WrapResult wraps an existing error and returns both the result and the error.
// Useful for last function calls that return a result and an error.
//
// Example:
//
//	result, err := errors.WrapResult(doSomething()) // doSomething() returns (result, error)
func WrapResult[T any](r T, err error) (T, error) {
	if err == nil {
		return r, nil
	}

	return r, WrapSkip(2, err)
}

// WrapResultf returns a formatter function that wraps an existing error and returns both the result and the error.
//
// The formatter function can be used to format the error message before returning it.
//
// Example:
//
//	result, err := errors.WrapResultf(doSomething())("failed to do something: %v", err)
func WrapResultf[T any](r T, err error) func(format string, params ...any) (T, error) {
	return func(format string, params ...any) (T, error) {
		if err == nil {
			return r, nil
		}

		return r, WrapSkipf(3, err, format, params...)
	}
}

// Wrapf creates a new formatted error that wraps an existing error.
// The new error message is formatted according to format and params.
//
// Parameters:
//   - inner: the error to wrap
//   - format: format string for the new error message
//   - params: arguments for the format string
//
// Example:
//
//	return errors.Wrapf(err, "failed to process user %s", username)
func Wrapf(inner error, format string, params ...any) error {
	return WrapSkipf(2, inner, format, params...)
}

// WrapSkip wraps an existing error, skipping the specified number
// of stack frames when recording the source location.
//
// Parameters:
//   - skip: number of stack frames to skip
//   - err: the error to wrap
func WrapSkip(skip int, err error) error {
	if err == nil {
		return nil
	}

	return newErrorChain(getCallerPC(skip), err, "")
}

// WrapSkipf creates a new formatted error wrapping an existing error,
// skipping the specified number of stack frames when recording the source location.
//
// Parameters:
//   - skip: number of stack frames to skip
//   - err: the error to wrap
//   - format: format string for the new error message
//   - params: arguments for the format string
func WrapSkipf(skip int, err error, format string, params ...any) error {
	if err == nil && format == "" {
		return nil
	}

	return newErrorChain(getCallerPC(skip), err, format, params...)
}
