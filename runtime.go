package errors

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
)

type ErrorConfig struct {
	ShowFuncName    bool
	ShowPackageName bool
	ShowFilePath    bool
}

var (
	errorConfig = ErrorConfig{
		ShowFuncName:    true,
		ShowPackageName: true,
		ShowFilePath:    true,
	}
	errorConfigLock = sync.RWMutex{}
)

type ErrorConfigOption func(config *ErrorConfig)

func WithShowFuncName(showFuncName bool) ErrorConfigOption {
	return func(config *ErrorConfig) {
		config.ShowFuncName = showFuncName
	}
}

func WithShowPackageName(showPackageName bool) ErrorConfigOption {
	return func(config *ErrorConfig) {
		config.ShowPackageName = showPackageName
	}
}

func WithShowFilePath(showFilePath bool) ErrorConfigOption {
	return func(config *ErrorConfig) {
		config.ShowFilePath = showFilePath
	}
}

func GetErrorConfig() ErrorConfig {
	errorConfigLock.RLock()
	defer errorConfigLock.RUnlock()
	return errorConfig
}

func SetErrorConfig(opts ...ErrorConfigOption) {
	errorConfigLock.Lock()
	defer errorConfigLock.Unlock()
	for _, opt := range opts {
		opt(&errorConfig)
	}
}

// getCallerPath returns the caller's source location with optional function name
func getCallerPath(skip int) string {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "<no source>"
	}

	config := GetErrorConfig()

	// If all settings are disabled, return an empty string
	if !config.ShowFuncName && !config.ShowPackageName && !config.ShowFilePath {
		return ""
	}

	var b strings.Builder

	// Only add "at " prefix if at least one setting is enabled
	if config.ShowFuncName || config.ShowPackageName || config.ShowFilePath {
		b.WriteString("at ")
	}

	// Add function name if enabled
	if config.ShowFuncName {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			name := fn.Name()
			// Check if we're in a test function
			if strings.Contains(name, "Test") {
				// For test functions, skip one more frame to get to testing.tRunner
				if pc2, _, _, ok := runtime.Caller(skip + 2); ok {
					if fn2 := runtime.FuncForPC(pc2); fn2 != nil {
						name = fn2.Name()
					}
				}
			}
			if !config.ShowPackageName {
				name = stripPackageName(name)
			}
			b.WriteString(name)
			b.WriteString(" ")
		}
	}

	// Add file path if enabled
	if config.ShowFilePath {
		// Check if this is called from a New function
		fn := runtime.FuncForPC(pc)
		if fn != nil && strings.Contains(fn.Name(), "New") {
			// For New functions, use the caller's location
			if pc, file, line, ok = runtime.Caller(skip + 2); ok {
				b.WriteString(fmt.Sprintf("%s:%d", path.Base(file), line))
			} else {
				b.WriteString("<no source>")
			}
		} else {
			// For other functions, use the current location
			b.WriteString(fmt.Sprintf("%s:%d", path.Base(file), line))
		}
	}

	// Return the result, trimming any trailing space
	return strings.TrimSpace(b.String())
}

// stripPackageName removes the package path from a fully qualified function name,
// returning only the base function name.
//
// Example: "github.com/user/pkg.Function" becomes "Function"
func stripPackageName(name string) string {
	return path.Base(name)
}
