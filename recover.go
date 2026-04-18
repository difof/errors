package errors

import (
	"reflect"
	"runtime"
	"sync"
)

// recoverError normalizes a recovered panic value into the package's error model.
//
// There are three cases:
//  1. nil: no panic happened
//  2. error: return it unchanged so package-owned Must/Wrap panics and foreign
//     panic(error) values keep their original structure
//  3. non-error: wrap it into a package-owned leaf error and recover the best
//     caller location from the runtime stack
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

// recoverCallerPC walks the recovered stack and returns the first user frame
// that should own a synthesized package error for a non-error panic value.
//
// This is only used for raw panic values like panic("boom"). Package helpers
// such as Must already panic with a fully-formed error, so they bypass this
// path completely.
func recoverCallerPC() uintptr {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(2, pcs)
	if n == 0 {
		return 0
	}

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if !shouldSkipRecoverFrame(frame) {
			return frame.PC
		}

		if !more {
			return 0
		}
	}
}

var (
	// recoverSkipFrameEntries caches exact helper frame entry points that should
	// never be reported as the recovered callsite.
	recoverSkipFrameEntries map[uintptr]struct{}
	// recoverSkipFrameNames covers panic helpers whose runtime frame may not match
	// the entry-point cache exactly, but whose fully-qualified function name is
	// still stable and precise enough to skip explicitly.
	recoverSkipFrameNames       map[string]struct{}
	recoverSkipFrameEntriesOnce sync.Once
)

// shouldSkipRecoverFrame reports whether a runtime frame belongs to the
// recovery machinery rather than the user code that triggered the panic.
func shouldSkipRecoverFrame(frame runtime.Frame) bool {
	if frame.Function == "runtime.gopanic" || frame.Function == "runtime.sigpanic" {
		return true
	}

	recoverSkipFrameEntriesOnce.Do(func() {
		recoverSkipFrameEntries = map[uintptr]struct{}{
			functionEntry(Recover):         {},
			functionEntry(RecoverFn):       {},
			functionEntry(recoverError):    {},
			functionEntry(recoverCallerPC): {},
		}

		recoverSkipFrameNames = map[string]struct{}{
			functionName(Assert):  {},
			functionName(Assertf): {},
		}
	})

	_, ok := recoverSkipFrameEntries[frame.Entry]
	if ok {
		return true
	}

	_, ok = recoverSkipFrameNames[frame.Function]
	return ok
}

// functionEntry returns the canonical entry PC for fn.
func functionEntry(fn any) uintptr {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Entry()
}

// functionName returns the fully-qualified runtime name for fn.
func functionName(fn any) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
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
