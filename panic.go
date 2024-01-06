package errors

func mayPanicf(err error, format string, params ...any) {
	if err == nil {
		return
	}

	if format == "" {
		panic(WrapSkip(2, err))
	} else {
		panic(WrapSkipf(2, err, format, params...))
	}
}
