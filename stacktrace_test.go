package errors

import (
	stderrors "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestStacktraceFormatsChainLeafFirstWithoutNestedFrameIndent(t *testing.T) {
	err := WrapSkipf(0, NewSkipf(0, "root line 1\nroot line 2"), "outer line 1\nouter line 2")
	entry := Expand(err)
	leafLocation, _ := formatStackLocation(&entry.Children[0].Resolved, &StacktraceOptions{})
	outerLocation, _ := formatStackLocation(&entry.Resolved, &StacktraceOptions{})

	got := Stacktrace(err, StacktraceWithColor(false), StacktraceWithTreePrefix(""))
	lines := strings.Split(got, "\n")
	if len(lines) != 4 {
		t.Fatalf("line count = %d, want 4\n%s", len(lines), got)
	}

	leafPrefix := "at " + leafLocation + ": "
	outerPrefix := "  at " + outerLocation + ": "
	if lines[0] != leafPrefix+"root line 1" {
		t.Fatalf("line 1 = %q, want %q", lines[0], leafPrefix+"root line 1")
	}

	if lines[1] != strings.Repeat(" ", len(leafPrefix))+"root line 2" {
		t.Fatalf("line 2 = %q", lines[1])
	}

	if lines[2] != outerPrefix+"outer line 1" {
		t.Fatalf("line 3 = %q, want %q", lines[2], outerPrefix+"outer line 1")
	}

	if lines[3] != strings.Repeat(" ", len(outerPrefix))+"outer line 2" {
		t.Fatalf("line 4 = %q", lines[3])
	}
}

func TestStacktraceFormatsForeignRootCauseBeforePackageFrames(t *testing.T) {
	err := WrapSkipf(0, stderrors.New("boom"), "context")
	entry := Expand(err)
	outerLocation, _ := formatStackLocation(&entry.Resolved, &StacktraceOptions{})

	got := Stacktrace(err, StacktraceWithColor(false), StacktraceWithTreePrefix(""))
	want := strings.Join([]string{
		"boom",
		"  at " + outerLocation + ": context",
	}, "\n")

	if got != want {
		t.Fatalf("Stacktrace() =\n%s\nwant\n%s", got, want)
	}
}

func TestStacktraceCanSuppressMessageLessFrames(t *testing.T) {
	err := WrapSkipf(0, NewSkipf(0, "boom"), "context")
	entry := Expand(err)
	outerLocation, _ := formatStackLocation(&entry.Resolved, &StacktraceOptions{})
	leafLocation, _ := formatStackLocation(&entry.Children[0].Resolved, &StacktraceOptions{})

	got := Stacktrace(
		err,
		StacktraceWithColor(false),
		StacktraceWithTreePrefix(""),
		StacktraceWithSuppressEmptyFrames(true),
	)
	want := strings.Join([]string{
		"at " + leafLocation + ": boom",
		"  at " + outerLocation + ": context",
	}, "\n")

	if got != want {
		t.Fatalf("Stacktrace() =\n%s\nwant\n%s", got, want)
	}

	if strings.Count(got, leafLocation) != 1 {
		t.Fatalf("Stacktrace() should contain exactly one leaf frame %q:\n%s", leafLocation, got)
	}
}

func TestStacktraceCollapsesDuplicateEmptyFramesAtSameLocation(t *testing.T) {
	err := Wrapf(Wrap(New("boom")), "context")
	entry := Expand(err)
	leafLocation := mustRawLocation(&entry.Children[0].Children[0].Resolved)
	outerLocation := mustRawLocation(&entry.Resolved)

	got := Stacktrace(err, StacktraceWithColor(false), StacktraceWithTreePrefix(""))
	want := strings.Join([]string{
		"at " + leafLocation + ": boom",
		"  at " + outerLocation + ": context",
	}, "\n")

	if got != want {
		t.Fatalf("Stacktrace() =\n%s\nwant\n%s", got, want)
	}
}

func TestStacktraceSuppressEmptyFramesDoesNotAddGhostDepth(t *testing.T) {
	err := WrapSkip(0, Join(NewSkipf(0, "left"), NewSkipf(0, "right")))
	entry := Expand(err)

	got := Stacktrace(
		err,
		StacktraceWithColor(false),
		StacktraceWithSuppressEmptyFrames(true),
	)
	want := strings.Join([]string{
		"at " + mustRawLocation(&entry.Children[0].Resolved) + ": joined errors",
		"| [1] at " + mustRawLocation(&entry.Children[0].Children[0].Resolved) + ": left",
		"| [2] at " + mustRawLocation(&entry.Children[0].Children[1].Resolved) + ": right",
	}, "\n")

	if got != want {
		t.Fatalf("Stacktrace() =\n%s\nwant\n%s", got, want)
	}
}

