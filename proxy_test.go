package errors

import (
	stderrors "errors"
	"fmt"
	"testing"
)

type proxyTargetError struct{}

func (proxyTargetError) Error() string { return "target" }

func TestProxyHelpersMirrorStdlibBehavior(t *testing.T) {
	sentinel := stderrors.New("boom")
	wrapped := Wrap(sentinel)

	if !Is(wrapped, sentinel) {
		t.Fatal("Is() did not match wrapped sentinel")
	}

	if got := Unwrap(wrapped); got != sentinel {
		t.Fatalf("Unwrap() = %v, want %v", got, sentinel)
	}

	foreign := fmt.Errorf("wrap: %w", proxyTargetError{})
	var target proxyTargetError
	if !As(foreign, &target) {
		t.Fatal("As() did not match wrapped target error")
	}
}
