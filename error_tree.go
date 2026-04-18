package errors

import "strings"

type ErrorTree struct {
	node     errorNode
	children []error
}

func newErrorTree(node errorNode, children []error) *ErrorTree {
	return &ErrorTree{node, children}
}

func (e *ErrorTree) Error() string {
	errors := make([]string, 0, len(e.children))

	for _, err := range e.children {
		errors = append(errors, err.Error())
	}

	return strings.Join(errors, "\n")
}

func (e *ErrorTree) Unwrap() []error { return e.children }

func Join(errors ...error) error {
	n := 0

	for _, err := range errors {
		if err != nil {
			n++
		}
	}

	if n == 0 {
		return nil
	}

	if n == 1 {
		for _, err := range errors {
			if err != nil {
				if _, ok := IsUnwrapMulti(err); ok {
					return err
				}
			}
		}
	}

	children := make([]error, 0, n)
	for _, err := range errors {
		if err != nil {
			children = append(children, err)
		}
	}

	node := newErrorNode(getCallerPC(1), "")
	return newErrorTree(node, children)
}
