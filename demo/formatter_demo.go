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

type demoOutput struct {
	title string
	body  func() string
}

type demoCase struct {
	title       string
	description string
	build       string
	err         error
	extra       []demoOutput
}

type demoTargetError struct{}

func (demoTargetError) Error() string { return "demo target" }

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
	constructorsAndWrap := errors.Join(
		errors.Wrapf(errors.Wrap(errors.New("disk offline")), "bootstrap service"),
		errors.Wrap(errors.Newf("config %q missing", "app.yaml")),
	)

	resultAndCatch := buildResultAndCatchError()
	mustAndRecover := buildMustAndRecoverError()

	proxySentinel := goerrors.New("proxy sentinel")
	wrappedSentinel := errors.Wrapf(proxySentinel, "proxy root")
	joinedForChecks := errors.Join(errors.New("left branch"), errors.Newf("right branch %d", 2))
	foreignForAs := fmt.Errorf("foreign wrap: %w", demoTargetError{})
	proxyAndIntrospection := errors.Wrapf(joinedForChecks, "introspection root")

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
			title:       "Constructors and wrappers",
			description: "Shows package-owned leaf creation with New and Newf, then adds plain and formatted wrapper context with Wrap and Wrapf.",
			build: strings.TrimSpace(`constructorsAndWrap := errors.Join(
	errors.Wrapf(errors.Wrap(errors.New("disk offline")), "bootstrap service"),
	errors.Wrap(errors.Newf("config %q missing", "app.yaml")),
)`),
			err: constructorsAndWrap,
		},
		{
			title:       "Result wrappers and catch helpers",
			description: "Covers WrapResult, WrapResultf, Catch, Catchf, CatchResult, CatchResultf, and IgnoreResult in one joined error tree.",
			build: strings.TrimSpace(`resultAndCatch := errors.Join(
	wrapResultError(),
	wrapResultfError(),
	errors.Catch(goerrors.New("close rows")),
	errors.Catchf(goerrors.New("permission denied"), "delete user %d", 7),
	errors.CatchResult("rows", goerrors.New("query failed"))(func(string) error { return nil }),
	errors.CatchResultf("row", nil)(func(string) error { return goerrors.New("scan failed") }, "scan row %d", 7),
	errors.CatchResultf("tx", goerrors.New("begin failed"))(errors.IgnoreResult[string](), "begin tx"),
)`),
			err: resultAndCatch,
		},
		{
			title:       "Must, recover and assertions",
			description: "Shows Must, Mustf, MustResult, MustResultf, MustResult2, MustResult2f, Recover, RecoverFn, Assert, Assertf, and Ignore.",
			build: strings.TrimSpace(`mustAndRecover := errors.Join(
	recoverMust(func() { errors.Must(goerrors.New("connect failed")) }),
	recoverMust(func() { errors.Mustf(goerrors.New("open store"))("bootstrap %s", "cache") }),
	recoverMust(func() { _ = errors.MustResult("cfg", goerrors.New("load config")) }),
	recoverMust(func() { _ = errors.MustResultf("user", goerrors.New("lookup user"))("load user %d", 7) }),
	recoverMust(func() { _, _ = errors.MustResult2("k", "v", goerrors.New("read pair")) }),
	recoverMust(func() { _, _ = errors.MustResult2f("k", "v", goerrors.New("read pair"))("load pair %s", "session") }),
	recoverMust(func() { errors.Assert(false, "state mismatch") }),
	recoverMust(func() { errors.Assertf(false, "state %s", "invalid") }),
	recoverFnError(),
)`),
			err: mustAndRecover,
			extra: []demoOutput{
				{
					title: "Value Helpers",
					body: func() string {
						return fmt.Sprintf("errors.Ignore(42, goerrors.New(\"ignored\")) => %d", errors.Ignore(42, goerrors.New("ignored")))
					},
				},
			},
		},
		{
			title:       "Proxy and introspection helpers",
			description: "Covers As, Is, Unwrap, Expand, and the package type/introspection helpers against package-owned and foreign errors.",
			build: strings.TrimSpace(`proxyAndIntrospection := errors.Wrapf(errors.Join(
	errors.New("left branch"),
	errors.Newf("right branch %d", 2),
), "introspection root")`),
			err: proxyAndIntrospection,
			extra: []demoOutput{
				{
					title: "Helper Checks",
					body: func() string {
						return formatProxySummary(proxySentinel, wrappedSentinel, joinedForChecks, foreignForAs)
					},
				},
			},
		},
		{
			title:       "Nested package and custom stacktraces",
			description: "Shows nested package frames, stdlib interop, joined branches, and the stacktrace formatting options around file paths, package.func names, branch labels, prefixes, colors, and indentation.",
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
			extra: []demoOutput{
				{
					title: "Verbose Stacktrace",
					body: func() string {
						return errors.Stacktrace(
							nestedMixedTree,
							errors.StacktraceWithColor(false),
							errors.StacktraceWithTrimFilePath(false),
							errors.StacktraceWithSuppressEmptyFrames(true),
						)
					},
				},
				{
					title: "Location Formats",
					body: func() string {
						return formatLocationFormats(nestedMixedTree)
					},
				},
				{
					title: "Custom Colored Stacktrace",
					body: func() string {
						return customColoredStacktrace(nestedMixedTree)
					},
				},
			},
		},
	}
}

