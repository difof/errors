package errors

import (
	goerrors "errors"
	"testing"
)

func TestMustHelpers(t *testing.T) {
	t.Run("MustResult returns value on nil error", func(t *testing.T) {
		if got := MustResult(42, nil); got != 42 {
			t.Fatalf("MustResult(...) = %d, want 42", got)
		}
	})

	t.Run("MustResult panics with wrapped error", func(t *testing.T) {
		defer func() {
			r := recover()
			err, ok := r.(error)
			if !ok {
				t.Fatalf("panic = %T, want error", r)
			}
			if got := RootMessage(err); got != "boom" {
				t.Fatalf("RootMessage(panic) = %q, want %q", got, "boom")
			}
		}()
		_ = MustResult(0, goerrors.New("boom"))
	})

	t.Run("MustResultf panics with formatted context", func(t *testing.T) {
		defer func() {
			r := recover()
			err, ok := r.(error)
			if !ok {
				t.Fatalf("panic = %T, want error", r)
			}
			if got := ChainMessages(err); got != "load user 7: boom" {
				t.Fatalf("ChainMessages(panic) = %q, want %q", got, "load user 7: boom")
			}
		}()
		_ = MustResultf(0, goerrors.New("boom"))("load user %d", 7)
	})

	t.Run("MustResult2 returns values on nil error", func(t *testing.T) {
		a, b := MustResult2(1, 2, nil)
		if a != 1 || b != 2 {
			t.Fatalf("MustResult2(...) = (%d, %d), want (1, 2)", a, b)
		}
	})

	t.Run("MustResult2f panics with formatted context", func(t *testing.T) {
		defer func() {
			r := recover()
			err, ok := r.(error)
			if !ok {
				t.Fatalf("panic = %T, want error", r)
			}
			if got := ChainMessages(err); got != "pair lookup: boom" {
				t.Fatalf("ChainMessages(panic) = %q, want %q", got, "pair lookup: boom")
			}
		}()
		_, _ = MustResult2f(1, 2, goerrors.New("boom"))("pair lookup")
	})

	t.Run("Must does not panic on nil error", func(t *testing.T) {
		Must(nil)
	})

	t.Run("Must panics on error", func(t *testing.T) {
		defer func() {
			r := recover()
			err, ok := r.(error)
			if !ok {
				t.Fatalf("panic = %T, want error", r)
			}
			if got := RootMessage(err); got != "boom" {
				t.Fatalf("RootMessage(panic) = %q, want %q", got, "boom")
			}
		}()
		Must(goerrors.New("boom"))
	})

	t.Run("Mustf panics with formatted context", func(t *testing.T) {
		defer func() {
			r := recover()
			err, ok := r.(error)
			if !ok {
				t.Fatalf("panic = %T, want error", r)
			}
			if got := ChainMessages(err); got != "do work: boom" {
				t.Fatalf("ChainMessages(panic) = %q, want %q", got, "do work: boom")
			}
		}()
		Mustf(goerrors.New("boom"))("do %s", "work")
	})

	t.Run("Ignore returns value", func(t *testing.T) {
		if got := Ignore(42, goerrors.New("boom")); got != 42 {
			t.Fatalf("Ignore(...) = %d, want 42", got)
		}
	})

	t.Run("Assert panics when false", func(t *testing.T) {
		defer func() {
			if got := recover(); got != "boom" {
				t.Fatalf("recover() = %v, want %q", got, "boom")
			}
		}()
		Assert(false, "boom")
	})

	t.Run("Assertf panics when false", func(t *testing.T) {
		defer func() {
			if got := recover(); got != "boom 7" {
				t.Fatalf("recover() = %v, want %q", got, "boom 7")
			}
		}()
		Assertf(false, "%s %d", "boom", 7)
	})
}
