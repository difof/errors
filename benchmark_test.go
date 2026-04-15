package errors

import (
	"runtime"
	"testing"
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

// recursiveGetCallerPath creates a deep call stack of specified depth and then calls getCallerPath
func recursiveGetCallerPath(depth, skip int) (string, string, int) {
	if depth <= 0 {
		return getCallerPath(skip)
	}
	return recursiveGetCallerPath(depth-1, skip)
}

func BenchmarkGetCallerPath(b *testing.B) {
	benchmarks := []struct {
		name      string
		stackSize int
		skip      int
	}{
		{"Shallow/Skip0", 1, 0},
		{"Shallow/Skip1", 1, 1},
		{"Medium/Skip0", 50, 0},
		{"Medium/Skip1", 50, 1},
		{"Deep/Skip0", 100, 0},
		{"Deep/Skip1", 100, 1},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				funcname, filepath, line := recursiveGetCallerPath(bm.stackSize, bm.skip)
				// Prevent compiler optimization
				if funcname == "" && filepath == "" && line == 0 {
					b.Fatal("unexpected empty result")
				}
			}
		})
	}
}

func BenchmarkWrapFunctions(b *testing.B) {
	baseErr := test_util_createErrorChain(50, false) // medium depth error chain
	result := 42

	benchmarks := []struct {
		name string
		fn   func()
	}{
		{
			name: "Wrap",
			fn: func() {
				_ = Wrap(baseErr)
			},
		},
		{
			name: "Wrapf",
			fn: func() {
				_ = Wrapf(baseErr, "wrapped error: %v", baseErr)
			},
		},
		{
			name: "WrapResult",
			fn: func() {
				_, _ = WrapResult(result, baseErr)
			},
		},
		{
			name: "WrapResultf",
			fn: func() {
				_, _ = WrapResultf(result, baseErr)("wrapped error: %v", baseErr)
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				bm.fn()
			}
		})
	}
}

func BenchmarkErrorString(b *testing.B) {
	benchmarks := []struct {
		name      string
		chainSize int
	}{
		{"Shallow", 1},
		{"Medium", 50},
		{"Deep", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			err := test_util_createErrorChain(bm.chainSize, false)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				s := err.Error()
				// Prevent compiler optimization
				if len(s) == 0 {
					b.Fatal("unexpected empty error string")
				}
			}
		})
	}
}
