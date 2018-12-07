package wrapspan

import (
	"context"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Wrap function with datadog trace span
func Wrap(ctx context.Context, operationName string, opts []tracer.StartSpanOption, f func(ctx context.Context) error) (err error) {
	span, ctx := tracer.StartSpanFromContext(
		ctx,
		operationName,
		opts...,
	)
	defer func() { span.Finish(tracer.WithError(err)) }()

	err = f(ctx)
	return
}
