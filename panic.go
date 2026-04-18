package errors

// mayPanicf panics with err wrapped as a package-owned error.
// When format is not empty, the panic value includes formatted context.
func mayPanicf(skip int, err error, format string, params ...any) {
	if err == nil {
		return
	}

	if format == "" {
		panic(WrapSkip(skip+1, err))
	} else {
		panic(WrapSkipf(skip+1, err, format, params...))
	}
}
