package errors

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCallerPath(t *testing.T) {
	// Helper function to test getCallerPath at different levels
	testFunc := func() (string, string, int) {
		return getCallerPath(0)
	}

	// Nested function to test multiple stack levels
	nestedTestFunc := func() (string, string, int) {
		return testFunc()
	}

	tests := []struct {
		name         string
		skipLevel    int
		caller       func() (string, string, int)
		wantFuncName string
		wantFilePath string
		wantLine     int
		wantContains bool // true if we want to check contains instead of exact match
	}{
		{
			name:         "direct call with skip 0",
			skipLevel:    0,
			caller:       func() (string, string, int) { return getCallerPath(0) },
			wantFuncName: "github.com/difof/errors.TestGetCallerPath.func",
			wantFilePath: "runtime_test.go",
			wantLine:     0, // Line number will vary, we'll check if > 0
			wantContains: true,
		},
		{
			name:         "call through helper function",
			skipLevel:    0,
			caller:       testFunc,
			wantFuncName: "github.com/difof/errors.TestGetCallerPath",
			wantFilePath: "runtime_test.go",
			wantLine:     0, // Line number will vary, we'll check if > 0
			wantContains: true,
		},
		{
			name:         "nested call",
			skipLevel:    0,
			caller:       nestedTestFunc,
			wantFuncName: "github.com/difof/errors.TestGetCallerPath",
			wantFilePath: "runtime_test.go",
			wantLine:     0, // Line number will vary, we'll check if > 0
			wantContains: true,
		},
		{
			name:         "invalid skip level",
			skipLevel:    1000, // Very high skip level to force failure
			caller:       func() (string, string, int) { return getCallerPath(1000) },
			wantFuncName: "",
			wantFilePath: NO_SOURCE,
			wantLine:     0,
			wantContains: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcName, filePath, line := tt.caller()

			if tt.wantContains {
				if tt.name == "direct call with skip 0" {
					// For direct call, just check if it contains the base pattern
					assert.True(t, strings.Contains(funcName, "github.com/difof/errors.TestGetCallerPath.func"),
						"Expected function name to contain %q, got %q",
						"github.com/difof/errors.TestGetCallerPath.func",
						funcName)
				} else {
					assert.Contains(t, funcName, tt.wantFuncName)
				}
				assert.Contains(t, filePath, tt.wantFilePath)
				assert.Greater(t, line, 0)
			} else {
				assert.Equal(t, tt.wantFuncName, funcName)
				assert.Equal(t, tt.wantFilePath, filePath)
				assert.Equal(t, tt.wantLine, line)
			}
		})
	}
}
