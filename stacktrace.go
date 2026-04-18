package errors

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

var (
	stacktraceSourceColor  = color.New(color.FgBlue)
	stacktraceFuncColor    = color.New(color.FgYellow)
	stacktraceMessageColor = color.New(color.FgRed)
	ansiRegexp             = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	workspaceRoot          = detectWorkspaceRoot()
)

// StacktraceBranchLabelFormatter formats the label shown for each child branch.
type StacktraceBranchLabelFormatter func(index int) string

// StacktraceTreePrefixFormatter returns the prefix used for one tree indent unit.
type StacktraceTreePrefixFormatter func(colorEnabled bool) string

// StacktraceFunctionFormat controls how function information is rendered in a stack location.
type StacktraceFunctionFormat int

const (
	// StacktraceFunctionPackageAndFunc renders package and function names.
	StacktraceFunctionPackageAndFunc StacktraceFunctionFormat = iota
	// StacktraceFunctionFuncOnly renders only the function name.
	StacktraceFunctionFuncOnly
	// StacktraceFunctionNone omits function information entirely.
	StacktraceFunctionNone
)

// StacktraceColors configures the colors used by Stacktrace when color is enabled.
type StacktraceColors struct {
	Source  *color.Color
	Func    *color.Color
	Message *color.Color
}

// StacktraceOptions holds the resolved configuration for Stacktrace rendering.
type StacktraceOptions struct {
	Indent              int
	PreIndent           int
	Color               bool
	SuppressEmptyFrames bool
	TrimFilePath        bool
	FunctionFormat      StacktraceFunctionFormat
	TreePrefixFormatter StacktraceTreePrefixFormatter
	BranchLabel         StacktraceBranchLabelFormatter
	Colors              StacktraceColors
}

// StacktraceOption mutates StacktraceOptions before rendering.
type StacktraceOption func(opt *StacktraceOptions)

// StacktraceWithIndent sets the number of columns added for each tree depth.
func StacktraceWithIndent(spaces int) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.Indent = spaces
	}
}

// StacktraceWithPreIndent sets a fixed left margin for all rendered lines.
func StacktraceWithPreIndent(spaces int) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.PreIndent = spaces
	}
}

// StacktraceWithColor enables or disables colorized output.
func StacktraceWithColor(color bool) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.Color = color
	}
}

// StacktraceWithSuppressEmptyFrames hides frame lines that have no local message.
func StacktraceWithSuppressEmptyFrames(suppress bool) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.SuppressEmptyFrames = suppress
	}
}

// StacktraceWithTrimFilePath renders file paths relative to the detected module root.
func StacktraceWithTrimFilePath(trim bool) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.TrimFilePath = trim
	}
}

// StacktraceWithFunctionFormat selects how function information is rendered.
func StacktraceWithFunctionFormat(format StacktraceFunctionFormat) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.FunctionFormat = format
	}
}

// StacktraceWithTreePrefix uses prefix for each tree indent unit.
func StacktraceWithTreePrefix(prefix string) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.TreePrefixFormatter = func(bool) string { return prefix }
	}
}

// StacktraceWithTreePrefixFormatter customizes the prefix for each tree indent unit.
func StacktraceWithTreePrefixFormatter(formatter StacktraceTreePrefixFormatter) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.TreePrefixFormatter = formatter
	}
}

// StacktraceWithBranchLabel customizes the label shown for child branches.
func StacktraceWithBranchLabel(formatter StacktraceBranchLabelFormatter) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.BranchLabel = formatter
	}
}

// StacktraceWithColors overrides one or more default output colors.
func StacktraceWithColors(colors StacktraceColors) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.Colors = colors
	}
}

// Stacktrace renders err as a readable tree/chain view for debugging.
// Package-owned errors are expanded structurally; foreign errors are rendered as opaque leaves.
func Stacktrace(err error, options ...StacktraceOption) string {
	config := &StacktraceOptions{
		Indent:              2,
		PreIndent:           0,
		Color:               true,
		SuppressEmptyFrames: false,
		TrimFilePath:        false,
		FunctionFormat:      StacktraceFunctionPackageAndFunc,
		TreePrefixFormatter: defaultStacktraceTreePrefix,
		BranchLabel:         defaultStacktraceBranchLabel,
		Colors:              defaultStacktraceColors(),
	}

	for _, opt := range options {
		opt(config)
	}

	if config.Indent < 0 {
		config.Indent = 0
	}

	if config.PreIndent < 0 {
		config.PreIndent = 0
	}

	if config.BranchLabel == nil {
		config.BranchLabel = defaultStacktraceBranchLabel
	}

	if config.TreePrefixFormatter == nil {
		config.TreePrefixFormatter = defaultStacktraceTreePrefix
	}

	config.Colors = resolvedStacktraceColors(config.Colors)

	entry := Expand(err)
	if entry == nil {
		return ""
	}

	renderer := stacktraceRenderer{options: config}
	renderer.renderEntry(entry, 0, "")

	return renderer.sb.String()
}