func TestStacktraceFormatsJoinedErrorsAsSiblingBranches(t *testing.T) {
	err := Join(
		NewSkipf(0, "left"),
		fmt.Errorf("foreign: %w", stderrors.New("boom")),
		WrapSkipf(0, stderrors.New("right"), "context"),
	)
	entry := Expand(err)
	rootLocation, _ := formatStackLocation(&entry.Resolved, &StacktraceOptions{})
	leftLocation, _ := formatStackLocation(&entry.Children[0].Resolved, &StacktraceOptions{})
	contextLocation, _ := formatStackLocation(&entry.Children[2].Resolved, &StacktraceOptions{})

	got := Stacktrace(err, StacktraceWithIndent(4), StacktraceWithColor(false), StacktraceWithTreePrefix(""))
	want := strings.Join([]string{
		"at " + rootLocation + ": joined errors",
		"    [1] at " + leftLocation + ": left",
		"    [2] foreign: boom",
		"    [3] right",
		"        at " + contextLocation + ": context",
	}, "\n")

	if got != want {
		t.Fatalf("Stacktrace() =\n%s\nwant\n%s", got, want)
	}
}

func TestStacktraceSupportsTreePrefixAndCustomBranchLabels(t *testing.T) {
	err := Join(NewSkipf(0, "left"), NewSkipf(0, "right"))
	entry := Expand(err)
	leftLocation, _ := formatStackLocation(&entry.Children[0].Resolved, &StacktraceOptions{})
	rightLocation, _ := formatStackLocation(&entry.Children[1].Resolved, &StacktraceOptions{})

	got := Stacktrace(
		err,
		StacktraceWithColor(false),
		StacktraceWithBranchLabel(func(index int) string { return "(" + fmt.Sprint(index) + ") " }),
	)

	wantLines := []string{
		"at " + mustRawLocation(&entry.Resolved) + ": joined errors",
		"| (1) at " + leftLocation + ": left",
		"| (2) at " + rightLocation + ": right",
	}

	if got != strings.Join(wantLines, "\n") {
		t.Fatalf("Stacktrace() =\n%s\nwant\n%s", got, strings.Join(wantLines, "\n"))
	}
}

func TestStacktracePassesColorFlagToTreePrefixFormatter(t *testing.T) {
	err := Join(NewSkipf(0, "left"), NewSkipf(0, "right"))
	entry := Expand(err)
	leftLocation, _ := formatStackLocation(&entry.Children[0].Resolved, &StacktraceOptions{})
	rightLocation, _ := formatStackLocation(&entry.Children[1].Resolved, &StacktraceOptions{})

	got := Stacktrace(
		err,
		StacktraceWithColor(false),
		StacktraceWithTreePrefixFormatter(func(colorEnabled bool) string {
			if colorEnabled {
				return "#"
			}

			return ">"
		}),
	)

	wantLines := []string{
		"at " + mustRawLocation(&entry.Resolved) + ": joined errors",
		"> [1] at " + leftLocation + ": left",
		"> [2] at " + rightLocation + ": right",
	}

	if got != strings.Join(wantLines, "\n") {
		t.Fatalf("Stacktrace() =\n%s\nwant\n%s", got, strings.Join(wantLines, "\n"))
	}
}

func TestStacktraceKeepsTreePrefixOnForeignMultilineContinuation(t *testing.T) {
	err := Join(NewSkipf(0, "left"), stderrors.Join(stderrors.New("db timeout"), stderrors.New("queue saturated")))

	got := Stacktrace(err, StacktraceWithColor(false))
	if !strings.Contains(got, "\n|     queue saturated") {
		t.Fatalf("Stacktrace() missing tree-prefixed continuation:\n%s", got)
	}
}

