package errors

import "testing"

func TestNewHelpers(t *testing.T) {
	t.Run("Newf formats message", func(t *testing.T) {
		err := Newf("hello %s", "world")
		if got := RootMessage(err); got != "hello world" {
			t.Fatalf("RootMessage(Newf(...)) = %q, want %q", got, "hello world")
		}
	})

	t.Run("NewSkipf formats message", func(t *testing.T) {
		err := NewSkipf(0, "hello %s", "world")
		if got := RootMessage(err); got != "hello world" {
			t.Fatalf("RootMessage(NewSkipf(...)) = %q, want %q", got, "hello world")
		}
	})
}
