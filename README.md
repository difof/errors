# errors

Drop-in replacement for golang's standard error handling with readable stacktrace and source location.

This package includes a few extra utilities for error handling.

# Usage

You need **go +1.18**

`go get github.com/difof/errors`

And import in your code

```go
import "github.com/difof/errors"
```

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

## Other utilities

#### Assert

```go
errors.Assert(condition bool, message string)
```

```go
errors.Assertf(condition bool, format string, args ...any)
```

#### Death on error

```go
errors.Must[T any](r T, err error) T
```

```go
jb := errors.Must(json.Marshal(data))
```

You also can recover errors with deferred `errors.Recover`:

```go
func doesSomething() (any, error) {
    return nil, errors.New("error message")
}

func handleManyThings (err error) {
    defer errors.Recover(&err)

    result := Must(doesSomething())
    r2 := Mustf(doSomethingElse(result))("the reason why it failed")

    ...

    return
}
```

This is useful to avoid many `if err != nil then return` statements.
The `Recover` and `Must` add the stacktrace so you know where the errors come from.
You can also use `Mustf` to format the error message.

#### Blasphemy

```go
errors.Ignore[T any](r T, _ error) T
```
