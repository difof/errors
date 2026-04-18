package errors

import (
	stderrors "errors"
	"testing"
)

func TestMustfWithRecoverReturnsWrappedError(t *testing.T) {
	boom := stderrors.New("boom")

	err := func() (err error) {
		defer Recover(&err)
		Mustf(boom)("do %s", "work")
		return nil
	}()

	if !stderrors.Is(err, boom) {
		t.Fatalf("Recover(Mustf(...)) does not wrap original: %v", err)
	}

	if got := err.Error(); got != "do work: boom" {
		t.Fatalf("Error() = %q, want %q", got, "do work: boom")
	}
}

func TestAssertfPanicsWithFormattedMessage(t *testing.T) {
	defer func() {
		if got := recover(); got != "boom 7" {
			t.Fatalf("recover() = %v, want %q", got, "boom 7")
		}
	}()

	Assertf(false, "%s %d", "boom", 7)
}
