package errors

import "strings"

// ErrorTree represents one package-owned error node with multiple children.
type ErrorTree struct {
	node     errorNode
	children []error
}

// newErrorTree builds an ErrorTree from a node and child errors.
func newErrorTree(node errorNode, children []error) *ErrorTree {
	return &ErrorTree{node, children}
}

// Error returns the default joined message form for the tree.
func (e *ErrorTree) Error() string {
	errors := make([]string, 0, len(e.children))

	for _, err := range e.children {
		errors = append(errors, err.Error())
	}

	return strings.Join(errors, "\n")
}

// Unwrap returns all child errors.
func (e *ErrorTree) Unwrap() []error { return e.children }

// Join combines multiple errors into one package-owned multi error.
// Nil inputs are ignored. A single non-nil input is returned unchanged.
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
				return err
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
