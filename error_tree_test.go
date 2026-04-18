package errors

import (
	stderrors "errors"
	"fmt"
	"strings"
	"testing"
)

func TestJoinFiltersNilAndHandlesSingleChildCases(t *testing.T) {
	leaf := stderrors.New("leaf")
	singleWrap := fmt.Errorf("single: %w", leaf)
	multiWrap := fmt.Errorf("multi: %w + %w", stderrors.New("left"), stderrors.New("right"))
	stdlibJoin := stderrors.Join(stderrors.New("join left"), stderrors.New("join right"))

	if got := Join(nil, nil); got != nil {
		t.Fatalf("Join(nil, nil) = %v, want nil", got)
	}

	got := Join(nil, singleWrap, nil)
	tree, ok := IsErrorTree(got)
	if !ok {
		t.Fatalf("Join(singleWrap) returned %T, want *ErrorTree", got)
	}

	children := tree.Unwrap()
	if len(children) != 1 || children[0] != singleWrap {
		t.Fatalf("Join(singleWrap) children = %#v, want [%#v]", children, singleWrap)
	}

	if got := Join(nil, multiWrap, nil); got != multiWrap {
		t.Fatalf("Join(multiWrap) returned %T, want original multi-wrap error", got)
	}

	if got := Join(nil, stdlibJoin, nil); got != stdlibJoin {
		t.Fatalf("Join(stdlibJoin) returned %T, want original joined error", got)
	}
}

func TestErrorTreeErrorUsesChildErrorsVerbatim(t *testing.T) {
	leaf := stderrors.New("leaf")
	singleWrap := fmt.Errorf("single: %w", leaf)
	multiWrap := fmt.Errorf("multi: %w + %w", stderrors.New("left"), stderrors.New("right"))
	stdlibJoin := stderrors.Join(stderrors.New("join left"), stderrors.New("join right"))

	err := Join(leaf, singleWrap, multiWrap, stdlibJoin)
	tree, ok := IsErrorTree(err)
	if !ok {
		t.Fatalf("Join(...) returned %T, want *ErrorTree", err)
	}

	wantChildren := []error{leaf, singleWrap, multiWrap, stdlibJoin}
	gotChildren := tree.Unwrap()
	if len(gotChildren) != len(wantChildren) {
		t.Fatalf("len(Unwrap()) = %d, want %d", len(gotChildren), len(wantChildren))
	}

	for i := range wantChildren {
		if gotChildren[i] != wantChildren[i] {
			t.Fatalf("Unwrap()[%d] = %#v, want %#v", i, gotChildren[i], wantChildren[i])
		}
	}

	want := strings.Join([]string{
		leaf.Error(),
		singleWrap.Error(),
		multiWrap.Error(),
		stdlibJoin.Error(),
	}, "\n")

	if got := tree.Error(); got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}
