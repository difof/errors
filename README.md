# errors

[![Go Reference](https://pkg.go.dev/badge/github.com/difof/errors.svg)](https://pkg.go.dev/github.com/difof/errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/difof/errors)](https://goreportcard.com/report/github.com/difof/errors)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/difof/errors)](https://golang.org/dl/)

A powerful drop-in replacement for Go's standard error handling with rich features:
- 📍 Stacktraces with source locations
- 🎯 Error wrapping with context
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
        fmt.Println(err.Error()) // Prints beautiful stacktrace
    }
}

func riskyOperation() error {
    return errors.New("something went wrong")
}
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
