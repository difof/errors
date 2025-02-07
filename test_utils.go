package errors

import (
	goerrors "errors"
)

func createErrorChain(depth int, rootIsStdError bool) error {
	var err error = New("root cause error")
	if rootIsStdError {
		err = goerrors.New("root cause go-std-error")
	}

	for i := 1; i < depth; i++ {
		if i == 3 {
			err = Wrapf(
				createNestedError(err, i),
				"outer error",
			)
		} else {
			err = Wrapf(
				err,
				"error at depth %d", i,
			)
		}
	}

	return err
}

func createNestedError(chain error, depth int) error {
	return Wrapf(chain, "nested error at depth %d", depth)
}
