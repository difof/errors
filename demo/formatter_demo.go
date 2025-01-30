package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/difof/errors"
)

// createDeepError creates a deeply nested error chain to demonstrate stack traces
func createDeepError(depth int) error {
	_, file, line, _ := runtime.Caller(0)
	if depth == 0 {
		return errors.NewError(fmt.Sprintf("%s:%d", file, line), fmt.Errorf("root cause error"), nil)
	}

	innerErr := createDeepError(depth - 1)
	return errors.NewError(
		fmt.Sprintf("%s:%d", file, line),
		fmt.Errorf("error at depth %d", depth),
		innerErr,
	)
}

func main() {
	// Create a deep error chain
	err := createDeepError(5)
	if err == nil {
		fmt.Println("No error occurred!")
		os.Exit(0)
	}

	// Cast to our Error type
	e := err.(*errors.Error)

	// Print error in different formats
	fmt.Println("=== Default Format ===")
	fmt.Println(e.Error())
	fmt.Println()

	fmt.Println("=== Colored Format (for terminals) ===")
	fmt.Println(e.Colored())
	fmt.Println()

	fmt.Println("=== JSON Format ===")
	fmt.Println(e.JSON())
	fmt.Println()

	fmt.Println("=== YAML Format ===")
	fmt.Println(e.YAML())
	fmt.Println()

	// Print just the error message without the stack trace
	fmt.Println("=== Error Message Only ===")
	fmt.Println(e.ErrorMessage())
}