func buildResultAndCatchError() error {
	_, wrapResultErr := errors.WrapResult("port", goerrors.New("parse failed"))
	_, wrapResultfErr := errors.WrapResultf("user", goerrors.New("lookup failed"))("load user %d", 7)
	catchErr := errors.Catch(goerrors.New("close rows"))
	catchfErr := errors.Catchf(goerrors.New("permission denied"), "delete user %d", 7)
	catchResultErr := errors.CatchResult("rows", goerrors.New("query failed"))(func(string) error {
		return nil
	})
	catchResultfErr := errors.CatchResultf("row", nil)(func(string) error {
		return goerrors.New("scan failed")
	}, "scan row %d", 7)
	ignoreResultErr := errors.CatchResultf("tx", goerrors.New("begin failed"))(errors.IgnoreResult[string](), "begin tx")

	return errors.Join(
		wrapResultErr,
		wrapResultfErr,
		catchErr,
		catchfErr,
		catchResultErr,
		catchResultfErr,
		ignoreResultErr,
	)
}

func buildMustAndRecoverError() error {
	return errors.Join(
		recoverMust(func() { errors.Must(goerrors.New("connect failed")) }),
		recoverMust(func() { errors.Mustf(goerrors.New("open store"))("bootstrap %s", "cache") }),
		recoverMust(func() { _ = errors.MustResult("cfg", goerrors.New("load config")) }),
		recoverMust(func() { _ = errors.MustResultf("user", goerrors.New("lookup user"))("load user %d", 7) }),
		recoverMust(func() { _, _ = errors.MustResult2("k", "v", goerrors.New("read pair")) }),
		recoverMust(func() { _, _ = errors.MustResult2f("k", "v", goerrors.New("read pair"))("load pair %s", "session") }),
		recoverMust(func() { errors.Assert(false, "state mismatch") }),
		recoverMust(func() { errors.Assertf(false, "state %s", "invalid") }),
		recoverFnError(),
	)
}

func recoverMust(fn func()) (err error) {
	defer errors.Recover(&err)
	fn()
	return nil
}

func recoverFnError() error {
	var err error

	func() {
		defer errors.RecoverFn(func(recovered error) {
			err = recovered
		})
		panic("recover fn panic")
	}()

	return err
}

func formatProxySummary(sentinel, wrappedSentinel, joinedForChecks, foreignForAs error) string {
	_, unwrapSingle := errors.IsUnwrapSingle(wrappedSentinel)
	_, unwrapMulti := errors.IsUnwrapMulti(joinedForChecks)
	single, _ := errors.TryUnwrapSingle(wrappedSentinel)
	multi, _ := errors.TryUnwrapMulti(joinedForChecks)
	_, isChain := errors.IsErrorChain(wrappedSentinel)
	_, isTree := errors.IsErrorTree(joinedForChecks)
	var target demoTargetError
	expanded := errors.Expand(joinedForChecks)

	return strings.Join([]string{
		fmt.Sprintf("errors.Is(wrappedSentinel, sentinel) = %v", errors.Is(wrappedSentinel, sentinel)),
		fmt.Sprintf("errors.Unwrap(wrappedSentinel) = %q", errors.Unwrap(wrappedSentinel).Error()),
		fmt.Sprintf("errors.As(foreignForAs, &target) = %v", errors.As(foreignForAs, &target)),
		fmt.Sprintf("errors.IsErrorChain(wrappedSentinel) = %v", isChain),
		fmt.Sprintf("errors.IsErrorTree(joinedForChecks) = %v", isTree),
		fmt.Sprintf("errors.IsUnwrapSingle(wrappedSentinel) = %v", unwrapSingle),
		fmt.Sprintf("errors.TryUnwrapSingle(wrappedSentinel) = %q", single.Error()),
		fmt.Sprintf("errors.IsUnwrapMulti(joinedForChecks) = %v", unwrapMulti),
		fmt.Sprintf("errors.TryUnwrapMulti(joinedForChecks) len = %d", len(multi)),
		fmt.Sprintf("errors.Expand(joinedForChecks) = message=%q multi=%v children=%d", expanded.Resolved.Message, expanded.Resolved.Multi, len(expanded.Children)),
	}, "\n")
}

