package errors

import (
	"fmt"
	"runtime"
)

// Expand converts package-native error nodes into a resolved tree.
//
// Foreign errors, including stdlib wrappers and stdlib joins, are treated as
// opaque leaves. This intentionally preserves their original text,
// especially for fmt.Errorf calls with multiple %w operands where rebuilding the
// tree from Unwrap() []error would lose wrapper-local formatting.
func Expand(err error) *ErrorEntry {
	lazyFrames, entryRoot := expandNode([]lazyFrame{}, err)
	resolveFrames(lazyFrames)

	return entryRoot
}

// expandNode expands err into an ErrorEntry tree and queues any package frames
// that still need runtime resolution.
func expandNode(lazyFrames []lazyFrame, err error) ([]lazyFrame, *ErrorEntry) {
	if err == nil {
		return lazyFrames, nil
	}

	if errorChain, ok := IsErrorChain(err); ok {
		return expandChain(lazyFrames, errorChain)
	}

	if errorTree, ok := IsErrorTree(err); ok {
		return expandTree(lazyFrames, errorTree)
	}

	entry := newErrorEntry(err.Error())
	entry.Resolved.Foreign = true
	return lazyFrames, entry
}

// expandChain expands one package-owned single-child node.
func expandChain(lazyFrames []lazyFrame, chain *ErrorChain) ([]lazyFrame, *ErrorEntry) {
	entry := newErrorEntry(formatNodeMessage(chain.node.format, chain.node.params))

	var child *ErrorEntry
	lazyFrames, child = expandNode(lazyFrames, chain.child)
	if child != nil {
		entry.Children = append(entry.Children, child)
	}

	return appendLazyFrame(lazyFrames, entry, chain.node.pc), entry
}

// expandTree expands one package-owned multi-child node.
func expandTree(lazyFrames []lazyFrame, tree *ErrorTree) ([]lazyFrame, *ErrorEntry) {
	entry := newErrorEntry(formatNodeMessage(tree.node.format, tree.node.params))
	entry.Resolved.Multi = true

	var child *ErrorEntry
	for _, treeChild := range tree.children {
		lazyFrames, child = expandNode(lazyFrames, treeChild)
		if child != nil {
			entry.Children = append(entry.Children, child)
		}
	}

	return appendLazyFrame(lazyFrames, entry, tree.node.pc), entry
}

// lazyFrame pairs an unresolved program counter with the entry that should receive it.
type lazyFrame struct {
	pc    uintptr
	entry *ErrorEntry
}

// newLazyFrame creates a lazyFrame for one entry.
func newLazyFrame(entry *ErrorEntry, pc uintptr) lazyFrame {
	return lazyFrame{pc, entry}
}

// appendLazyFrame adds an entry to the resolution queue when it has a callsite.
func appendLazyFrame(frames []lazyFrame, entry *ErrorEntry, pc uintptr) []lazyFrame {
	if pc != 0 {
		frames = append(frames, newLazyFrame(entry, pc))
	}

	return frames
}

// formatNodeMessage formats the package-owned message stored in a node.
func formatNodeMessage(format string, params []any) string {
	return fmt.Sprintf(format, params...)
}

// resolveFrames resolves all queued program counters in one runtime pass.
func resolveFrames(lazyFrames []lazyFrame) {
	bufLen := len(lazyFrames)
	pcBuffer := make([]uintptr, bufLen)

	for i, f := range lazyFrames {
		pcBuffer[i] = f.pc
	}

	frames := runtime.CallersFrames(pcBuffer)

	i := 0
	for frame, more := frames.Next(); i < bufLen; frame, more = frames.Next() {
		lazyFrames[i].entry.Resolved.FilePath = frame.File
		lazyFrames[i].entry.Resolved.FuncPath = frame.Function
		lazyFrames[i].entry.Resolved.Line = frame.Line

		i++

		if !more {
			break
		}
	}
}
