package errors

import (
	"runtime"
	"strings"
	"testing"
)

func TestSetShowFuncName(t *testing.T) {
	tests := []struct {
		name     string
		state    bool
		expected bool
	}{
		{"Enable function name display", true, true},
		{"Disable function name display", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetShowFuncName(tt.state)
			if showFuncName.Load() != tt.expected {
				t.Errorf("SetShowFuncName(%v) = %v, want %v", tt.state, showFuncName.Load(), tt.expected)
			}
		})
	}
}

func TestSetShowPackageName(t *testing.T) {
	tests := []struct {
		name     string
		state    bool
		expected bool
	}{
		{"Enable package name display", true, true},
		{"Disable package name display", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetShowPackageName(tt.state)
			if showPackageName.Load() != tt.expected {
				t.Errorf("SetShowPackageName(%v) = %v, want %v", tt.state, showPackageName.Load(), tt.expected)
			}
		})
	}
}

func TestGetCallerPath(t *testing.T) {
	tests := []struct {
		name            string
		skipFrames      int
		showFunc        bool
		showPackage     bool
		expectedPattern string
	}{
		{
			name:            "No function name",
			skipFrames:      0,
			showFunc:        false,
			showPackage:     false,
			expectedPattern: ".go:[0-9]+$",
		},
		{
			name:            "With function name, no package",
			skipFrames:      0,
			showFunc:        true,
			showPackage:     false,
			expectedPattern: "at TestGetCallerPath .go:[0-9]+$",
		},
		{
			name:            "With function name and package",
			skipFrames:      0,
			showFunc:        true,
			showPackage:     true,
			expectedPattern: "at github.com/difof/errors.TestGetCallerPath .go:[0-9]+$",
		},
		{
			name:            "Invalid skip frames",
			skipFrames:      1000,
			showFunc:        false,
			showPackage:     false,
			expectedPattern: "^<no source>$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetShowFuncName(tt.showFunc)
			SetShowPackageName(tt.showPackage)

			result := getCallerPath(tt.skipFrames)

			if !strings.Contains(result, ".go:") && tt.expectedPattern != "^<no source>$" {
				t.Errorf("getCallerPath(%d) = %q, expected to contain '.go:'", tt.skipFrames, result)
			}

			if tt.expectedPattern == "^<no source>$" && result != "<no source>" {
				t.Errorf("getCallerPath(%d) = %q, want '<no source>'", tt.skipFrames, result)
			}
		})
	}
}

func TestStripPackageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Full package path",
			input:    "github.com/user/pkg.Function",
			expected: "pkg.Function",
		},
		{
			name:     "No package path",
			input:    "Function",
			expected: "Function",
		},
		{
			name:     "Multiple dots",
			input:    "github.com.user.pkg.Function",
			expected: "github.com.user.pkg.Function",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripPackageName(tt.input)
			if result != tt.expected {
				t.Errorf("stripPackageName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRuntimeIntegration(t *testing.T) {
	// Test the interaction between different settings
	tests := []struct {
		name        string
		showFunc    bool
		showPackage bool
		validate    func(string) bool
	}{
		{
			name:        "Both disabled",
			showFunc:    false,
			showPackage: false,
			validate: func(s string) bool {
				return !strings.Contains(s, "at ") && strings.Contains(s, ".go:")
			},
		},
		{
			name:        "Only function enabled",
			showFunc:    true,
			showPackage: false,
			validate: func(s string) bool {
				return strings.Contains(s, "at ") &&
					strings.Contains(s, ".go:") &&
					!strings.Contains(s, "github.com")
			},
		},
		{
			name:        "Both enabled",
			showFunc:    true,
			showPackage: true,
			validate: func(s string) bool {
				return strings.Contains(s, "at ") &&
					strings.Contains(s, ".go:") &&
					strings.Contains(s, "testing.tRunner")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetShowFuncName(tt.showFunc)
			SetShowPackageName(tt.showPackage)

			result := getCallerPath(0)
			if !tt.validate(result) {
				t.Errorf("Integration test failed for %s, got: %s", tt.name, result)
			}
		})
	}
}

// Helper function to get current package name
func getCurrentPackage() string {
	pc, _, _, _ := runtime.Caller(0)
	return runtime.FuncForPC(pc).Name()
}
