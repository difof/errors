package main

import (
	goerrors "errors"
	"flag"
	"fmt"
	"strings"

	"github.com/fatih/color"

	"github.com/difof/errors"
	"github.com/difof/errors/demo/nested_package"
)

type demoConfig struct {
	full       bool
	error      bool
	stacktrace bool
	colored    bool
}

type demoCase struct {
	title       string
	description string
	build       string
	err         error
}

var (
	headerColor = color.New(color.FgHiCyan, color.Bold)
	lineColor   = color.New(color.FgHiBlack)
	labelColor  = color.New(color.FgHiYellow, color.Bold)
)

// createErrorChain creates a deeply nested error chain to demonstrate stack traces.
func createErrorChain(depth int) error {
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
	cfg := parseFlags()

	for i, demo := range demoCases() {
		printSectionHeader(fmt.Sprintf("Demo %d: %s", i+1, demo.title))
		fmt.Println(demo.description)
		fmt.Println()
		printOutputs(cfg, demo)
	}
}

func parseFlags() demoConfig {
	cfg := demoConfig{}

	flag.BoolVar(&cfg.full, "full", false, "show the plain stacktrace output")
	flag.BoolVar(&cfg.error, "error", false, "show the plain Error() output")
	flag.BoolVar(&cfg.stacktrace, "stacktrace", false, "show the stacktrace output")
	flag.BoolVar(&cfg.colored, "color", false, "show the colored stacktrace output")
	flag.Parse()

	if cfg.full {
		cfg.stacktrace = true
	}

	if !cfg.error && !cfg.stacktrace && !cfg.colored {
		cfg.error = true
		cfg.stacktrace = true
		cfg.colored = true
	}

	return cfg
}

func demoCases() []demoCase {
	packageOnly := errors.Wrapf(
		errors.Wrapf(
			errors.Wrapf(errors.New("disk offline"), "persist draft"),
			"save user",
		),
		"http handler",
	)

	packageJoinWithForeignBranches := errors.Wrapf(
		errors.Join(
			errors.Wrapf(errors.New("cache unavailable"), "refresh cache"),
			goerrors.Join(
				goerrors.New("db timeout"),
				goerrors.New("queue saturated"),
			),
			fmt.Errorf(
				"replication failed: %w | %w",
				errors.Wrapf(errors.New("primary stale"), "replica sync"),
				goerrors.New("secondary missing"),
			),
		),
		"request fanout failed",
	)

	stdlibJoinWrappedByPackage := errors.Wrapf(
		goerrors.Join(
			errors.Wrapf(errors.New("disk offline"), "persist draft"),
			fmt.Errorf(
				"publish failed: %w | %w",
				goerrors.New("cache unavailable"),
				errors.Wrapf(errors.New("db timeout"), "flush queue"),
			),
			createErrorChain(5),
		),
		"worker run",
	)

	nestedMixedTree := errors.Join(
		errors.Wrapf(createErrorChain(4), "api handler"),
		fmt.Errorf(
			"shutdown failed: %w + %w",
			errors.Join(
				errors.New("drain connections"),
				goerrors.New("close listener"),
			),
			goerrors.Join(
				goerrors.New("stop metrics"),
				errors.Wrapf(errors.New("flush spans"), "telemetry cleanup"),
			),
		),
		errors.Wrapf(
			fmt.Errorf(
				"batch failed: %w and %w",
				goerrors.New("notify webhook"),
				errors.Wrapf(errors.New("write audit log"), "finalize request"),
			),
			"request cleanup",
		),
	)

	return []demoCase{
		{
			title:       "Package-only chain",
			description: "Pure package wrapping shows the default Error() output beside the resolved stacktrace view.",
			build: strings.TrimSpace(`packageOnly := errors.Wrapf(
	errors.Wrapf(
		errors.Wrapf(errors.New("disk offline"), "persist draft"),
		"save user",
	),
	"http handler",
)`),
			err: packageOnly,
		},
		{
			title:       "Package join with foreign branches",
			description: "A package-owned join combines package chains, stdlib errors.Join, and multi-%w fmt.Errorf branches under one package wrapper.",
			build: strings.TrimSpace(`packageJoinWithForeignBranches := errors.Wrapf(
	errors.Join(
		errors.Wrapf(errors.New("cache unavailable"), "refresh cache"),
		goerrors.Join(
			goerrors.New("db timeout"),
			goerrors.New("queue saturated"),
		),
		fmt.Errorf(
			"replication failed: %w | %w",
			errors.Wrapf(errors.New("primary stale"), "replica sync"),
			goerrors.New("secondary missing"),
		),
	),
	"request fanout failed",
)`),
			err: packageJoinWithForeignBranches,
		},
		{
			title:       "Stdlib join wrapped by package",
			description: "A stdlib errors.Join tree is wrapped by the package so the stacktrace shows the package node plus opaque foreign branches.",
			build: strings.TrimSpace(`stdlibJoinWrappedByPackage := errors.Wrapf(
	goerrors.Join(
		errors.Wrapf(errors.New("disk offline"), "persist draft"),
		fmt.Errorf(
			"publish failed: %w | %w",
			goerrors.New("cache unavailable"),
			errors.Wrapf(errors.New("db timeout"), "flush queue"),
		),
		createErrorChain(5),
	),
	"worker run",
)`),
			err: stdlibJoinWrappedByPackage,
		},
		{
			title:       "Nested mixed tree",
			description: "Nested package joins, stdlib joins, and multi-%w fmt.Errorf calls produce a larger tree to compare Error() with the expanded stacktrace.",
			build: strings.TrimSpace(`nestedMixedTree := errors.Join(
	errors.Wrapf(createErrorChain(4), "api handler"),
	fmt.Errorf(
		"shutdown failed: %w + %w",
		errors.Join(
			errors.New("drain connections"),
			goerrors.New("close listener"),
		),
		goerrors.Join(
			goerrors.New("stop metrics"),
			errors.Wrapf(errors.New("flush spans"), "telemetry cleanup"),
		),
	),
	errors.Wrapf(
		fmt.Errorf(
			"batch failed: %w and %w",
			goerrors.New("notify webhook"),
			errors.Wrapf(errors.New("write audit log"), "finalize request"),
		),
		"request cleanup",
	),
)`),
			err: nestedMixedTree,
		},
	}
}

func printOutputs(cfg demoConfig, demo demoCase) {
	type output struct {
		enabled bool
		title   string
		body    func() string
	}

	printOutputBlock("Construction", demo.build)

	outputs := []output{
		{enabled: cfg.error, title: "Error()", body: demo.err.Error},
		{enabled: cfg.stacktrace, title: "Stacktrace", body: func() string {
			return errors.Stacktrace(demo.err, errors.StacktraceWithColor(false), errors.StacktraceWithTrimFilePath(true))
		}},
		{enabled: cfg.colored, title: "Colored Stacktrace", body: func() string {
			return errors.Stacktrace(demo.err, errors.StacktraceWithColor(true), errors.StacktraceWithTrimFilePath(true))
		}},
	}

	for _, out := range outputs {
		if !out.enabled {
			continue
		}

		printOutputBlock(out.title, out.body())
	}
}

func printSectionHeader(title string) {
	line := strings.Repeat("=", max(72, len(title)+8))
	fmt.Println(headerColor.Sprintf("== %s ==", title))
	fmt.Println(lineColor.Sprint(line))
}

func printOutputBlock(title, body string) {
	fmt.Println(labelColor.Sprintf("-- %s --", title))
	fmt.Println(body)
	fmt.Println()
}
