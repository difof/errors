package errors

import (
	"fmt"
	"path"
	"runtime"
)

var (
	// showFuncName controls whether function names are included in stack traces
	showFuncName = true
	// showPackageName controls whether full package paths are shown in function names
	showPackageName = true
)

// SetShowFuncName sets whether to show the function name in the stack trace before the file and line.
//
// Parameters:
//   - state: true to show function names, false to hide them
func SetShowFuncName(state bool) {
	showFuncName = state
}

// SetShowPackageName sets whether to show the full function name (package name/function name)
// or just the base function name.
//
// Parameters:
//   - state: true to show full package path, false to show only function name
func SetShowPackageName(state bool) {
	showPackageName = state
}

// getCallerPath returns the file and line which called any of New functions as string.
//
// The returned string format depends on showFuncName and showPackageName settings:
//   - With showFuncName=true: "at function_name file:line"
//   - With showFuncName=false: "file:line"
//
// Parameters:
//   - skipFrames: number of stack frames to skip to find the caller
//
// Returns:
//   - A formatted string containing the caller's location
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

// stripPackageName removes the package path from a fully qualified function name,
// returning only the base function name.
//
// Example: "github.com/user/pkg.Function" becomes "Function"
func stripPackageName(name string) string {
	return path.Base(name)
}
