package errors

// recoverError converts a recovered panic value into an error.
func recoverError(r any, skipFrames int) error {
	if r == nil {
		return nil
	}

	if err, ok := r.(error); ok {
		return err
	}
	return NewSkipf(skipFrames, "%v", r)
}

// Recover turns a panic into an error stored in errp.
// If the recovered value is not already an error, Recover wraps it in a new
// package-owned error.
//
// Recover is typically used in a deferred call at a function boundary, often in
// combination with Must helpers.
//
// A nil error pointer is ignored.
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

// RecoverFn is like Recover but passes the recovered error to fn instead of
// storing it through an error pointer.
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
