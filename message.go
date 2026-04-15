package errors

import (
	goerrors "errors"
	"fmt"
	"strings"
)

type unwrapperMany interface {
	Unwrap() []error
}

// RootMessage returns the innermost error message without stack/source metadata.
func RootMessage(err error) string {
	return rootMessage(err)
}

// ChainMessages returns the wrapped error chain as a stackless message string.
func ChainMessages(err error) string {
	return chainMessages(err)
}

func rootMessage(err error) string {
	if err == nil {
		return ""
	}

	if _, ok := err.(unwrapperMany); ok {
		return err.Error()
	}

	if inner := goerrors.Unwrap(err); inner != nil {
		return rootMessage(inner)
	}

	if ec, ok := err.(*ErrorChain); ok {
		if msg := errorChainMessage(ec); msg != "" {
			return msg
		}
	}

	return err.Error()
}

func chainMessages(err error) string {
	if err == nil {
		return ""
	}

	if ec, ok := err.(*ErrorChain); ok {
		return chainMessagesFromErrorChain(ec)
	}

	if _, ok := err.(unwrapperMany); ok {
		return err.Error()
	}

	if inner := goerrors.Unwrap(err); inner != nil {
		full := err.Error()
		innerFull := inner.Error()
		innerStackless := chainMessages(inner)

		if innerFull != "" && strings.Contains(full, innerFull) {
			return strings.Replace(full, innerFull, innerStackless, 1)
		}

		return full
	}

	return err.Error()
}

func chainMessagesFromErrorChain(err *ErrorChain) string {
	local := errorChainMessage(err)
	child := chainMessages(err.inner)

	switch {
	case local == "":
		return child
	case child == "":
		return local
	default:
		return local + ": " + child
	}
}

func errorChainMessage(err *ErrorChain) string {
	if err == nil || err.format == "" {
		return ""
	}

	if len(err.params) == 0 {
		return err.format
	}

	return fmt.Sprintf(err.format, err.params...)
}
