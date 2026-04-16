package errors

import (
	goerrors "errors"
	"testing"
)

func TestWrapHelpers(t *testing.T) {
	t.Run("WrapResult returns nil error when input nil", func(t *testing.T) {
		got, err := WrapResult(42, nil)
		if got != 42 || err != nil {
			t.Fatalf("WrapResult(...) = (%d, %v), want (42, nil)", got, err)
		}
	})

	t.Run("WrapResult wraps error", func(t *testing.T) {
		got, err := WrapResult(42, goerrors.New("boom"))
		if got != 42 {
			t.Fatalf("WrapResult value = %d, want 42", got)
		}
		if gotMsg := RootMessage(err); gotMsg != "boom" {
			t.Fatalf("RootMessage(WrapResult err) = %q, want %q", gotMsg, "boom")
		}
	})

	t.Run("WrapResultf returns nil error when input nil", func(t *testing.T) {
		got, err := WrapResultf(42, nil)("ignored")
		if got != 42 || err != nil {
			t.Fatalf("WrapResultf(...) = (%d, %v), want (42, nil)", got, err)
		}
	})

	t.Run("WrapResultf adds context", func(t *testing.T) {
		got, err := WrapResultf(42, goerrors.New("boom"))("load %s", "user")
		if got != 42 {
			t.Fatalf("WrapResultf value = %d, want 42", got)
		}
		if gotMsg := ChainMessages(err); gotMsg != "load user: boom" {
			t.Fatalf("ChainMessages(WrapResultf err) = %q, want %q", gotMsg, "load user: boom")
		}
	})

	t.Run("WrapSkip returns nil for nil error", func(t *testing.T) {
		if err := WrapSkip(0, nil); err != nil {
			t.Fatalf("WrapSkip(nil) = %v, want nil", err)
		}
	})

	t.Run("WrapSkipf returns nil for nil error and empty format", func(t *testing.T) {
		if err := WrapSkipf(0, nil, ""); err != nil {
			t.Fatalf("WrapSkipf(nil, empty) = %v, want nil", err)
		}
	})

	t.Run("WrapSkipf creates message even when inner error is nil", func(t *testing.T) {
		err := WrapSkipf(0, nil, "hello %s", "world")
		if got := ChainMessages(err); got != "hello world" {
			t.Fatalf("ChainMessages(WrapSkipf(...)) = %q, want %q", got, "hello world")
		}
	})
}
