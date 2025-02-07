package errors

import (
	"runtime"
)

// getCallerPC returns the program counter of the caller,
// skipping the specified number of frames + the getCallerPC function itself.
func getCallerPC(skip int) uintptr {
	var pcs [1]uintptr
	if runtime.Callers(skip+2, pcs[:]) == 0 {
		return 0
	}
	return pcs[0]
}
