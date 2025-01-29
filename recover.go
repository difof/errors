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
// convert them to errors that can be returned from the function.
//
// It can be used in conjunction with Must to handle any errors that may occur,
// so you don't have to handle every error.
//
// Parameters:
//   - errp: pointer to error variable that will receive the recovered error
//
// Example:
//
//	func DoSomething() (err error) {
//	    defer errors.Recover(&err)
//	    // ... code that might panic
//	    return
//	}
func Recover(errp *error) {
	if errp == nil {
		// If errp is nil, let the panic propagate
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

// HandlePanic is an alias for Recover, providing a more descriptive name for the
// operation of handling a panic and converting it to an error.
//
// Example:
//
//	func DoSomething() (err error) {
//	    defer errors.HandlePanic(&err)
//	    // ... code that might panic
//	    return
//	}
func HandlePanic(errp *error) {
	Recover(errp)
}
