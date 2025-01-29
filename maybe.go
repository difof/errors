package errors

// Maybe panics if err is not nil and doesn't implement type E.
// Returns err as type E if err implements E, otherwise panics.
//
// This function is useful when you expect a specific error type and want to
// handle it, but want to panic for any other unexpected errors.
//
// Type Parameters:
//   - E: the expected error type
//
// Parameters:
//   - err: the error to check
//
// Example:
//
//	var netErr *net.OpError
//	netErr = errors.Maybe[*net.OpError](err) // panics if err is not *net.OpError
func Maybe[E error](err error) (t E) {
	if err == nil {
		return
	}

	// Try direct type assertion first
	if target, ok := err.(E); ok {
		return target
	}

	// Try unwrapping
	if As(err, &t) {
		return t
	}

	mayPanicf(err, "")
	return
}

// Maybef is same as Maybe, but returns a formatter function that panics
// with a formatted message if the error doesn't match the expected type.
//
// Type Parameters:
//   - E: the expected error type
//
// Parameters:
//   - err: the error to check
//
// Returns a function that takes:
//   - format: format string for error message if panic occurs
//   - params: arguments for the format string
//
// Example:
//
//	var netErr *net.OpError
//	netErr = errors.Maybef[*net.OpError](err)("unexpected error type: %v", err)
func Maybef[E error](err error) func(format string, params ...any) E {
	return func(format string, params ...any) (t E) {
		if err == nil {
			return
		}

		// Try direct type assertion first
		if target, ok := err.(E); ok {
			return target
		}

		// Try unwrapping
		if As(err, &t) {
			return t
		}

		mayPanicf(err, format, params...)
		return
	}
}

// MaybeResult is same as Maybe but works with a result value and an error.
// Returns the result and the error as type E if err implements E,
// otherwise panics.
//
// Type Parameters:
//   - T: the result type
//   - E: the expected error type
//
// Parameters:
//   - r: the result value
//   - err: the error to check
//
// Example:
//
//	value, timeoutErr := errors.MaybeResult[int, *TimeoutError](result, err)
func MaybeResult[T any, E error](r T, err error) (t T, e E) {
	t = r

	if err == nil {
		return
	}

	// Try direct type assertion first
	if target, ok := err.(E); ok {
		e = target
		return
	}

	// Try unwrapping
	if As(err, &e) {
		return
	}

	mayPanicf(err, "")
	return
}

// MaybeResultf is same as MaybeResult but returns a formatter function
// that panics with a formatted message if the error doesn't match the
// expected type.
//
// Type Parameters:
//   - T: the result type
//   - E: the expected error type
//
// Parameters:
//   - r: the result value
//   - err: the error to check
//
// Returns a function that takes:
//   - format: format string for error message if panic occurs
//   - params: arguments for the format string
//
// Example:
//
//	value, timeoutErr := errors.MaybeResultf[int, *TimeoutError](result, err)(
//	    "unexpected error type: %v", err)
func MaybeResultf[T any, E error](r T, err error) func(format string, params ...any) (t T, e E) {
	return func(format string, params ...any) (t T, e E) {
		t = r

		if err == nil {
			return
		}

		// Try direct type assertion first
		if target, ok := err.(E); ok {
			e = target
			return
		}

		// Try unwrapping
		if As(err, &e) {
			return
		}

		mayPanicf(err, format, params...)
		return
	}
}
