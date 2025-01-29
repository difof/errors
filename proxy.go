package errors

import goerrors "errors"

// As attempts to match the error against the target type.
// This is a wrapper around Go's standard errors.As function.
// It unwraps the error chain as needed looking for a match.
//
// Parameters:
//   - err: the error to examine
//   - target: pointer to the error type to match against
//
// Returns true if a match was found and target was set.
//
// Example:
//
//	var pathError *os.PathError
//	if errors.As(err, &pathError) {
//	    fmt.Println("Path:", pathError.Path)
//	}
func As(err error, target any) bool { return goerrors.As(err, target) }

// Is reports whether any error in err's chain matches target.
// This is a wrapper around Go's standard errors.Is function.
//
// Parameters:
//   - err: the error to examine
//   - target: the error to match against
//
// Example:
//
//	if errors.Is(err, os.ErrNotExist) {
//	    // handle not exists case
//	}
func Is(err, target error) bool { return goerrors.Is(err, target) }

// Unwrap returns the result of calling the Unwrap method on err,
// if err implements Unwrap. Otherwise returns nil.
// This is a wrapper around Go's standard errors.Unwrap function.
//
// Example:
//
//	innerErr := errors.Unwrap(err)
func Unwrap(err error) error { return goerrors.Unwrap(err) }