// stacktraceRenderer writes a stacktrace incrementally while preserving branch state.
type stacktraceRenderer struct {
	sb           strings.Builder
	options      *StacktraceOptions
	needsNewline bool
}

// renderEntry renders one entry subtree at the requested depth.
func (r *stacktraceRenderer) renderEntry(entry *ErrorEntry, depth int, firstLineLabel string) {
	if entry == nil {
		return
	}

	path, tail := collectPackageChain(entry)

	if tail == nil {
		r.renderPackageChain(path, depth, firstLineLabel)
		return
	}

	if tail.Resolved.Multi {
		r.renderMultiBranch(path, tail, depth, firstLineLabel)
		return
	}

	r.renderForeignTerminatedChain(path, tail, depth, firstLineLabel)
}

// renderPackageChain renders a package-owned single-child chain leaf-first.
func (r *stacktraceRenderer) renderPackageChain(path []*ErrorEntry, depth int, firstLineLabel string) {
	if len(path) == 0 {
		return
	}

	firstFrameDepth := depth
	childFrameDepth := depth
	firstFrameLabel := firstLineLabel
	firstRendered := true

	var lastFrame *ResolvedEntry
	lastMessage := ""

	for i := len(path) - 1; i >= 0; i-- {
		label := firstFrameLabel
		firstFrameLabel = ""

		message := path[i].Resolved.Message

		if shouldSuppressFrame(message, r.options) {
			firstFrameLabel = label
			continue
		}

		if shouldCollapseFrame(lastFrame, lastMessage, &path[i].Resolved, message) {
			continue
		}

		frameDepth := childFrameDepth
		if firstRendered {
			frameDepth = firstFrameDepth
			if message != "" {
				childFrameDepth = depth + 1
			}
			firstRendered = false
		}

		r.writeLine(formatFrameLine(frameDepth, label, &path[i].Resolved, message, r.options))
		lastFrame = &path[i].Resolved
		lastMessage = message
	}
}

// renderForeignTerminatedChain renders a package chain whose leaf is a foreign error.
func (r *stacktraceRenderer) renderForeignTerminatedChain(path []*ErrorEntry, tail *ErrorEntry, depth int, firstLineLabel string) {
	hasRootMessage := tail.Resolved.Message != ""
	if hasRootMessage {
		r.writeLine(formatMessageLine(depth, firstLineLabel, tail.Resolved.Message, r.options))
	}

	frameDepth := depth
	firstFrameLabel := firstLineLabel
	if hasRootMessage {
		frameDepth++
		firstFrameLabel = ""
	}

	var lastFrame *ResolvedEntry
	lastMessage := ""

	for i := len(path) - 1; i >= 0; i-- {
		label := firstFrameLabel
		firstFrameLabel = ""

		message := path[i].Resolved.Message
		if shouldSuppressFrame(message, r.options) {
			firstFrameLabel = label
			continue
		}

		if shouldCollapseFrame(lastFrame, lastMessage, &path[i].Resolved, message) {
			continue
		}

		r.writeLine(formatFrameLine(frameDepth, label, &path[i].Resolved, message, r.options))
		lastFrame = &path[i].Resolved
		lastMessage = message
	}
}

// renderMultiBranch renders a package-owned multi node and its child branches.
func (r *stacktraceRenderer) renderMultiBranch(path []*ErrorEntry, multi *ErrorEntry, depth int, firstLineLabel string) {
	currentDepth := depth
	label := firstLineLabel
	var lastFrame *ResolvedEntry
	lastMessage := ""

	for _, node := range path {
		labelForNode := label
		if shouldSuppressFrame(node.Resolved.Message, r.options) {
			label = labelForNode
			continue
		}

		if shouldCollapseFrame(lastFrame, lastMessage, &node.Resolved, node.Resolved.Message) {
			label = ""
			continue
		}

		r.writeLine(formatFrameLine(currentDepth, labelForNode, &node.Resolved, node.Resolved.Message, r.options))
		label = ""
		currentDepth++
		lastFrame = &node.Resolved
		lastMessage = node.Resolved.Message
	}

	multiMessage := multi.Resolved.Message
	if multiMessage == "" {
		multiMessage = "joined errors"
	}

	r.writeLine(formatFrameLine(currentDepth, label, &multi.Resolved, multiMessage, r.options))

	for i, child := range multi.Children {
		branchLabel := r.options.BranchLabel(i + 1)
		r.renderEntry(child, currentDepth+1, branchLabel)
	}
}

