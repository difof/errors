package errors

func IsUnwrapSingle(err error) (error, bool) {
	if _, ok := err.(interface{ Unwrap() error }); ok {
		return err, ok
	}

	return nil, false
}

func IsUnwrapMulti(err error) (error, bool) {
	if _, ok := err.(interface{ Unwrap() []error }); ok {
		return err, ok
	}

	return nil, false
}

func TryUnwrapSingle(err error) (result error, ok bool) {
	if cast, ok := err.(interface{ Unwrap() error }); ok {
		return cast.Unwrap(), ok
	}

	return nil, false
}

func TryUnwrapMulti(err error) ([]error, bool) {
	if cast, ok := err.(interface{ Unwrap() []error }); ok {
		return cast.Unwrap(), ok
	}

	return nil, false
}

func IsErrorChain(err error) (*ErrorChain, bool) {
	if cast, ok := err.(*ErrorChain); ok {
		return cast, ok
	}

	return nil, false
}

func IsErrorTree(err error) (*ErrorTree, bool) {
	if cast, ok := err.(*ErrorTree); ok {
		return cast, ok
	}

	return nil, false
}
