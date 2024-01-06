package errors

// Maybe panics if err is not nil and doesn't implement T.
// Returns T if err is of type T.
//
// Useful for panic unless error can be handled.
func Maybe[E error](err error) (t E) {
	if err == nil {
		return
	}

	if Is(err, t) {
		return err.(E)
	}

	mayPanicf(err, "")
	return
}

// Maybef is same as Maybe, but returns a formatter function which panics with the given formatted error message.
func Maybef[E error](err error) func(format string, params ...any) E {
	return func(format string, params ...any) (t E) {
		if err == nil {
			return
		}

		if Is(err, t) {
			return err.(E)
		}

		mayPanicf(err, format, params...)
		return
	}
}

// MaybeResult is same as Maybe, but it's used for returning a result.
func MaybeResult[T any, E error](r T, err error) (t T, e E) {
	t = r

	if err == nil {
		t = r
		return
	}

	if Is(err, e) {
		e = err.(E)
		return
	}

	mayPanicf(err, "")
	return
}

// MaybeResultf is same as MaybeResult, but returns a formatter function which panics with the given formatted error message.
func MaybeResultf[T any, E error](r T, err error) func(format string, params ...any) (t T, e E) {
	return func(format string, params ...any) (t T, e E) {
		t = r

		if err == nil {
			t = r
			return
		}

		if Is(err, e) {
			e = err.(E)
			return
		}

		mayPanicf(err, format, params...)
		return
	}
}
