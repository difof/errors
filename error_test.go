package errors

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestNewError(t *testing.T) {
	tests := []struct {
		name     string
		funcPath string
		filePath string
		line     int
		message  error
		inner    error
		want     *Error
	}{
		{
			name:     "basic error without inner",
			funcPath: "pkg.func",
			filePath: "file.go",
			line:     42,
			message:  errors.New("test error"),
			inner:    nil,
			want: &Error{
				FuncPath: "pkg.func",
				FilePath: "file.go",
				Line:     42,
				Message:  errors.New("test error"),
				Inner:    nil,
			},
		},
		{
			name:     "error with inner error",
			funcPath: "pkg.func",
			filePath: "file.go",
			line:     42,
			message:  errors.New("outer error"),
			inner:    errors.New("inner error"),
			want: &Error{
				FuncPath: "pkg.func",
				FilePath: "file.go",
				Line:     42,
				Message:  errors.New("outer error"),
				Inner:    errors.New("inner error"),
			},
		},
		{
			name:     "nil message",
			funcPath: "pkg.func",
			filePath: "file.go",
			line:     42,
			message:  nil,
			inner:    nil,
			want: &Error{
				FuncPath: "pkg.func",
				FilePath: "file.go",
				Line:     42,
				Message:  nil,
				Inner:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewError(tt.funcPath, tt.filePath, tt.line, tt.message, tt.inner)
			assert.Equal(t, tt.want.FuncPath, got.FuncPath)
			assert.Equal(t, tt.want.FilePath, got.FilePath)
			assert.Equal(t, tt.want.Line, got.Line)

			if tt.want.Message != nil {
				assert.Equal(t, tt.want.Message.Error(), got.Message.Error())
			} else {
				assert.Nil(t, got.Message)
			}

			if tt.want.Inner != nil {
				assert.Equal(t, tt.want.Inner.Error(), got.Inner.Error())
			} else {
				assert.Nil(t, got.Inner)
			}
		})
	}
}

func TestErrorMessageOf(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
		{
			name: "standard error",
			err:  errors.New("standard error"),
			want: "standard error",
		},
		{
			name: "custom error without inner",
			err:  NewError("pkg.func", "file.go", 42, errors.New("test error"), nil),
			want: "test error",
		},
		{
			name: "custom error with inner",
			err:  NewError("pkg.func", "file.go", 42, errors.New("outer error"), errors.New("inner error")),
			want: "inner error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorMessageOf(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestError_ExtractEntries(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		wantLen  int
		wantLast Error
	}{
		{
			name:    "single error",
			err:     NewError("pkg.func", "file.go", 42, errors.New("test error"), nil),
			wantLen: 1,
			wantLast: Error{
				FuncPath: "pkg.func",
				FilePath: "file.go",
				Line:     42,
				Message:  errors.New("test error"),
			},
		},
		{
			name: "nested errors",
			err: NewError("pkg.func1", "file1.go", 42,
				errors.New("error1"),
				NewError("pkg.func2", "file2.go", 24, errors.New("error2"), nil),
			),
			wantLen: 2,
			wantLast: Error{
				FuncPath: "pkg.func2",
				FilePath: "file2.go",
				Line:     24,
				Message:  errors.New("error2"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := tt.err.ExtractEntries()
			assert.Equal(t, tt.wantLen, len(entries))
			assert.Equal(t, tt.wantLast.FuncPath, entries[len(entries)-1].FuncPath)
			assert.Equal(t, tt.wantLast.FilePath, entries[len(entries)-1].FilePath)
			assert.Equal(t, tt.wantLast.Line, entries[len(entries)-1].Line)
			assert.Equal(t, tt.wantLast.Message.Error(), entries[len(entries)-1].Message)
		})
	}
}

func TestError_Each(t *testing.T) {
	tests := []struct {
		name      string
		err       *Error
		wantCount int
		stopAfter int // number of iterations after which to stop
	}{
		{
			name:      "nil callback",
			err:       NewError("pkg.func", "file.go", 42, errors.New("test error"), nil),
			wantCount: 0,
			stopAfter: -1,
		},
		{
			name:      "single error",
			err:       NewError("pkg.func", "file.go", 42, errors.New("test error"), nil),
			wantCount: 1,
			stopAfter: -1,
		},
		{
			name: "nested errors",
			err: NewError("pkg.func1", "file1.go", 42,
				errors.New("error1"),
				NewError("pkg.func2", "file2.go", 24, errors.New("error2"), nil),
			),
			wantCount: 2,
			stopAfter: -1,
		},
		{
			name: "stop iteration early",
			err: NewError("pkg.func1", "file1.go", 42,
				errors.New("error1"),
				NewError("pkg.func2", "file2.go", 24, errors.New("error2"), nil),
			),
			wantCount: 1,
			stopAfter: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			if tt.name == "nil callback" {
				tt.err.Each(nil)
				assert.Equal(t, tt.wantCount, count)
			} else {
				tt.err.Each(func(err error) bool {
					count++
					return count != tt.stopAfter
				})
				assert.Equal(t, tt.wantCount, count)
			}
		})
	}
}

