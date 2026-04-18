package errors

// Wrap adds a package-owned callsite above err.
//
// Example:
//
//	if err := doSomething(); err != nil {
//	    return errors.Wrap(err)
//	}
func Wrap(err error) error {
	return WrapSkip(2, err)
}

// WrapResult returns r unchanged and wraps err when err is not nil.
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

// WrapResultf is like WrapResult but delays the formatted context until the returned closure is called.
//
// Example:
//
//	result, err := errors.WrapResultf(doSomething())("failed to do something: %v", err)
func WrapResultf[T any](r T, err error) func(format string, params ...any) (T, error) {
	return func(format string, params ...any) (T, error) {
		if err == nil {
			return r, nil
		}

		return r, WrapSkipf(2, err, format, params...)
	}
}

// Wrapf adds formatted context above err.
//
// Example:
//
//	return errors.Wrapf(err, "failed to process user %s", username)
func Wrapf(inner error, format string, params ...any) error {
	return WrapSkipf(2, inner, format, params...)
}

// WrapSkip is like Wrap but skips additional stack frames before recording the callsite.
func WrapSkip(skip int, err error) error {
	if err == nil {
		return nil
	}

	node := newErrorNode(getCallerPC(skip), "")
	return newErrorChain(node, err)
}

// WrapSkipf is like Wrapf but skips additional stack frames before recording the callsite.
func WrapSkipf(skip int, err error, format string, params ...any) error {
	if err == nil && format == "" {
		return nil
	}

	node := newErrorNode(getCallerPC(skip), format, params...)
	return newErrorChain(node, err)
}
