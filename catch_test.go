package errors

import (
	goerrors "errors"
	"testing"
)

func TestCatchHelpers(t *testing.T) {
	t.Run("Catch returns nil for nil input", func(t *testing.T) {
		if err := Catch(nil); err != nil {
			t.Fatalf("Catch(nil) = %v, want nil", err)
		}
	})

	t.Run("Catch wraps error without changing root message", func(t *testing.T) {
		err := Catch(goerrors.New("boom"))
		if got := RootMessage(err); got != "boom" {
			t.Fatalf("RootMessage(Catch(...)) = %q, want %q", got, "boom")
		}
	})

	t.Run("Catchf adds context", func(t *testing.T) {
		err := Catchf(goerrors.New("boom"), "while saving user %d", 42)
		if got := ChainMessages(err); got != "while saving user 42: boom" {
			t.Fatalf("ChainMessages(Catchf(...)) = %q, want %q", got, "while saving user 42: boom")
		}
	})

	t.Run("IgnoreResult ignores input", func(t *testing.T) {
		if err := IgnoreResult[int]()(123); err != nil {
			t.Fatalf("IgnoreResult callback returned %v, want nil", err)
		}
	})

	t.Run("CatchResult short circuits original error", func(t *testing.T) {
		called := false
		err := CatchResult(10, goerrors.New("boom"))(func(v int) error {
			called = true
			return nil
		})
		if called {
			t.Fatal("CatchResult callback was called despite input error")
		}
		if got := RootMessage(err); got != "boom" {
			t.Fatalf("RootMessage(CatchResult(...)) = %q, want %q", got, "boom")
		}
	})

	t.Run("CatchResult wraps callback error", func(t *testing.T) {
		err := CatchResult(10, nil)(func(v int) error {
			if v != 10 {
				t.Fatalf("callback received %d, want 10", v)
			}
			return goerrors.New("callback failed")
		})
		if got := RootMessage(err); got != "callback failed" {
			t.Fatalf("RootMessage(CatchResult callback) = %q, want %q", got, "callback failed")
		}
	})

	t.Run("CatchResult returns nil on success", func(t *testing.T) {
		if err := CatchResult(10, nil)(func(v int) error { return nil }); err != nil {
			t.Fatalf("CatchResult success = %v, want nil", err)
		}
	})

	t.Run("CatchResultf wraps original error with format", func(t *testing.T) {
		called := false
		err := CatchResultf(10, goerrors.New("boom"))(func(v int) error {
			called = true
			return nil
		}, "operation %s", "save")
		if called {
			t.Fatal("CatchResultf callback was called despite input error")
		}
		if got := ChainMessages(err); got != "operation save: boom" {
			t.Fatalf("ChainMessages(CatchResultf input err) = %q, want %q", got, "operation save: boom")
		}
	})

	t.Run("CatchResultf wraps callback error with format", func(t *testing.T) {
		err := CatchResultf(10, nil)(func(v int) error {
			return goerrors.New("callback failed")
		}, "operation %s", "save")
		if got := ChainMessages(err); got != "operation save: callback failed" {
			t.Fatalf("ChainMessages(CatchResultf callback err) = %q, want %q", got, "operation save: callback failed")
		}
	})
}
