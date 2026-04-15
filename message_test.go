package errors

import (
	goerrors "errors"
	"fmt"
	"strings"
	"testing"
)

func TestRootMessage(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := RootMessage(nil); got != "" {
			t.Fatalf("RootMessage(nil) = %q, want empty string", got)
		}
	})

	t.Run("plain std error", func(t *testing.T) {
		err := goerrors.New("inner")
		if got := RootMessage(err); got != "inner" {
			t.Fatalf("RootMessage() = %q, want %q", got, "inner")
		}
	})

	t.Run("package chain", func(t *testing.T) {
		err := Wrapf(Wrapf(New("inner"), "middle"), "outer")
		if got := RootMessage(err); got != "inner" {
			t.Fatalf("RootMessage() = %q, want %q", got, "inner")
		}
	})

	t.Run("empty wrapper", func(t *testing.T) {
		err := Wrap(New("inner"))
		if got := RootMessage(err); got != "inner" {
			t.Fatalf("RootMessage() = %q, want %q", got, "inner")
		}
	})

	t.Run("stdlib around package chain", func(t *testing.T) {
		err := fmt.Errorf("outer: %w", Wrapf(New("inner"), "middle"))
		if got := RootMessage(err); got != "inner" {
			t.Fatalf("RootMessage() = %q, want %q", got, "inner")
		}
	})

	t.Run("package around stdlib chain", func(t *testing.T) {
		err := Wrapf(fmt.Errorf("middle: %w", goerrors.New("inner")), "outer")
		if got := RootMessage(err); got != "inner" {
			t.Fatalf("RootMessage() = %q, want %q", got, "inner")
		}
	})

	t.Run("joined leaf", func(t *testing.T) {
		err := Wrapf(goerrors.Join(goerrors.New("a"), goerrors.New("b")), "outer")
		want := "a\nb"
		if got := RootMessage(err); got != want {
			t.Fatalf("RootMessage() = %q, want %q", got, want)
		}
	})
}

func TestChainMessages(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := ChainMessages(nil); got != "" {
			t.Fatalf("ChainMessages(nil) = %q, want empty string", got)
		}
	})

	t.Run("plain std error", func(t *testing.T) {
		err := goerrors.New("inner")
		if got := ChainMessages(err); got != "inner" {
			t.Fatalf("ChainMessages() = %q, want %q", got, "inner")
		}
	})

	t.Run("new", func(t *testing.T) {
		err := New("inner")
		if got := ChainMessages(err); got != "inner" {
			t.Fatalf("ChainMessages() = %q, want %q", got, "inner")
		}
	})

	t.Run("empty wrapper omitted", func(t *testing.T) {
		err := Wrap(New("inner"))
		if got := ChainMessages(err); got != "inner" {
			t.Fatalf("ChainMessages() = %q, want %q", got, "inner")
		}
	})

	t.Run("package chain", func(t *testing.T) {
		err := Wrapf(Wrapf(New("inner"), "middle"), "outer")
		if got := ChainMessages(err); got != "outer: middle: inner" {
			t.Fatalf("ChainMessages() = %q, want %q", got, "outer: middle: inner")
		}
	})

	t.Run("stdlib wrap around package chain", func(t *testing.T) {
		err := fmt.Errorf("outer: %w", Wrapf(New("inner"), "middle"))
		if got := ChainMessages(err); got != "outer: middle: inner" {
			t.Fatalf("ChainMessages() = %q, want %q", got, "outer: middle: inner")
		}
	})

	t.Run("package wrap around stdlib chain", func(t *testing.T) {
		err := Wrapf(fmt.Errorf("middle: %w", goerrors.New("inner")), "outer")
		if got := ChainMessages(err); got != "outer: middle: inner" {
			t.Fatalf("ChainMessages() = %q, want %q", got, "outer: middle: inner")
		}
	})

	t.Run("nontrivial stdlib wrapper text", func(t *testing.T) {
		err := fmt.Errorf("while saving (%w)", Wrapf(New("inner"), "middle"))
		if got := ChainMessages(err); got != "while saving (middle: inner)" {
			t.Fatalf("ChainMessages() = %q, want %q", got, "while saving (middle: inner)")
		}
	})

	t.Run("joined leaf preserved", func(t *testing.T) {
		err := Wrapf(goerrors.Join(goerrors.New("a"), goerrors.New("b")), "outer")
		want := "outer: a\nb"
		if got := ChainMessages(err); got != want {
			t.Fatalf("ChainMessages() = %q, want %q", got, want)
		}
	})

	t.Run("no stack metadata", func(t *testing.T) {
		err := Wrapf(New("inner"), "outer")
		got := ChainMessages(err)

		if strings.Contains(got, "at ") {
			t.Fatalf("ChainMessages() = %q, unexpectedly contains stack marker", got)
		}
		if strings.Contains(got, ".go:") {
			t.Fatalf("ChainMessages() = %q, unexpectedly contains source location", got)
		}
	})
}
