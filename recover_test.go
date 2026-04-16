package errors

import (
	goerrors "errors"
	"testing"
)

func TestRecoverHelpers(t *testing.T) {
	t.Run("recoverError nil", func(t *testing.T) {
		if err := recoverError(nil, 0); err != nil {
			t.Fatalf("recoverError(nil) = %v, want nil", err)
		}
	})

	t.Run("recoverError returns error panic as is", func(t *testing.T) {
		boom := goerrors.New("boom")
		if err := recoverError(boom, 0); !goerrors.Is(err, boom) {
			t.Fatalf("recoverError(error) = %v, want original error", err)
		}
	})

	t.Run("recoverError wraps non-error panic", func(t *testing.T) {
		err := recoverError("boom", 0)
		if got := RootMessage(err); got != "boom" {
			t.Fatalf("RootMessage(recoverError(...)) = %q, want %q", got, "boom")
		}
	})

	t.Run("Recover ignores nil error pointer", func(t *testing.T) {
		func() {
			defer Recover(nil)
		}()
	})

	t.Run("Recover captures panic error", func(t *testing.T) {
		var err error
		func() {
			defer Recover(&err)
			panic(goerrors.New("boom"))
		}()
		if got := RootMessage(err); got != "boom" {
			t.Fatalf("RootMessage(Recover) = %q, want %q", got, "boom")
		}
	})

	t.Run("Recover wraps non-error panic", func(t *testing.T) {
		var err error
		func() {
			defer Recover(&err)
			panic("boom")
		}()
		if got := RootMessage(err); got != "boom" {
			t.Fatalf("RootMessage(Recover string panic) = %q, want %q", got, "boom")
		}
	})

	t.Run("RecoverFn captures panic", func(t *testing.T) {
		var gotErr error
		func() {
			defer RecoverFn(func(err error) {
				gotErr = err
			})
			panic(goerrors.New("boom"))
		}()
		if got := RootMessage(gotErr); got != "boom" {
			t.Fatalf("RootMessage(RecoverFn) = %q, want %q", got, "boom")
		}
	})
}