func formatLocationFormats(err error) string {
	packageAndFunc := takeLines(errors.Stacktrace(
		err,
		errors.StacktraceWithColor(false),
		errors.StacktraceWithTrimFilePath(false),
		errors.StacktraceWithSuppressEmptyFrames(true),
		errors.StacktraceWithFunctionFormat(errors.StacktraceFunctionPackageAndFunc),
	), 5)

	funcOnly := takeLines(errors.Stacktrace(
		err,
		errors.StacktraceWithColor(false),
		errors.StacktraceWithTrimFilePath(true),
		errors.StacktraceWithSuppressEmptyFrames(true),
		errors.StacktraceWithFunctionFormat(errors.StacktraceFunctionFuncOnly),
	), 5)

	fileOnly := takeLines(errors.Stacktrace(
		err,
		errors.StacktraceWithColor(false),
		errors.StacktraceWithTrimFilePath(true),
		errors.StacktraceWithSuppressEmptyFrames(true),
		errors.StacktraceWithFunctionFormat(errors.StacktraceFunctionNone),
	), 5)

	return strings.Join([]string{
		"package+func:\n" + packageAndFunc,
		"func-only:\n" + funcOnly,
		"file-only:\n" + fileOnly,
	}, "\n\n")
}

func customColoredStacktrace(err error) string {
	return errors.Stacktrace(
		err,
		errors.StacktraceWithColor(true),
		errors.StacktraceWithTrimFilePath(false),
		errors.StacktraceWithSuppressEmptyFrames(true),
		errors.StacktraceWithPreIndent(2),
		errors.StacktraceWithIndent(4),
		errors.StacktraceWithFunctionFormat(errors.StacktraceFunctionPackageAndFunc),
		errors.StacktraceWithTreePrefixFormatter(func(colorEnabled bool) string {
			if !colorEnabled {
				return ">>"
			}

			return lineColor.Sprint(">>")
		}),
		errors.StacktraceWithBranchLabel(func(index int) string {
			return fmt.Sprintf("branch %d: ", index)
		}),
		errors.StacktraceWithColors(errors.StacktraceColors{
			Source:  color.New(color.FgHiBlue),
			Func:    color.New(color.FgHiMagenta),
			Message: color.New(color.FgHiWhite),
		}),
	)
}

func takeLines(body string, maxLines int) string {
	lines := strings.Split(body, "\n")
	if len(lines) <= maxLines {
		return body
	}

	return strings.Join(lines[:maxLines], "\n")
}

func printOutputs(cfg demoConfig, demo demoCase) {
	printOutputBlock("Construction", demo.build)

	outputs := []demoOutput{}
	if cfg.error {
		outputs = append(outputs, demoOutput{title: "Error()", body: demo.err.Error})
	}

	if cfg.stacktrace {
		outputs = append(outputs, demoOutput{title: "Stacktrace", body: func() string {
			return errors.Stacktrace(demo.err, errors.StacktraceWithColor(false), errors.StacktraceWithTrimFilePath(true))
		}})
	}

	if cfg.colored {
		outputs = append(outputs, demoOutput{title: "Colored Stacktrace", body: func() string {
			return errors.Stacktrace(demo.err, errors.StacktraceWithColor(true), errors.StacktraceWithTrimFilePath(true))
		}})
	}

	outputs = append(outputs, demo.extra...)

	for _, out := range outputs {
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
