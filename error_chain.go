package errors

import "fmt"

// ErrorChain represents one package-owned error node with at most one child.
type ErrorChain struct {
	node  errorNode
	child error
}

// newErrorChain builds an ErrorChain from a node and child error.
func newErrorChain(node errorNode, child error) *ErrorChain {
	return &ErrorChain{node, child}
}

// Error returns the default wrapped message form for the chain.
func (e *ErrorChain) Error() string {
	if e.node.format == "" {
		if e.child == nil {
			return ""
		}

		return e.child.Error()
	}

	base := fmt.Sprintf(e.node.format, e.node.params...)

	if e.child == nil {
		return base
	}

	return fmt.Sprintf("%s: %s", base, e.child.Error())
}

// Unwrap returns the next error in the chain.
func (e *ErrorChain) Unwrap() error { return e.child }
