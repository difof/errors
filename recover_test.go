package errors

import "testing"

func recoverableMustError() (int, error) {
	return 0, New("must fail")
}

func recoverable() (err error) {
	defer Recover(&err)

	// Must(recoverableMustError())
	Mustf(recoverableMustError())("this must cause death, but it didn't")

	return
}

func TestRecover(t *testing.T) {
	err := recoverable()
	Assert(err != nil, "err should not be nil")

	t.Log("\n\b", err)
}
