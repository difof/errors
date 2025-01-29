package errors

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
	if r := recover(); r != nil {
		if errp == nil {
			return
		}

		err, ok := r.(error)
		if !ok {
			// Skipping 3 stack frames: recover.go, panic.go, must.go
			err = NewSkipf(3, "%v", r)
		}

		*errp = err
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
		err, ok := r.(error)
		if !ok {
			// Skipping 3 stack frames: recover.go, panic.go, must.go
			err = NewSkipf(3, "%v", r)
		}

		fn(err)
	}
}

// PassBack is an alias for Recover, providing a more descriptive name for the
// operation of passing a panic back as an error.
//
// Example:
//
//	func DoSomething() (err error) {
//	    defer errors.PassBack(&err)
//	    // ... code that might panic
//	    return
//	}
func PassBack(errp *error) {
	Recover(errp)
}
