package errors

import (
	"runtime"
	"strings"
)

// recoverError converts a recovered panic value into an error.
func recoverError(r any) error {
	if r == nil {
		return nil
	}

	if err, ok := r.(error); ok {
		return err
	}

	node := newErrorNode(recoverCallerPC(), "%v", r)
	return newErrorChain(node, nil)
}

func recoverCallerPC() uintptr {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(2, pcs)
	if n == 0 {
		return 0
	}

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if !shouldSkipRecoverFrame(frame.Function) {
			return frame.PC
		}

		if !more {
			return 0
		}
	}
}

func shouldSkipRecoverFrame(function string) bool {
	if function == "runtime.gopanic" || function == "runtime.sigpanic" {
		return true
	}

	return strings.HasSuffix(function, ".Recover") ||
		strings.HasSuffix(function, ".RecoverFn") ||
		strings.HasSuffix(function, ".recoverError") ||
		strings.HasSuffix(function, ".recoverCallerPC") ||
		strings.HasSuffix(function, ".Assert") ||
		strings.HasSuffix(function, ".Assertf")
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
		*errp = recoverError(r)
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
		fn(recoverError(r))
	}
}
