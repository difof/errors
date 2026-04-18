package errors

import (
	"runtime"
)

// getCallerPC returns the caller program counter after skipping skip frames.
func getCallerPC(skip int) uintptr {
	var pcs [1]uintptr
	if runtime.Callers(skip+2, pcs[:]) == 0 {
		return 0
	}
	return pcs[0]
}
