package errors

import (
	"fmt"
	"runtime"
)

// ErrorEntry is the textual representation of an error entry in the chain.
// It is used for error formatting in text, JSON and YAML.
type ErrorEntry struct {
	pc       uintptr
	Message  string `json:"message,omitempty" yaml:"message,omitempty"`
	FuncPath string `json:"func_path,omitempty" yaml:"func_path,omitempty"`
	FilePath string `json:"file_path,omitempty" yaml:"file_path,omitempty"`
	Line     int    `json:"line,omitempty" yaml:"line,omitempty"`
}

// Collapse unwraps all the errors in the chain (either ErrorChain or standard error)
// and returns a slice of ErrorEntry with string representation of the error
// and the stacktrace.
func Collapse(err error) (entries []ErrorEntry) {
	entries = []ErrorEntry{}

	for current := err; current != nil; {
		// var ec *ErrorChain
		if ec, ok := current.(*ErrorChain); ok {
			msg := ec.format
			if len(ec.params) > 0 {
				msg = fmt.Sprintf(ec.format, ec.params...)
			}

			entries = append(entries, ErrorEntry{
				Message: msg,
				pc:      ec.pc,
			})

			current = ec.inner
		} else {
			entries = append(entries, ErrorEntry{
				Message: current.Error(),
			})

			current = nil
		}
	}

	elen := len(entries)
	pcbuf := make([]uintptr, 0, elen)
	entriesWithPC := make([]int, 0, elen)

	for i, entry := range entries {
		if entry.pc != 0 {
			pcbuf = append(pcbuf, entry.pc)
			entriesWithPC = append(entriesWithPC, i)
		}
	}

	frames := runtime.CallersFrames(pcbuf)

	i := 0
	for frame, _ := frames.Next(); frame.PC != 0; frame, _ = frames.Next() {
		entryIndex := entriesWithPC[i]
		i++

		entries[entryIndex].FilePath = frame.File
		entries[entryIndex].Line = frame.Line
		entries[entryIndex].FuncPath = frame.Function
	}

	// reverse the entries
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return
}
