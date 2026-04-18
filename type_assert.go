package errors

// IsUnwrapSingle reports whether err implements Unwrap() error.
func IsUnwrapSingle(err error) (error, bool) {
	if _, ok := err.(interface{ Unwrap() error }); ok {
		return err, ok
	}

	return nil, false
}

// IsUnwrapMulti reports whether err implements Unwrap() []error.
func IsUnwrapMulti(err error) (error, bool) {
	if _, ok := err.(interface{ Unwrap() []error }); ok {
		return err, ok
	}

	return nil, false
}

// TryUnwrapSingle returns the result of Unwrap() error when available.
func TryUnwrapSingle(err error) (result error, ok bool) {
	if cast, ok := err.(interface{ Unwrap() error }); ok {
		return cast.Unwrap(), ok
	}

	return nil, false
}

// TryUnwrapMulti returns the result of Unwrap() []error when available.
func TryUnwrapMulti(err error) ([]error, bool) {
	if cast, ok := err.(interface{ Unwrap() []error }); ok {
		return cast.Unwrap(), ok
	}

	return nil, false
}

// IsErrorChain reports whether err is a package-owned ErrorChain.
func IsErrorChain(err error) (*ErrorChain, bool) {
	if cast, ok := err.(*ErrorChain); ok {
		return cast, ok
	}

	return nil, false
}

// IsErrorTree reports whether err is a package-owned ErrorTree.
func IsErrorTree(err error) (*ErrorTree, bool) {
	if cast, ok := err.(*ErrorTree); ok {
		return cast, ok
	}

	return nil, false
}
