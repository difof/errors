package errors

// recoverError handles panic recovery and converts the recovered value into an error.
// skipFrames specifies how many stack frames to skip when creating a new error.
func recoverError(r any, skipFrames int) error {
	if r == nil {
		return nil
	}

	if err, ok := r.(error); ok {
		return err
	}
	return NewSkipf(skipFrames, "%v", r)
}

// Recover recovers from panic and sets the error pointer to the recovered error.
// If the recovered value is not an error, it will be wrapped in a new error.
//
// This function is typically used in a defer statement to handle panics and
// convert them to errors that can be returned from the function, specially
// in combination with Must.
//
// The error pointer should not be nil, otherwise the panic will be propagated.
//
// Example:
//
//	func DoSomething() (err error) {
//	    defer errors.Recover(&err)
//	    // ... call chain that might panic
//	    return
//	}
func Recover(errp *error) {
	if errp == nil {
		return
	}

	if r := recover(); r != nil {
		*errp = recoverError(r, 3)
	}
}

// RecoverFn recovers from panic and calls the given function with the recovered error.
// If the recovered value is not an error, it will be wrapped in a new error.
//
// This function provides more flexibility than Recover by allowing custom error
// handling through a callback function.
//
// Parameters:
//   - fn: callback function that will be called with the recovered error
//
// Example:
//
//	defer errors.RecoverFn(func(err error) {
//	    log.Printf("recovered from panic: %v", err)
//	})
func RecoverFn(fn func(error)) {
	if r := recover(); r != nil {
		fn(recoverError(r, 3))
	}
}