// writeLine appends one rendered line, inserting a newline when needed.
func (r *stacktraceRenderer) writeLine(line string) {
	if r.needsNewline {
		r.sb.WriteString("\n")
	}

	r.sb.WriteString(line)
	r.needsNewline = true
}

// collectPackageChain walks package-owned single-child nodes until a leaf or foreign boundary.
func collectPackageChain(entry *ErrorEntry) ([]*ErrorEntry, *ErrorEntry) {
	path := make([]*ErrorEntry, 0, 4)
	current := entry

	for current != nil && !current.Resolved.Foreign && !current.Resolved.Multi {
		path = append(path, current)

		switch len(current.Children) {
		case 0:
			return path, nil
		case 1:
			current = current.Children[0]
		default:
			return path, nil
		}
	}

	return path, current
}

// defaultStacktraceBranchLabel formats the default `[n] ` branch label.
func defaultStacktraceBranchLabel(index int) string {
	return "[" + strconv.Itoa(index) + "] "
}

// shouldSuppressFrame reports whether a frame line should be skipped.
func shouldSuppressFrame(message string, options *StacktraceOptions) bool {
	return options.SuppressEmptyFrames && message == ""
}

func shouldCollapseFrame(last *ResolvedEntry, lastMessage string, current *ResolvedEntry, currentMessage string) bool {
	if last == nil || currentMessage != "" {
		return false
	}

	_ = lastMessage
	return last.FilePath == current.FilePath && last.FuncPath == current.FuncPath && last.Line == current.Line
}

// defaultStacktraceTreePrefix returns the default tree prefix.
func defaultStacktraceTreePrefix(bool) string {
	return "|"
}

// defaultStacktraceColors returns the built-in stacktrace colors.
func defaultStacktraceColors() StacktraceColors {
	return StacktraceColors{
		Source:  stacktraceSourceColor,
		Func:    stacktraceFuncColor,
		Message: stacktraceMessageColor,
	}
}

// resolvedStacktraceColors fills in any missing colors with defaults.
func resolvedStacktraceColors(colors StacktraceColors) StacktraceColors {
	defaults := defaultStacktraceColors()

	if colors.Source == nil {
		colors.Source = defaults.Source
	}

	if colors.Func == nil {
		colors.Func = defaults.Func
	}

	if colors.Message == nil {
		colors.Message = defaults.Message
	}

	return colors
}

// formatMessageLine renders one message-only line.
func formatMessageLine(depth int, label, message string, options *StacktraceOptions) string {
	indent := indentPrefix(depth, options)
	renderedPrefix := indent + label
	continuationPrefix := indent + strings.Repeat(" ", visibleWidth(label))
	return formatIndentedMessage(renderedPrefix, continuationPrefix, message, options)
}

// formatFrameLine renders one stack frame line with an optional local message.
func formatFrameLine(depth int, label string, entry *ResolvedEntry, message string, options *StacktraceOptions) string {
	indent := indentPrefix(depth, options)
	renderedPrefix := indent + label
	rawLocation, renderedLocation := formatStackLocation(entry, options)

	if rawLocation == "" {
		continuationPrefix := indent + strings.Repeat(" ", visibleWidth(label))
		return formatIndentedMessage(renderedPrefix, continuationPrefix, message, options)
	}

	renderedPrefix += "at " + renderedLocation
	continuationPrefix := indent + strings.Repeat(" ", visibleWidth(label)+len("at ")+visibleWidth(rawLocation))
	if message == "" {
		return renderedPrefix
	}

	return formatIndentedMessage(renderedPrefix+": ", continuationPrefix+"  ", message, options)
}

// indentPrefix returns the full prefix for a given tree depth.
func indentPrefix(depth int, options *StacktraceOptions) string {
	prefix := strings.Repeat(" ", options.PreIndent)
	if depth <= 0 {
		return prefix
	}

	return prefix + strings.Repeat(indentUnit(options), depth)
}

