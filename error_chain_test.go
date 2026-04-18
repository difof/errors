package errors

import (
	stderrors "errors"
	"testing"
)

func TestErrorChainError(t *testing.T) {
	leaf := stderrors.New("boom")

	tests := []struct {
		name  string
		chain *ErrorChain
		want  string
	}{
		{
			name:  "formatted leaf",
			chain: newErrorChain(newErrorNode(0, "load %s", "user"), nil),
			want:  "load user",
		},
		{
			name:  "formatted wrap",
			chain: newErrorChain(newErrorNode(0, "load %s", "user"), leaf),
			want:  "load user: boom",
		},
		{
			name:  "message less wrap",
			chain: newErrorChain(newErrorNode(0, ""), leaf),
			want:  "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.chain.Error(); got != tt.want {
				t.Fatalf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorChainUnwrap(t *testing.T) {
	leaf := stderrors.New("boom")
	chain := newErrorChain(newErrorNode(0, "context"), leaf)

	if got := chain.Unwrap(); got != leaf {
		t.Fatalf("Unwrap() = %v, want %v", got, leaf)
	}
}
