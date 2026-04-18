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

type StacktraceBranchLabelFormatter func(index int) string
type StacktraceTreePrefixFormatter func(colorEnabled bool) string

type StacktraceFunctionFormat int

const (
	StacktraceFunctionPackageAndFunc StacktraceFunctionFormat = iota
	StacktraceFunctionFuncOnly
	StacktraceFunctionNone
)

type StacktraceColors struct {
	Source  *color.Color
	Func    *color.Color
	Message *color.Color
}

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

type StacktraceOption func(opt *StacktraceOptions)

func StacktraceWithIndent(spaces int) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.Indent = spaces
	}
}

func StacktraceWithPreIndent(spaces int) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.PreIndent = spaces
	}
}

func StacktraceWithColor(color bool) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.Color = color
	}
}

func StacktraceWithSuppressEmptyFrames(suppress bool) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.SuppressEmptyFrames = suppress
	}
}

func StacktraceWithTrimFilePath(trim bool) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.TrimFilePath = trim
	}
}

func StacktraceWithFunctionFormat(format StacktraceFunctionFormat) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.FunctionFormat = format
	}
}

func StacktraceWithTreePrefix(prefix string) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.TreePrefixFormatter = func(bool) string { return prefix }
	}
}

func StacktraceWithTreePrefixFormatter(formatter StacktraceTreePrefixFormatter) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.TreePrefixFormatter = formatter
	}
}

func StacktraceWithBranchLabel(formatter StacktraceBranchLabelFormatter) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.BranchLabel = formatter
	}
}

func StacktraceWithColors(colors StacktraceColors) StacktraceOption {
	return func(opt *StacktraceOptions) {
		opt.Colors = colors
	}
}

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

type stacktraceRenderer struct {
	sb           strings.Builder
	options      *StacktraceOptions
	needsNewline bool
}

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

func (r *stacktraceRenderer) renderPackageChain(path []*ErrorEntry, depth int, firstLineLabel string) {
	if len(path) == 0 {
		return
	}

	leaf := path[len(path)-1]
	hasRootMessage := leaf.Resolved.Message != ""
	if hasRootMessage {
		r.writeLine(formatMessageLine(depth, firstLineLabel, leaf.Resolved.Message, r.options))
	}

	frameDepth := depth
	firstFrameLabel := firstLineLabel
	if hasRootMessage {
		frameDepth++
		firstFrameLabel = ""
	}

	for i := len(path) - 1; i >= 0; i-- {
		label := firstFrameLabel
		firstFrameLabel = ""

		message := path[i].Resolved.Message
		if i == len(path)-1 && hasRootMessage {
			message = ""
		}

		if shouldSuppressFrame(message, r.options) {
			firstFrameLabel = label
			continue
		}

		r.writeLine(formatFrameLine(frameDepth, label, &path[i].Resolved, message, r.options))
	}
}

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

	for i := len(path) - 1; i >= 0; i-- {
		label := firstFrameLabel
		firstFrameLabel = ""

		message := path[i].Resolved.Message
		if shouldSuppressFrame(message, r.options) {
			firstFrameLabel = label
			continue
		}

		r.writeLine(formatFrameLine(frameDepth, label, &path[i].Resolved, message, r.options))
	}
}

func (r *stacktraceRenderer) renderMultiBranch(path []*ErrorEntry, multi *ErrorEntry, depth int, firstLineLabel string) {
	currentDepth := depth
	label := firstLineLabel

	for _, node := range path {
		labelForNode := label
		if shouldSuppressFrame(node.Resolved.Message, r.options) {
			label = labelForNode
			continue
		}

		r.writeLine(formatFrameLine(currentDepth, labelForNode, &node.Resolved, node.Resolved.Message, r.options))
		label = ""
		currentDepth++
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

func (r *stacktraceRenderer) writeLine(line string) {
	if r.needsNewline {
		r.sb.WriteString("\n")
	}

	r.sb.WriteString(line)
	r.needsNewline = true
}

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

func defaultStacktraceBranchLabel(index int) string {
	return "[" + strconv.Itoa(index) + "] "
}

func shouldSuppressFrame(message string, options *StacktraceOptions) bool {
	return options.SuppressEmptyFrames && message == ""
}

func defaultStacktraceTreePrefix(bool) string {
	return "|"
}

func defaultStacktraceColors() StacktraceColors {
	return StacktraceColors{
		Source:  stacktraceSourceColor,
		Func:    stacktraceFuncColor,
		Message: stacktraceMessageColor,
	}
}

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

func formatMessageLine(depth int, label, message string, options *StacktraceOptions) string {
	indent := indentPrefix(depth, options)
	renderedPrefix := indent + label
	continuationPrefix := indent + strings.Repeat(" ", visibleWidth(label))
	return formatIndentedMessage(renderedPrefix, continuationPrefix, message, options)
}

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

func indentPrefix(depth int, options *StacktraceOptions) string {
	prefix := strings.Repeat(" ", options.PreIndent)
	if depth <= 0 {
		return prefix
	}

	return prefix + strings.Repeat(indentUnit(options), depth)
}

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

func colorizeSource(value string, options *StacktraceOptions) string {
	if options.Color && options.Colors.Source != nil {
		return options.Colors.Source.Sprint(value)
	}

	return value
}

func colorizeFunc(value string, options *StacktraceOptions) string {
	if options.Color && options.Colors.Func != nil {
		return options.Colors.Func.Sprint(value)
	}

	return value
}

func colorizeMessage(value string, options *StacktraceOptions) string {
	if options.Color && options.Colors.Message != nil {
		return options.Colors.Message.Sprint(value)
	}

	return value
}

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

func visibleWidth(value string) int {
	return len(ansiRegexp.ReplaceAllString(value, ""))
}

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