func TestError_ErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "single error",
			err:  NewError("pkg.func", "file.go", 42, errors.New("test error"), nil),
			want: "test error",
		},
		{
			name: "nested errors",
			err: NewError("pkg.func1", "file1.go", 42,
				errors.New("error1"),
				NewError("pkg.func2", "file2.go", 24, errors.New("error2"), nil),
			),
			want: "error2",
		},
		{
			name: "standard error as inner",
			err: NewError("pkg.func1", "file1.go", 42,
				errors.New("error1"),
				errors.New("standard error"),
			),
			want: "standard error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.ErrorMessage()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	inner := errors.New("inner error")
	err := NewError("pkg.func", "file.go", 42, errors.New("test error"), inner)

	assert.Equal(t, inner, err.Unwrap())
	assert.Nil(t, NewError("pkg.func", "file.go", 42, errors.New("test error"), nil).Unwrap())
}

func TestError_Is(t *testing.T) {
	target := errors.New("target error")
	tests := []struct {
		name   string
		err    *Error
		target error
		want   bool
	}{
		{
			name:   "nil target",
			err:    NewError("pkg.func", "file.go", 42, errors.New("test error"), nil),
			target: nil,
			want:   false,
		},
		{
			name:   "matching message",
			err:    NewError("pkg.func", "file.go", 42, target, nil),
			target: target,
			want:   true,
		},
		{
			name:   "matching inner",
			err:    NewError("pkg.func", "file.go", 42, errors.New("test error"), target),
			target: target,
			want:   true,
		},
		{
			name:   "no match",
			err:    NewError("pkg.func", "file.go", 42, errors.New("test error"), errors.New("inner error")),
			target: target,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Is(tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

type customError struct{ msg string }

func (e *customError) Error() string { return e.msg }

func TestError_As(t *testing.T) {
	var target *customError

	tests := []struct {
		name   string
		err    *Error
		target interface{}
		want   bool
	}{
		{
			name:   "matching message type",
			err:    NewError("pkg.func", "file.go", 42, &customError{msg: "test error"}, nil),
			target: &target,
			want:   true,
		},
		{
			name:   "matching inner type",
			err:    NewError("pkg.func", "file.go", 42, errors.New("test error"), &customError{msg: "inner error"}),
			target: &target,
			want:   true,
		},
		{
			name:   "no match",
			err:    NewError("pkg.func", "file.go", 42, errors.New("test error"), errors.New("inner error")),
			target: &target,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.As(tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestError_Formatters(t *testing.T) {
	err := NewError("pkg.func", "file.go", 42, errors.New("test error"), nil)

	t.Run("JSON", func(t *testing.T) {
		jsonStr := err.JSON()
		var parsed []map[string]interface{}
		assert.NoError(t, json.Unmarshal([]byte(jsonStr), &parsed))
		assert.Equal(t, 1, len(parsed))
		assert.Equal(t, "pkg.func", parsed[0]["funcpath"])
	})

	t.Run("YAML", func(t *testing.T) {
		yamlStr := err.YAML()
		var parsed []map[string]interface{}
		assert.NoError(t, yaml.Unmarshal([]byte(yamlStr), &parsed))
		assert.Equal(t, 1, len(parsed))
		assert.Equal(t, "pkg.func", parsed[0]["funcpath"])
	})

	t.Run("Colored", func(t *testing.T) {
		colored := err.Colored()
		assert.Contains(t, colored, "pkg.func")
		assert.Contains(t, colored, "file.go")
	})

	t.Run("Error", func(t *testing.T) {
		errStr := err.Error()
		assert.Contains(t, errStr, "pkg.func")
		assert.Contains(t, errStr, "file.go")
		assert.Contains(t, errStr, "test error")
	})
}

func TestError_DeepStdError(t *testing.T) {
	// Create a standard error as root cause
	rootCause := errors.New("root cause error")

	// Build a deep error chain with 3 levels
	level3 := NewError("pkg.level3", "level3.go", 30, errors.New("level 3 error"), rootCause)
	level2 := NewError("pkg.level2", "level2.go", 20, errors.New("level 2 error"), level3)
	level1 := NewError("pkg.level1", "level1.go", 10, errors.New("level 1 error"), level2)

	// Test error message (should be root cause)
	assert.Equal(t, "root cause error", level1.ErrorMessage())

	// Test error entries
	entries := level1.ExtractEntries()
	assert.Equal(t, 4, len(entries))

	// Verify each level's details
	assert.Equal(t, "pkg.level1", entries[0].FuncPath)
	assert.Equal(t, "level1.go", entries[0].FilePath)
	assert.Equal(t, 10, entries[0].Line)
	assert.Equal(t, "level 1 error", entries[0].Message)

	assert.Equal(t, "pkg.level2", entries[1].FuncPath)
	assert.Equal(t, "level2.go", entries[1].FilePath)
	assert.Equal(t, 20, entries[1].Line)
	assert.Equal(t, "level 2 error", entries[1].Message)

	assert.Equal(t, "pkg.level3", entries[2].FuncPath)
	assert.Equal(t, "level3.go", entries[2].FilePath)
	assert.Equal(t, 30, entries[2].Line)
	assert.Equal(t, "level 3 error", entries[2].Message)

	// Test error string contains all levels
	errStr := level1.Error()
	assert.Contains(t, errStr, "level 1 error")
	assert.Contains(t, errStr, "level 2 error")
	assert.Contains(t, errStr, "level 3 error")

	// Test unwrapping to root cause
	var current error = level1
	var levels []string
	for current != nil {
		switch e := current.(type) {
		case *Error:
			levels = append(levels, e.Message.Error())
			current = e.Unwrap()
		default:
			levels = append(levels, e.Error())
			current = nil
		}
	}

	assert.Equal(t, []string{
		"level 1 error",
		"level 2 error",
		"level 3 error",
		"root cause error",
	}, levels)

	// Test Is works with root cause
	assert.True(t, errors.Is(level1, rootCause))
}
