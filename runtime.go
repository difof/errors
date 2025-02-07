package errors

import (
	"runtime"
)

// getCallerPath returns the caller's source location with optional function name
func getCallerPath(skip int) (funcname string, filepath string, line int) {
	pc, filepath, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "", "", 0
	}

	funcname = runtime.FuncForPC(pc).Name()

	return
}

func getCallerPC(skip int) uintptr {
	var pcs [1]uintptr
	if runtime.Callers(skip+1, pcs[:]) == 0 {
		return 0
	}
	return pcs[0]
}
