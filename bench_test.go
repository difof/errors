package errors

import (
	"testing"
)

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

// createErrorChain creates a chain of errors of specified depth
func createErrorChain(depth int) error {
	if depth <= 0 {
		return New("base error")
	}
	return Wrap(createErrorChain(depth - 1))
}

func BenchmarkWrapFunctions(b *testing.B) {
	baseErr := createErrorChain(50) // medium depth error chain
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

func BenchmarkNewFunctions(b *testing.B) {
	benchmarks := []struct {
		name string
		fn   func()
	}{
		{
			name: "New",
			fn: func() {
				_ = New("test error")
			},
		},
		{
			name: "Newf",
			fn: func() {
				_ = Newf("formatted error: %d", 42)
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
