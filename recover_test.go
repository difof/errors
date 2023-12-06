package errors

import (
	"fmt"
	"testing"
)

func recoverableMustError() error {
	return New("must fail")
}

func recoverable() (err error) {
	defer Recover(&err)

	Mustf(recoverableMustError())("this must cause death, but it didn't")

	return
}

func TestRecover(t *testing.T) {
	err := recoverable()
	Assert(err != nil, "err should not be nil")

	t.Log("\n\b", err)
}

func recoverable2() {
	defer RecoverFn(func(err error) {
		fmt.Printf("recovered from: %v\n", err)
	})

	defer Must(New("some error"))
}

func TestRecoverFn(t *testing.T) {
	recoverable2()
}
