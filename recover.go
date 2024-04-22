package errors

// Recover recovers from panic and sets the error pointer to the recovered error.
//
// Can be used in conjunction with Must to handle any errors that may occur.
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

// PassBack an alias for Recover
func PassBack(errp *error) {
	Recover(errp)
}
