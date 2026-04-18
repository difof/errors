package errors

import (
	stderrors "errors"
	"testing"
)

func TestCatchResultShortCircuitsInputError(t *testing.T) {
	boom := stderrors.New("boom")
	called := false

	err := CatchResult(10, boom)(func(v int) error {
		called = true
		return nil
	})

	if called {
		t.Fatal("callback was called despite input error")
	}

	if !stderrors.Is(err, boom) {
		t.Fatalf("CatchResult error does not wrap original: %v", err)
	}

	entry := Expand(err)
	if entry == nil || entry.Resolved.Foreign {
		t.Fatalf("Expand(CatchResult(...)) = %#v, want package-owned entry", entry)
	}

	if len(entry.Children) != 1 || entry.Children[0].Resolved.Message != "boom" {
		t.Fatalf("Expand(CatchResult(...)) children = %#v", entry.Children)
	}
}

func TestCatchResultfWrapsCallbackErrorWithContext(t *testing.T) {
	callbackErr := stderrors.New("callback failed")

	err := CatchResultf(10, nil)(func(v int) error {
		if v != 10 {
			t.Fatalf("callback received %d, want 10", v)
		}

		return callbackErr
	}, "query user %d", 7)

	if !stderrors.Is(err, callbackErr) {
		t.Fatalf("CatchResultf error does not wrap callback error: %v", err)
	}

	if got := err.Error(); got != "query user 7: callback failed" {
		t.Fatalf("Error() = %q, want %q", got, "query user 7: callback failed")
	}
}

func TestCatchfPreservesEscapedPercentInContext(t *testing.T) {
	err := Catchf(stderrors.New("boom"), "save is 100%% done")

	if got := err.Error(); got != "save is 100% done: boom" {
		t.Fatalf("Error() = %q, want %q", got, "save is 100% done: boom")
	}
}
