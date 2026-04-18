package errors

// New creates a package-owned leaf error and records the caller location.
//
// Example:
//
//	return errors.New("connection failed")
func New(msg string) error {
	return NewSkip(2, msg)
}

// Newf is like New but formats the message.
//
// Example:
//
//	return errors.Newf("failed to connect to %s: %v", host, err)
func Newf(format string, params ...any) error {
	return NewSkipf(2, format, params...)
}

// NewSkip is like New but skips additional stack frames before recording the callsite.
func NewSkip(skip int, msg string) error {
	node := newErrorNode(getCallerPC(skip), msg)
	return newErrorChain(node, nil)
}

// NewSkipf is like Newf but skips additional stack frames before recording the callsite.
func NewSkipf(skip int, format string, params ...any) error {
	node := newErrorNode(getCallerPC(skip), format, params...)
	return newErrorChain(node, nil)
}
