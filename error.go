package errors

import (
	"fmt"
	"strings"
)

// Error is a lightweight drop-in replacement for standard errors package with stacktrace.
type Error struct {
	Source  string
	Message error
	Inner   error
}

func NewError(source string, message, inner error) *Error {
	return &Error{
		Source:  source,
		Message: message,
		Inner:   inner,
	}
}

// Each iterates over all inner errors of Error
func (e *Error) Each(it func(err error) bool) {
	if it == nil {
		return
	}

	var current error = e
	for current != nil {
		if !it(current) {
			break
		}

		var cast *Error
		if As(current, &cast) {
			current = cast.Unwrap()
		} else {
			current = nil
		}
	}
}

// StackTrace builds the stack trace of all inner errors of Error
func (e *Error) StackTrace() (list []string) {
	list = make([]string, 0, 5)

	defer func() {
		// reverse
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
	}()

	e.Each(func(err error) bool {
		var e *Error
		if As(err, &e) {
			list = append(list, e.String())
		} else {
			list = append(list, err.Error())
		}
		return true
	})

	return
}

// String returns current error's message and source
func (e *Error) String() string {
	if e.Message == nil {
		return e.Source
	}
	return fmt.Sprintf("%v: %v", e.Source, e.Message)
}

// Error returns the stack trace of this error
func (e *Error) Error() string {
	return strings.Join(e.StackTrace(), "\n")
}

// Unwrap returns the inner error
func (e *Error) Unwrap() error { return e.Inner }
