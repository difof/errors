package errors

import (
	"runtime"
)

const NO_SOURCE = "<no source>"

// getCallerPath returns the caller's source location with optional function name
func getCallerPath(skip int) (funcname string, filepath string, line int) {
	pc, filepath, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "", NO_SOURCE, 0
	}

	funcname = runtime.FuncForPC(pc).Name()

	return
}
