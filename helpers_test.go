package errors

import goerrors "errors"

func test_util_createErrorChain(depth int, rootIsStdError bool) error {
	var err error = New("root cause error")
	if rootIsStdError {
		err = goerrors.New("root cause go-std-error")
	}

	for i := 1; i < depth; i++ {
		err = Wrapf(
			err,
			"error at depth %d", i,
		)
	}

	return err
}