// indentUnit returns one unit of tree indentation.
func indentUnit(options *StacktraceOptions) string {
	prefix := ""
	if options.TreePrefixFormatter != nil {
		prefix = options.TreePrefixFormatter(options.Color)
	}

	if prefix == "" {
		return strings.Repeat(" ", options.Indent)
	}

	padding := options.Indent - visibleWidth(prefix)
	if padding < 0 {
		padding = 0
	}

	return prefix + strings.Repeat(" ", padding)
}

// formatStackLocation builds the raw and rendered location strings for one entry.
func formatStackLocation(entry *ResolvedEntry, options *StacktraceOptions) (string, string) {
	rawParts := make([]string, 0, 3)
	renderedParts := make([]string, 0, 3)

	if entry.FilePath != "" {
		filePath := entry.FilePath
		if options.TrimFilePath {
			filePath = trimWorkspacePath(filePath)
		}

		rawParts = append(rawParts, filePath)
		renderedParts = append(renderedParts, colorizeSource(filePath, options))
	}

	if entry.FuncPath != "" {
		funcPath := formatFunctionPath(entry.FuncPath, options.FunctionFormat)
		if funcPath != "" {
			rawParts = append(rawParts, funcPath)
			renderedParts = append(renderedParts, colorizeFunc(funcPath, options))
		}
	}

	if entry.Line != 0 {
		line := strconv.Itoa(entry.Line)
		rawParts = append(rawParts, line)
		renderedParts = append(renderedParts, colorizeSource(line, options))
	}

	return strings.Join(rawParts, ":"), strings.Join(renderedParts, ":")
}

// formatIndentedMessage renders a possibly multiline message with aligned continuation lines.
func formatIndentedMessage(renderedPrefix, continuationPrefix, message string, options *StacktraceOptions) string {
	if message == "" {
		return renderedPrefix
	}

	parts := strings.Split(message, "\n")
	for i, part := range parts {
		parts[i] = colorizeMessage(part, options)
	}

	return renderedPrefix + strings.Join(parts, "\n"+continuationPrefix)
}

// colorizeSource applies the configured source color when enabled.
func colorizeSource(value string, options *StacktraceOptions) string {
	if options.Color && options.Colors.Source != nil {
		return options.Colors.Source.Sprint(value)
	}

	return value
}

// colorizeFunc applies the configured function color when enabled.
func colorizeFunc(value string, options *StacktraceOptions) string {
	if options.Color && options.Colors.Func != nil {
		return options.Colors.Func.Sprint(value)
	}

	return value
}

// colorizeMessage applies the configured message color when enabled.
func colorizeMessage(value string, options *StacktraceOptions) string {
	if options.Color && options.Colors.Message != nil {
		return options.Colors.Message.Sprint(value)
	}

	return value
}

// trimWorkspacePath converts an absolute file path into a module-relative path when possible.
func trimWorkspacePath(path string) string {
	if workspaceRoot == "" || path == "" {
		return path
	}

	rel, err := filepath.Rel(workspaceRoot, path)
	if err != nil {
		return path
	}

	if rel == "." || strings.HasPrefix(rel, "..") {
		return path
	}

	return filepath.ToSlash(rel)
}

// detectWorkspaceRoot walks upward from this source file until it finds go.mod.
func detectWorkspaceRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return filepath.Dir(file)
		}

		dir = parent
	}
}

// visibleWidth returns the display width after stripping ANSI escape codes.
func visibleWidth(value string) int {
	return len(ansiRegexp.ReplaceAllString(value, ""))
}

// formatFunctionPath reduces a runtime function path according to format.
func formatFunctionPath(funcPath string, format StacktraceFunctionFormat) string {
	if funcPath == "" || format == StacktraceFunctionNone {
		return ""
	}

	suffix := funcPath
	if slash := strings.LastIndex(funcPath, "/"); slash >= 0 && slash+1 < len(funcPath) {
		suffix = funcPath[slash+1:]
	}

	dot := strings.IndexByte(suffix, '.')
	if dot < 0 {
		if format == StacktraceFunctionFuncOnly {
			return suffix
		}

		return suffix
	}

	pkg := suffix[:dot]
	fn := suffix[dot+1:]

	switch format {
	case StacktraceFunctionFuncOnly:
		return fn
	case StacktraceFunctionPackageAndFunc:
		return pkg + "." + fn
	default:
		return ""
	}
}
