package errors

import (
	stderrors "errors"
	"runtime"
	"strings"
	"testing"
)

type captureSite struct {
	file string
	fn   string
	line int
}

func currentSite(offset int) captureSite {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return captureSite{}
	}

	return captureSite{
		file: file,
		fn:   runtime.FuncForPC(pc).Name(),
		line: line + offset,
	}
}

func assertExactRootSite(t *testing.T, err error, want captureSite) {
	t.Helper()

	entry := Expand(err)
	if entry == nil {
		t.Fatal("Expand(err) returned nil")
	}

	got := entry.Resolved
	if got.FilePath != want.file || got.FuncPath != want.fn || got.Line != want.line {
		t.Fatalf("got %s %s:%d, want %s %s:%d", got.Message, got.FilePath, got.Line, want.file, want.fn, want.line)
	}
}

func assertApproxRecoveredRootSite(t *testing.T, err error, want captureSite, funcContains string) {
	t.Helper()

	entry := Expand(err)
	if entry == nil {
		t.Fatal("Expand(err) returned nil")
	}

	got := entry.Resolved
	if got.FilePath != want.file {
		t.Fatalf("got file %s, want %s", got.FilePath, want.file)
	}

	if funcContains != "" && !strings.Contains(got.FuncPath, funcContains) {
		t.Fatalf("got func %s, want it to contain %s", got.FuncPath, funcContains)
	}

	delta := got.Line - want.line
	if delta < 0 {
		delta = -delta
	}

	if delta > 1 {
		t.Fatalf("got line %d, want near %d", got.Line, want.line)
	}
}

func captureNew() (error, captureSite) {
	want := currentSite(1)
	err := New("boom")
	return err, want
}

func captureNewf() (error, captureSite) {
	want := currentSite(1)
	err := Newf("boom %d", 7)
	return err, want
}

func captureWrap() (error, captureSite) {
	want := currentSite(1)
	err := Wrap(stderrors.New("boom"))
	return err, want
}

func captureWrapf() (error, captureSite) {
	want := currentSite(1)
	err := Wrapf(stderrors.New("boom"), "ctx")
	return err, want
}

func captureWrapResult() (error, captureSite) {
	want := currentSite(1)
	_, err := WrapResult("x", stderrors.New("boom"))
	return err, want
}

func captureWrapResultf() (error, captureSite) {
	f := WrapResultf("x", stderrors.New("boom"))
	want := currentSite(1)
	_, err := f("ctx")
	return err, want
}

func captureCatch() (error, captureSite) {
	want := currentSite(1)
	err := Catch(stderrors.New("boom"))
	return err, want
}

func captureCatchf() (error, captureSite) {
	want := currentSite(1)
	err := Catchf(stderrors.New("boom"), "ctx")
	return err, want
}

func captureCatchResultInputErr() (error, captureSite) {
	f := CatchResult("x", stderrors.New("boom"))
	want := currentSite(1)
	err := f(func(string) error { return nil })
	return err, want
}

func captureCatchResultCallbackErr() (error, captureSite) {
	f := CatchResult("x", nil)
	want := currentSite(1)
	err := f(func(string) error { return stderrors.New("boom") })
	return err, want
}

func captureCatchResultfInputErr() (error, captureSite) {
	f := CatchResultf("x", stderrors.New("boom"))
	want := currentSite(1)
	err := f(IgnoreResult[string](), "ctx")
	return err, want
}

func captureCatchResultfCallbackErr() (error, captureSite) {
	f := CatchResultf("x", nil)
	want := currentSite(1)
	err := f(func(string) error { return stderrors.New("boom") }, "ctx")
	return err, want
}

func captureMust() (err error, want captureSite) {
	want = currentSite(2)
	defer Recover(&err)
	Must(stderrors.New("boom"))
	return nil, want
}

func captureMustf() (err error, want captureSite) {
	p := Mustf(stderrors.New("boom"))
	want = currentSite(2)
	defer Recover(&err)
	p("ctx")
	return nil, want
}

func captureMustResult() (err error, want captureSite) {
	want = currentSite(2)
	defer Recover(&err)
	_ = MustResult("x", stderrors.New("boom"))
	return nil, want
}

func captureMustResultf() (err error, want captureSite) {
	p := MustResultf("x", stderrors.New("boom"))
	want = currentSite(2)
	defer Recover(&err)
	_ = p("ctx")
	return nil, want
}

