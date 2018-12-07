# wrapspan
for https://github.com/DataDog/dd-trace-go

Want to know more application tracing, can use this.
```go
ctx := context.Background()

err := Wrap(ctx, "span-somethings", nil, func(ctx context.Context) error {
    if err := Wrap(ctx, "span-1", nil, func(ctx context.Context) (err error) {
        err = X.Somethings1(ctx)
        return
    }); err != nil {
        return err
    }

    if err := Wrap(ctx, "span-2", nil, func(ctx context.Context) (err error) {
        err = X.Somethings2(ctx)
        return
    }); err != nil {
        return err
    }

    return nil
})
```

then DataDog trace shows...
```
|----------------- span-somethings -----------------|
 |----- span-1 -----| |---------- span-2 ----------|
```
