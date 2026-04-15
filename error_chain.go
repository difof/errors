package errors

// ErrorChain is a linked list of errors with stacktrace context info.
//
// It's fully compatible with standard errors package,
// meaning you can use it as a drop-in replacement for standard errors package.
// It also means you can wrap it using fmt.Errorf(), errors.Join() and vice versa.
//
// Iterating over the chain will extract all wrapped errors.
type ErrorChain struct {
	// pc is the code location of the error at creation.
	pc uintptr

	// format is the format string of this error entry.
	format string

	// params is the parameters for the format string of this error entry.
	params []any

	// inner is the error interface attached to this error entry.
	inner error
}

// newErrorChain creates a new ErrorChain with the given parameters.
func newErrorChain(pc uintptr, inner error, format string, params ...any) *ErrorChain {
	return &ErrorChain{
		pc:     pc,
		format: format,
		params: params,
		inner:  inner,
	}
}

// JSON returns a detailed JSON representation of the error chain,
// including stack/source metadata when available.
func (e *ErrorChain) JSON() string {
	return JSONFormatter(DefaultJSONConfig()).FormatError(e)
}

// YAML returns a detailed YAML representation of the error chain,
// including stack/source metadata when available.
func (e *ErrorChain) YAML() string {
	return YAMLFormatter(DefaultYAMLConfig()).FormatError(e)
}

// Colored returns a detailed colored representation of the error chain
// for terminal output, including stack/source metadata when available.
func (e *ErrorChain) Colored() string {
	return ColoredFormatter(DefaultColorConfig()).FormatError(e)
}

// Error implements the error interface and returns the default detailed
// text rendering of this error, including stack/source metadata.
func (e *ErrorChain) Error() string {
	return TextFormatter(DefaultTextConfig()).FormatError(e)
}

// Unwrap returns the inner error, implementing the errors.Unwrap interface
func (e *ErrorChain) Unwrap() error { return e.inner }

// HasErrorChain returns true if the error is an ErrorChain
func HasErrorChain(err error) bool {
	return Is(err, &ErrorChain{})
}
