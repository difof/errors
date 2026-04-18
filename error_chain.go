package errors

import "fmt"

type ErrorChain struct {
	node  errorNode
	child error
}

func newErrorChain(node errorNode, child error) *ErrorChain {
	return &ErrorChain{node, child}
}

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

func (e *ErrorChain) Unwrap() error { return e.child }