func captureMustResult2() (err error, want captureSite) {
	want = currentSite(2)
	defer Recover(&err)
	_, _ = MustResult2("x", "y", stderrors.New("boom"))
	return nil, want
}

func captureMustResult2f() (err error, want captureSite) {
	p := MustResult2f("x", "y", stderrors.New("boom"))
	want = currentSite(2)
	defer Recover(&err)
	_, _ = p("ctx")
	return nil, want
}

func captureRecoverAssert() (err error, want captureSite) {
	want = currentSite(2)
	defer Recover(&err)
	Assert(false, "boom")
	return nil, want
}

func captureRecoverAssertf() (err error, want captureSite) {
	want = currentSite(2)
	defer Recover(&err)
	Assertf(false, "boom %d", 7)
	return nil, want
}

func captureRecoverPanic() (err error, want captureSite) {
	want = currentSite(2)
	defer Recover(&err)
	panic("boom")
}

func captureRecoverFnPanic() (error, captureSite) {
	var err error
	want := currentSite(5)
	func() {
		defer RecoverFn(func(recovered error) {
			err = recovered
		})
		panic("boom")
	}()
	return err, want
}

func TestCallsiteCapture(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		err, want := captureNew()
		assertExactRootSite(t, err, want)
	})

	t.Run("Newf", func(t *testing.T) {
		err, want := captureNewf()
		assertExactRootSite(t, err, want)
	})

	t.Run("Wrap", func(t *testing.T) {
		err, want := captureWrap()
		assertExactRootSite(t, err, want)
	})

	t.Run("Wrapf", func(t *testing.T) {
		err, want := captureWrapf()
		assertExactRootSite(t, err, want)
	})

	t.Run("WrapResult", func(t *testing.T) {
		err, want := captureWrapResult()
		assertExactRootSite(t, err, want)
	})

	t.Run("WrapResultf", func(t *testing.T) {
		err, want := captureWrapResultf()
		assertExactRootSite(t, err, want)
	})

	t.Run("Catch", func(t *testing.T) {
		err, want := captureCatch()
		assertExactRootSite(t, err, want)
	})

	t.Run("Catchf", func(t *testing.T) {
		err, want := captureCatchf()
		assertExactRootSite(t, err, want)
	})

	t.Run("CatchResult input err", func(t *testing.T) {
		err, want := captureCatchResultInputErr()
		assertExactRootSite(t, err, want)
	})

	t.Run("CatchResult callback err", func(t *testing.T) {
		err, want := captureCatchResultCallbackErr()
		assertExactRootSite(t, err, want)
	})

	t.Run("CatchResultf input err", func(t *testing.T) {
		err, want := captureCatchResultfInputErr()
		assertExactRootSite(t, err, want)
	})

	t.Run("CatchResultf callback err", func(t *testing.T) {
		err, want := captureCatchResultfCallbackErr()
		assertExactRootSite(t, err, want)
	})

	t.Run("Must", func(t *testing.T) {
		err, want := captureMust()
		assertExactRootSite(t, err, want)
	})

	t.Run("Mustf", func(t *testing.T) {
		err, want := captureMustf()
		assertExactRootSite(t, err, want)
	})

	t.Run("MustResult", func(t *testing.T) {
		err, want := captureMustResult()
		assertExactRootSite(t, err, want)
	})

	t.Run("MustResultf", func(t *testing.T) {
		err, want := captureMustResultf()
		assertExactRootSite(t, err, want)
	})

	t.Run("MustResult2", func(t *testing.T) {
		err, want := captureMustResult2()
		assertExactRootSite(t, err, want)
	})

	t.Run("MustResult2f", func(t *testing.T) {
		err, want := captureMustResult2f()
		assertExactRootSite(t, err, want)
	})

	t.Run("Recover Assert", func(t *testing.T) {
		err, want := captureRecoverAssert()
		assertApproxRecoveredRootSite(t, err, want, "captureRecoverAssert")
	})

	t.Run("Recover Assertf", func(t *testing.T) {
		err, want := captureRecoverAssertf()
		assertApproxRecoveredRootSite(t, err, want, "captureRecoverAssertf")
	})

	t.Run("Recover panic", func(t *testing.T) {
		err, want := captureRecoverPanic()
		assertApproxRecoveredRootSite(t, err, want, "captureRecoverPanic")
	})

	t.Run("RecoverFn panic", func(t *testing.T) {
		err, want := captureRecoverFnPanic()
		assertApproxRecoveredRootSite(t, err, want, "captureRecoverFnPanic")
	})
}
