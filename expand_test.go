package errors

import (
	stderrors "errors"
	"fmt"
	"testing"
)

func TestExpandNilAndForeignLeaf(t *testing.T) {
	if got := Expand(nil); got != nil {
		t.Fatalf("Expand(nil) = %#v, want nil", got)
	}

	foreign := fmt.Errorf("outer: %w", stderrors.New("boom"))
	entry := Expand(foreign)

	assertEntry(t, entry, foreign.Error(), true, false, 0)
}

func TestExpandPackageChainBuildsNestedEntries(t *testing.T) {
	err := WrapSkipf(0, stderrors.New("boom"), "load %s", "user")
	entry := Expand(err)

	assertEntry(t, entry, "load user", false, false, 1)
	assertEntry(t, entry.Children[0], "boom", true, false, 0)
}

func TestExpandPackageTreePreservesPackageNodesAndKeepsForeignWrappersOpaque(t *testing.T) {
	leaf := stderrors.New("leaf")
	singleWrap := fmt.Errorf("single: %w", stderrors.New("wrapped"))
	multiWrap := fmt.Errorf("multi: %w + %w", stderrors.New("left"), stderrors.New("right"))
	stdlibJoin := stderrors.Join(stderrors.New("join left"), stderrors.New("join right"))
	nested := Join(NewSkipf(0, "nested package"), stderrors.New("nested foreign"))

	err := Join(leaf, singleWrap, multiWrap, stdlibJoin, nested)
	entry := Expand(err)

	assertEntry(t, entry, "", false, true, 5)
	assertEntry(t, entry.Children[0], leaf.Error(), true, false, 0)
	assertEntry(t, entry.Children[1], singleWrap.Error(), true, false, 0)
	assertEntry(t, entry.Children[2], multiWrap.Error(), true, false, 0)
	assertEntry(t, entry.Children[3], stdlibJoin.Error(), true, false, 0)

	nestedEntry := entry.Children[4]
	assertEntry(t, nestedEntry, "", false, true, 2)
	assertEntry(t, nestedEntry.Children[0], "nested package", false, false, 0)
	assertEntry(t, nestedEntry.Children[1], "nested foreign", true, false, 0)
}

func assertEntry(t *testing.T, entry *ErrorEntry, message string, foreign, multi bool, childCount int) {
	t.Helper()

	if entry == nil {
		t.Fatal("entry is nil")
	}

	if entry.Resolved.Message != message {
		t.Fatalf("Resolved.Message = %q, want %q", entry.Resolved.Message, message)
	}

	if entry.Resolved.Foreign != foreign {
		t.Fatalf("Resolved.Foreign = %v, want %v", entry.Resolved.Foreign, foreign)
	}

	if entry.Resolved.Multi != multi {
		t.Fatalf("Resolved.Multi = %v, want %v", entry.Resolved.Multi, multi)
	}

	if len(entry.Children) != childCount {
		t.Fatalf("len(Children) = %d, want %d", len(entry.Children), childCount)
	}
}
