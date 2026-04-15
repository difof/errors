package errors

import (
	"testing"
)

func TestErrorEntry_Collapse(t *testing.T) {
	err := test_util_createErrorChain(5, false)
	_ = err.Error()
}
