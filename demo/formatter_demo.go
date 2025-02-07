package main

import (
	goerrors "errors"
	"fmt"
	"os"

	"github.com/difof/errors"
	"github.com/difof/errors/demo/nested_package"
)

// createErrorChain creates a deeply nested error chain to demonstrate stack traces
func createErrorChain(depth int) error {
	// var err error = errors.New("root cause error")
	var err error = goerrors.New("root cause go-std-error")

	for i := 1; i < depth; i++ {
		if i == 3 {
			err = errors.Wrapf(
				nested_package.CreateNestedError(err, i),
				"outer error",
			)
		} else {
			err = errors.Wrapf(
				err,
				"error at depth %d", i,
			)
		}
	}

	return err
}

func main() {
	// Create a deep error chain
	err := errors.Wrapf(
		errors.Wrapf(
			fmt.Errorf("std wrapped error %w", createErrorChain(5)),
			"wrapped error 2",
		),
		"wrapped error 1",
	)

	if err == nil {
		fmt.Println("No error occurred!")
		os.Exit(0)
	}

	// config
	showColor := os.Getenv("SHOW_COLOR") == "true"
	showJSON := os.Getenv("SHOW_JSON") == "true"
	showYAML := os.Getenv("SHOW_YAML") == "true"
	showAll := os.Getenv("SHOW_ALL") == "true"

	if showAll {
		showColor = true
		showJSON = true
		showYAML = true
	}

	// Print error in different formats
	fmt.Println("=== Default Format ===")
	fmt.Println(err.Error())
	fmt.Println()

	if showColor {
		fmt.Println("=== Colored Format (for terminals) ===")
		fmt.Println(err.(*errors.ErrorChain).Colored())
		fmt.Println()
	}

	if showJSON {
		fmt.Println("=== JSON Format ===")
		fmt.Println(err.(*errors.ErrorChain).JSON())
		fmt.Println()
	}

	if showYAML {
		fmt.Println("=== YAML Format ===")
		fmt.Println(err.(*errors.ErrorChain).YAML())
		fmt.Println()
	}
}
