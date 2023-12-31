# errors

Drop-in replacement for golang's standard error handling with readable stacktrace and source location.

This package also adds some QoL error handling helpers.

Error messages look like this:

```
at github.com/difof/errors.c /dev/errors/error_test.go:17: in 'c' error from 'messageError'
at github.com/difof/errors.b /dev/errors/error_test.go:13: in 'b' error from 'c'
at github.com/difof/errors.a /dev/errors/error_test.go:9: in 'a' error from 'b'
at github.com/difof/errors.TestHasError.func1 /dev/errors/error_test.go:29
at github.com/difof/errors.TestHasError /dev/errors/error_test.go:32
```

# Usage

You need **go +1.18**

`go get github.com/difof/errors`

And import in your code

```go
import "github.com/difof/errors"
```

**Docs:** [pkg.go.dev](https://pkg.go.dev/github.com/difof/errors)

## Creating errors

You can create both unformatted and formatted errors:

```go
err1 := errors.New("error message")
err2 := errors.Newf("error %s", "message")
errors.Assert(!errors.Is(err1, err2), "errors are equal")
```

## Wrapping errors

Wrapping existing errors adds current stacktrace to the error:

```go
func ReadJson(input []byte) (data MyData, err error) {
    err = errors.Wrapf(json.Unmarshal(input, &data), "failed to unmarshal json")
    return
}
```

## Printing errors

You can print errors with stacktrace:

```go
func RecursiveFunction(i int) (err error) {
    if i > 3 {
        return errors.New("end of recursion")
    }

    return errors.Wrap(RecursiveFunction(i++))
}

func main() {
    fmt.Println(RecursiveFunction(0).Error())
}
```

## Error catching

Sometimes you return an error as last statement anyways,
so instead of using if and wrap then return, you do it in one line:

```go
func Bar() error {
    return errors.New("error message")
}

func Foo() error {
    // This is equivalent to:
    //
    //   if err := Bar(); err != nil {
    //       return errors.Wrapf(err, "something bad happened")
    //   }
    //   return nil
    return errors.Catchf(Bar(), "something bad happened")
}
```

And if you're dealing with a result and an error returned from functions,
but you only return the error or handle the result, you can use `CatchResult`:

```go
func Bar() (int, error) {
    return 0, errors.New("error message")
}

func Foo() error {
    // This is equivalent to:
    //
    //   result, err := Bar()
    //   if err != nil {
    //       return 0, errors.Wrapf(err, "something bad happened")
    //   }
    //   return result, nil
    return errors.CatchResultf(
        Bar(),
    )(
        // This function is called if Bar doesn't return an error.
        // Returning an error from this function will cause CatchResultf to wrap and return it.
        func(result int) error {
            // Deal with result in callback hell!

            return nil
        },
        "something bad happened",
    )
}

func DoQuery() (result MyData, err error) {
    sqlStatement := queryBuilder.Build()

    err = errors.CatchResultf(
        db.Query(sqlStatement, &result),
    )(
        errors.IgnoreResult[sql.Result](),
        "failed to execute query",
    )

    return
}
```

## QoL `if err != nil` helper

Most of the time you have to constantly check if an error is not nil, and handle it many times in every function.
It would be nice automatically return error or handle the result like in other languages such as Zig.

You can use the `Must`, `Mustf`, `MustResult`, `MustResultf` functions and defer call to `Recover` to do this:

```go
func Foo() (err error) {
    defer errors.Recover(&err)

    filename := "foo.txt"
    file := errors.Mustf(os.Open(filename))("failed to open file %s", filename)
    defer errors.Mustf(file.Close())("failed to close file %s", filename)

    // and so on...

    return
}
```

## Other utilities

#### Assert

```go
errors.Assert(condition bool, message string)
```

```go
errors.Assertf(condition bool, format string, args ...any)
```

#### Blasphemy

```go
errors.Ignore[T any](r T, _ error) T
```
