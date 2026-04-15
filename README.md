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

## Message Rendering

Use the default `Error()` / formatter APIs when you want diagnostics with source
locations, and the message helpers when you only want plain wrapped text.

- `err.Error()` gives you the full stack-aware debug view
- `errors.ChainMessages(err)` gives you `%w`-style wrapped text without source info
- `errors.RootMessage(err)` gives you the final underlying message only

```go
func createUser() error {
    err := fmt.Errorf("db write failed: %w", errors.New("connection reset"))
    err = errors.Wrapf(err, "create user")

    fmt.Println(err.Error())               // full-detail stack-aware output
    fmt.Println(errors.ChainMessages(err)) // create user: db write failed: connection reset
    fmt.Println(errors.RootMessage(err))   // connection reset

    return err
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

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
