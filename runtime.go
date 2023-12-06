package errors

import (
	"fmt"
	"path"
	"runtime"
)

var (
	showFuncName    = true
	showPackageName = true
)

// SetShowFuncName sets whether to show the function name in the stack trace before the file and line
func SetShowFuncName(state bool) {
	showFuncName = state
}

// SetShowPackageName sets whether to show the full function name (package name/function name)
func SetShowPackageName(state bool) {
	showPackageName = state
}

// getCallerPath returns the file and line which called any of New functions as string.
//
// skipFrames parameter defines how many functions to skip.
func getCallerPath(skipFrames int) string {
	pc, file, line, ok := runtime.Caller(2 + skipFrames)
	if !ok {
		return "<no source>"
	}

	f := runtime.FuncForPC(pc).Name()

	if showFuncName {
		if !showPackageName {
			f = stripPackageName(f)
		}

		return fmt.Sprintf("at %s %s:%d", f, file, line)
	}

	return fmt.Sprintf("%s:%d", file, line)
}

func stripPackageName(name string) string {
	return path.Base(name)
}
