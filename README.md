# errors

[![Go Reference](https://pkg.go.dev/badge/github.com/difof/errors.svg)](https://pkg.go.dev/github.com/difof/errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/difof/errors)](https://goreportcard.com/report/github.com/difof/errors)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/difof/errors)](https://golang.org/dl/)

`errors` is a higher-level error handling package for Go.

It is not trying to replace the standard library. It is for the codebases where
plain `fmt.Errorf("...: %w", err)` everywhere starts to feel thin: you want
consistent callsite-aware errors, better wrapping helpers, panic-to-error flow,
and a readable stacktrace view when things go sideways.

The package gives you a structured error tree for package-owned errors, plus
small utilities that make the usual boring error plumbing less annoying:

- `New`, `Wrap`, `Join`
- `Catch`, `CatchResult`
- `Must`, `MustResult`, `Recover`
- `Stacktrace`

Why use it instead of raw stdlib errors alone?

Because most projects end up reinventing the same patterns anyway. The stdlib
version of a setup flow usually turns into a staircase of `if err != nil`
checks. Here you can write the same thing more directly:

```go
func bootstrap() (err error) {
	defer errors.Recover(&err)

	errors.Must(someFunc())

	r := errors.MustResult(someResultingFunc())
	_ = r

	return 
}
```

It still plays fine with stdlib errors, but it is opinionated about one thing:
package-owned errors are fully structured, stdlib-created errors are treated as
opaque text during expansion. That keeps `fmt.Errorf` and `errors.Join` output
intact instead of trying to reverse-engineer formatting that Go does not expose.

## Installation

Requires Go 1.21 or higher.

```bash
go get github.com/difof/errors
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/difof/errors"
)

func main() {
	err := saveUser(42)
	if err == nil {
		return
	}

	fmt.Println(err.Error())
	fmt.Println(errors.Stacktrace(err, errors.StacktraceWithTrimFilePath(true)))
}

func saveUser(id int) error {
	return errors.Wrapf(errors.New("disk offline"), "save user %d", id)
}
```

Typical output shape:

```text
save user 42: disk offline

disk offline
| at main.go:main.saveUser:18
| at main.go:main.saveUser:18: save user 42
```

The first line is the normal error string you can return, compare, or log. The
second view is the debugging view.

## Demo

Run the bundled showcase:

```bash
task demo
```

Or run it directly:

```bash
go run ./demo
```

By default the demo shows, for each scenario:

- the code used to construct the error
- the normal `Error()` output
- the plain stacktrace view
- the colored stacktrace view

Useful flags:

```bash
go run ./demo -error
go run ./demo -stacktrace
go run ./demo -color
go run ./demo -full
```

`-full` is just a convenience alias for the plain stacktrace view.

The demo includes package-only chains, package joins, stdlib joins wrapped by
package errors, and mixed `fmt.Errorf` / `errors.Join` scenarios so you can see
exactly where the package is structured and where stdlib errors stay opaque.

## Basic Usage

### `New` / `Newf`

Use `New` when you want a real package-owned error node with callsite metadata.

```go
func loadConfig(path string) error {
	if path == "" {
		return errors.New("config path is empty")
	}

	return errors.Newf("config file %q is invalid", path)
}
```

### `Wrap` / `Wrapf`

Use `Wrap` to add context without flattening the original cause into a dead
string.

```go
func saveUser(user User) error {
	if err := writeUser(user); err != nil {
		return errors.Wrapf(err, "save user %d", user.ID)
	}

	return nil
}
```

### `Join`

Use `Join` when you actually have multiple errors and want them to stay multiple.

```go
func shutdown() error {
	return errors.Join(
		stopHTTP(),
		stopWorkers(),
		flushMetrics(),
	)
}
```

### `WrapResult` / `WrapResultf`

Useful when you want to keep the returned value untouched and only wrap the
error.

```go
func parsePort(raw string) (int, error) {
	port, err := strconv.Atoi(raw)
	return errors.WrapResultf(port, err)("parse port %q", raw)
}
```

### `Catch` / `Catchf`

Good near the bottom of a function when you just want “return this error with
context if it exists”.

```go
func deleteUser(id int) error {
	err := repo.Delete(id)
	return errors.Catchf(err, "delete user %d", id)
}
```

### `CatchResult` / `CatchResultf`

These are handy when a function returns `(value, error)` and you want to handle
the success path inline without losing the error path.

```go
func loadUser(id int) error {
	rows, err := db.Query("SELECT * FROM users WHERE id = ?", id)
	return errors.CatchResultf(rows, err)(func(rows *sql.Rows) error {
		defer rows.Close()
		return scanUser(rows)
	}, "query user %d", id)
}
```

### `Must` / `MustResult` / `Recover`

This combo is for flows where panicking locally and converting back to an error
at the boundary is cleaner than threading checks through every line.

```go
func bootstrap() (err error) {
	defer errors.Recover(&err)

	conn := errors.MustResult(openConnection())
	errors.Must(ping(conn))

	return nil
}
```

That style is not for every function, but in setup code, orchestration code, or
batch flows it can make the happy path much easier to read.

## Stacktrace

`Stacktrace` is the human-facing debug view. It renders package-owned chains and
joins structurally and lets you tune the output shape.

```go
trace := errors.Stacktrace(
	err,
	errors.StacktraceWithTrimFilePath(true),
	errors.StacktraceWithSuppressEmptyFrames(true),
	errors.StacktraceWithTreePrefix("|"),
)
```

There are options for indentation, colors, branch labels, tree prefixes, file
path trimming, function formatting, and suppressing empty frame lines.

## Stdlib Interop

This package works with stdlib errors, but it does not try to pretend that all
stdlib errors are structurally recoverable.

Current rule of thumb:

- package errors wrapping stdlib errors: package nodes stay structured, stdlib
  children remain opaque
- stdlib errors wrapping package errors: the outer stdlib wrapper is treated as
  opaque text

That means:

```go
errors.Wrapf(fmt.Errorf("outer: %w", err), "save user")
```

keeps `save user` as a package-owned node, but the `fmt.Errorf` part stays a
foreign leaf.

And:

```go
fmt.Errorf("outer std: %w", errors.Join(a, b))
```

is treated as one opaque stdlib wrapper, even though the inner value is a
package error tree.

That is intentional. Go exposes `Unwrap() []error` for multi-wrap `fmt.Errorf`,
but it does not expose the wrapper-local formatting in a way that lets this
package rebuild the tree without risking broken or misleading output. Better to
keep stdlib-created formatting intact than fake precision.

There is a longer note on that tradeoff in [`STDLIB-INTEROP.md`](STDLIB-INTEROP.md).

## Building from Source

```bash
git clone https://github.com/difof/errors.git
cd errors
task test
task demo
```

## Contributing

Contributions are welcome. If the package is useful to you and you have a clean
idea to improve it, send a PR.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
