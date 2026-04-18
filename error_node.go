package errors

// errorNode stores the package-owned metadata for a chain or tree node.
type errorNode struct {
	pc     uintptr
	format string
	params []any
}

// newErrorNode creates an errorNode with an optional format and parameters.
func newErrorNode(pc uintptr, format string, params ...any) errorNode {
	return errorNode{
		pc:     pc,
		format: format, params: params,
	}
}
