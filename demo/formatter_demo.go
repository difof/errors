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
	full    bool
	chain   bool
	root    bool
	colored bool
	json    bool
	yaml    bool
}

type demoCase struct {
	title       string
	description string
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
		printOutputs(cfg, demo.err)
	}
}

func parseFlags() demoConfig {
	cfg := demoConfig{}

	flag.BoolVar(&cfg.full, "full", false, "show the full stack-aware text output")
	flag.BoolVar(&cfg.chain, "chain", false, "show the wrapped-message-only output")
	flag.BoolVar(&cfg.root, "root", false, "show only the innermost/root message")
	flag.BoolVar(&cfg.colored, "color", false, "show the colored formatter output")
	flag.BoolVar(&cfg.json, "json", false, "show the JSON formatter output")
	flag.BoolVar(&cfg.yaml, "yaml", false, "show the YAML formatter output")
	flag.Parse()

	if !cfg.full && !cfg.chain && !cfg.root && !cfg.colored && !cfg.json && !cfg.yaml {
		cfg.full = true
		cfg.chain = true
		cfg.root = true
		cfg.colored = true
		cfg.json = true
		cfg.yaml = true
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

	mixedWraps := errors.Wrapf(
		errors.Wrapf(
			fmt.Errorf("std wrapped error %w", createErrorChain(5)),
			"wrapped error 2",
		),
		"wrapped error 1",
	)

	joinedLeaf := errors.Wrapf(
		fmt.Errorf(
			"publish failed: %w",
			goerrors.Join(
				goerrors.New("cache unavailable"),
				goerrors.New("db timeout"),
			),
		),
		"worker run",
	)

	return []demoCase{
		{
			title:       "Package-only chain",
			description: "Pure package wrapping shows the difference between detailed stack output and plain wrapped messages.",
			err:         packageOnly,
		},
		{
			title:       "Mixed stdlib + package wrapping",
			description: "Mixes fmt.Errorf(%w) with this package to show the best-effort stackless rendering.",
			err:         mixedWraps,
		},
		{
			title:       "Joined leaf error",
			description: "Shows how errors.Join is preserved as Go's default joined text while outer wrappers stay stackless.",
			err:         joinedLeaf,
		},
	}
}

func printOutputs(cfg demoConfig, err error) {
	chain, ok := err.(*errors.ErrorChain)
	if !ok {
		fmt.Println("demo setup error: expected *errors.ErrorChain")
		return
	}

	type output struct {
		enabled bool
		title   string
		body    func() string
	}

	outputs := []output{
		{enabled: cfg.full, title: "Full Detail", body: err.Error},
		{enabled: cfg.chain, title: "Wrapped Messages Only", body: func() string { return errors.ChainMessages(err) }},
		{enabled: cfg.root, title: "Root Message Only", body: func() string { return errors.RootMessage(err) }},
		{enabled: cfg.colored, title: "Colored Format", body: chain.Colored},
		{enabled: cfg.json, title: "JSON Format", body: chain.JSON},
		{enabled: cfg.yaml, title: "YAML Format", body: chain.YAML},
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

