# errors

[![Go Reference](https://pkg.go.dev/badge/github.com/difof/errors.svg)](https://pkg.go.dev/github.com/difof/errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/difof/errors)](https://goreportcard.com/report/github.com/difof/errors)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/difof/errors)](https://golang.org/dl/)

A powerful drop-in replacement for Go's standard error handling with rich features:
- 📍 Stacktraces with source locations
- 🎯 Error wrapping with context
- 📝 Stackless message helpers
- 🔄 Error catching and result handling
- ⚡ Panic recovery utilities
- 🛡️ Assert functions
- 🎨 Formatted error messages
- 🎁 Quality of life error handling helpers

## 📦 Installation

Requires Go 1.21 or higher.

```bash
go get github.com/difof/errors
```

## 🏗️ Building from Source

```bash
# Clone the repository
git clone https://github.com/difof/errors.git
cd errors

# Install task (if not already installed)
go install github.com/go-task/task/v3/cmd/task@latest
# Or go to https://taskfile.dev/installation

# Run tests
task test

# Run benchmarks
task bench

# Run demo
task demo
```

## 🚀 Quick Start

```go
import "github.com/difof/errors"

func main() {
    if err := riskyOperation(); err != nil {
        fmt.Println(err.Error())               // Full-detail stack-aware output
        fmt.Println(errors.ChainMessages(err)) // Stackless wrapped messages
        fmt.Println(errors.RootMessage(err))   // Innermost/root message
    }
}

func riskyOperation() error {
    return errors.New("something went wrong")
}
```

## Demo

Run the bundled showcase:

```bash
task demo
```

No flags shows everything for every demo case:

- full-detail stack-aware output
- wrapped-message-only output via `errors.ChainMessages(err)`
- root-message-only output via `errors.RootMessage(err)`
- colored/JSON/YAML formatter views
- multiple scenarios, including package-only wraps, mixed `%w` wraps, and joined errors

Use flags to focus the showcase:

```bash
task demo -- -chain -root
task demo -- -full
task demo -- -color
task demo -- -json -yaml
```

Available flags:

- `-full` full stack-aware text output
- `-chain` stackless wrapped-message output
- `-root` innermost/root message only
- `-color` colored formatter output
- `-json` JSON formatter output
- `-yaml` YAML formatter output

## Basic Usage

### `New` / `Newf`

Create a new error with source location metadata attached.

```go
func loadConfig(path string) error {
    if path == "" {
        return errors.New("config path is empty")
    }

    return errors.Newf("config file %q is invalid", path)
}
```

### `Wrap` / `Wrapf`

Wrap an existing error to add context while keeping the original cause.

```go
func saveUser(user User) error {
    if err := writeUser(user); err != nil {
        return errors.Wrapf(err, "save user %d", user.ID)
    }

    return nil
}
```

### `WrapResult` / `WrapResultf`

Use result-aware wrap helpers when you want to return the value unchanged while
adding error context.

```go
func parsePort(raw string) (int, error) {
    port, err := strconv.Atoi(raw)
    return errors.WrapResultf(port, err)("parse port %q", raw)
}
```

### `Catch` / `Catchf`

Use `Catch` helpers as compact return helpers near the end of a function.

```go
func deleteUser(id int) error {
    err := repo.Delete(id)
    return errors.Catchf(err, "delete user %d", id)
}
```

### `CatchResult` / `CatchResultf` / `IgnoreResult`

Use result-aware catch helpers when a function returns a value and an error.

```go
func closeRows(rows *sql.Rows) error {
    return errors.CatchResult(rows, nil)(func(rows *sql.Rows) error {
        return rows.Close()
    })
}

func loadUser(id int) error {
    rows, err := db.Query("SELECT * FROM users WHERE id = ?", id)
    return errors.CatchResultf(rows, err)(
        errors.IgnoreResult[*sql.Rows](),
        "query user %d",
        id,
    )
}
```

### `Recover`

Convert panics into returned errors, especially when using `Must`.

```go
func loadSettings() (err error) {
    defer errors.Recover(&err)

    cfg := errors.MustResult(readConfig())
    errors.Must(validateConfig(cfg))

    return nil
}
```

### `Must` / `MustResult`

Use `Must` helpers when failure should panic and be handled by a higher-level
`Recover`.

```go
func bootstrap() (err error) {
    defer errors.Recover(&err)

    conn := errors.MustResult(openConnection())
    errors.Must(ping(conn))

    return nil
}
```

### Message Helpers

Use message helpers when you want plain text instead of full stack-aware output.

```go
func handler() error {
    err := errors.Wrapf(errors.New("permission denied"), "update account")

    log.Println(err.Error())               // detailed output with source locations
    log.Println(errors.ChainMessages(err)) // update account: permission denied
    log.Println(errors.RootMessage(err))   // permission denied

    return err
}
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
