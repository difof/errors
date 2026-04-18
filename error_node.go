package errors

type errorNode struct {
	pc     uintptr
	format string
	params []any
}

func newErrorNode(pc uintptr, format string, params ...any) errorNode {
	return errorNode{
		pc:     pc,
		format: format, params: params,
	}
}
