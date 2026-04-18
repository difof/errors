package errors

import (
	stderrors "errors"
	"testing"
)

func TestRecoverWrapsNonErrorPanics(t *testing.T) {
	err := func() (err error) {
		defer Recover(&err)
		panic("boom")
	}()

	if err == nil {
		t.Fatal("Recover returned nil error")
	}

	if got := err.Error(); got != "boom" {
		t.Fatalf("Error() = %q, want %q", got, "boom")
	}

	entry := Expand(err)
	if entry == nil || entry.Resolved.Foreign {
		t.Fatalf("Expand(Recover string panic) = %#v, want package-owned entry", entry)
	}
}

func TestRecoverFnReceivesRecoveredError(t *testing.T) {
	boom := stderrors.New("boom")
	var got error

	func() {
		defer RecoverFn(func(err error) {
			got = err
		})
		panic(boom)
	}()

	if !stderrors.Is(got, boom) {
		t.Fatalf("RecoverFn received %v, want wrapped original error", got)
	}
}

func TestRecoverNilPointerLeavesPanicUntouched(t *testing.T) {
	boom := stderrors.New("boom")

	defer func() {
		if got := recover(); got != boom {
			t.Fatalf("recover() = %v, want original panic %v", got, boom)
		}
	}()

	func() {
		defer Recover(nil)
		panic(boom)
	}()
}

func TestRecoverFnDoesNotCallCallbackWithoutPanic(t *testing.T) {
	called := false

	func() {
		defer RecoverFn(func(err error) {
			called = true
		})
	}()

	if called {
		t.Fatal("RecoverFn callback was called without panic")
	}
}