func TestStacktraceDoesNotColorContinuationTreePrefixAsMessage(t *testing.T) {
	oldNoColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = oldNoColor }()

	cyan := color.New(color.FgCyan)
	red := color.New(color.FgRed)
	err := Join(NewSkipf(0, "left"), stderrors.Join(stderrors.New("db timeout"), stderrors.New("queue saturated")))

	got := Stacktrace(
		err,
		StacktraceWithColor(true),
		StacktraceWithTreePrefixFormatter(func(colorEnabled bool) string {
			if !colorEnabled {
				return "|"
			}

			return cyan.Sprint("|")
		}),
		StacktraceWithColors(StacktraceColors{Message: red}),
	)

	if !strings.Contains(got, "\n"+cyan.Sprint("|")+"     "+red.Sprint("queue saturated")) {
		t.Fatalf("Stacktrace() continuation did not preserve tree-prefix color:\n%s", got)
	}
}

func TestStacktraceResolvesDefaultColorsForPartialOverrides(t *testing.T) {
	colors := resolvedStacktraceColors(StacktraceColors{Message: color.New(color.FgGreen)})

	if colors.Source == nil {
		t.Fatal("Source color was nil")
	}

	if colors.Func == nil {
		t.Fatal("Func color was nil")
	}

	if colors.Message == nil {
		t.Fatal("Message color was nil")
	}
}

func TestStacktraceCanTrimFilePathsToWorkspaceRelative(t *testing.T) {
	err := NewSkipf(0, "boom")
	entry := Expand(err)
	fullRawLocation, _ := formatStackLocation(&entry.Resolved, &StacktraceOptions{})
	trimmedRawLocation, _ := formatStackLocation(&entry.Resolved, &StacktraceOptions{TrimFilePath: true})

	if fullRawLocation == trimmedRawLocation {
		t.Fatalf("expected trimmed location to differ from full location: %q", trimmedRawLocation)
	}

	if strings.Contains(trimmedRawLocation, workspaceRoot) {
		t.Fatalf("trimmed location still contains workspace root: %q", trimmedRawLocation)
	}

	if strings.HasPrefix(trimmedRawLocation, "/") {
		t.Fatalf("trimmed location is still absolute: %q", trimmedRawLocation)
	}

	if !strings.Contains(trimmedRawLocation, ".go:") {
		t.Fatalf("trimmed location missing go file component: %q", trimmedRawLocation)
	}
}

func TestStacktraceFunctionFormatOptions(t *testing.T) {
	err := NewSkipf(0, "boom")
	entry := Expand(err)

	withPackage, _ := formatStackLocation(&entry.Resolved, &StacktraceOptions{FunctionFormat: StacktraceFunctionPackageAndFunc})
	funcOnly, _ := formatStackLocation(&entry.Resolved, &StacktraceOptions{FunctionFormat: StacktraceFunctionFuncOnly})
	withoutFunc, _ := formatStackLocation(&entry.Resolved, &StacktraceOptions{FunctionFormat: StacktraceFunctionNone})

	if !strings.Contains(withPackage, ":errors.") {
		t.Fatalf("package+func location = %q, want package-qualified function", withPackage)
	}

	if !strings.Contains(funcOnly, ":NewSkipf:") {
		t.Fatalf("func-only location = %q, want bare function name", funcOnly)
	}

	if strings.Contains(funcOnly, ":errors.") {
		t.Fatalf("func-only location = %q, want no package qualifier", funcOnly)
	}

	if strings.Contains(withoutFunc, "TestStacktraceFunctionFormatOptions") || strings.Contains(withoutFunc, ":errors.") {
		t.Fatalf("none location = %q, want file:line only", withoutFunc)
	}

	if strings.Count(withoutFunc, ":") != 1 {
		t.Fatalf("none location = %q, want exactly file:line", withoutFunc)
	}
}

func TestStacktraceSupportsPreIndent(t *testing.T) {
	err := Join(NewSkipf(0, "left"), NewSkipf(0, "right"))
	entry := Expand(err)

	got := Stacktrace(err, StacktraceWithColor(false), StacktraceWithPreIndent(3))
	want := strings.Join([]string{
		"   at " + mustRawLocation(&entry.Resolved) + ": joined errors",
		"   | [1] at " + mustRawLocation(&entry.Children[0].Resolved) + ": left",
		"   | [2] at " + mustRawLocation(&entry.Children[1].Resolved) + ": right",
	}, "\n")

	if got != want {
		t.Fatalf("Stacktrace() =\n%s\nwant\n%s", got, want)
	}
}

func mustRawLocation(entry *ResolvedEntry) string {
	rawLocation, _ := formatStackLocation(entry, &StacktraceOptions{})
	return rawLocation
}
