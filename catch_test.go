package errors

import "testing"

func TestCatchResult(t *testing.T) {
	err := Catchf(
		func() error {
			return New("test error")
		}(),
		"test error",
	)

	t.Log(err)
}
