package errors

import (
	"testing"
)

func TestErrorEntry_Collapse(t *testing.T) {
	err := createErrorChain(5, false)
	_ = err.Error()
}
