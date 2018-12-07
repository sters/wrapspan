package wrapspan

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/mocktracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func TestWrap(t *testing.T) {
	type span struct {
		operationName string
		tags          map[string]interface{}
	}

	tests := []struct {
		name     string
		testcase func(ctx context.Context) error
		wantErr  bool
		wantSpan []span
	}{
		{
			"simple Wrap",
			func(ctx context.Context) error {
				opts := []ddtrace.StartSpanOption{}
				return Wrap(ctx, "test", opts, func(ctx context.Context) error {
					return nil
				})
			},
			false,
			[]span{
				{
					"test",
					map[string]interface{}{
						"resource.name": "test",
					},
				},
			},
		},
		{
			"Wrap in Wrap with tag",
			func(ctx context.Context) error {
				opts := []ddtrace.StartSpanOption{
					tracer.Tag("foo", "bar"),
				}

				return Wrap(ctx, "test-outer", opts, func(ctx context.Context) error {
					opts := []ddtrace.StartSpanOption{
						tracer.Tag("inner", "tag"),
					}

					return Wrap(ctx, "test-inner", opts, func(ctx context.Context) error {
						return nil
					})
				})
			},
			false,
			[]span{
				{
					"test-inner",
					map[string]interface{}{
						"resource.name": "test-inner",
						"inner":         "tag",
					},
				},
				{
					"test-outer",
					map[string]interface{}{
						"resource.name": "test-outer",
						"foo":           "bar",
					},
				},
			},
		},
		{
			"Wrap in Wrap with error",
			func(ctx context.Context) error {
				opts := []ddtrace.StartSpanOption{}
				return Wrap(ctx, "test-outer", opts, func(ctx context.Context) error {
					opts := []ddtrace.StartSpanOption{}
					err := Wrap(ctx, "test-inner", opts, func(ctx context.Context) error {
						return fmt.Errorf("something error")
					})
					return fmt.Errorf("inner is error - %s", err.Error())
				})
			},
			true,
			[]span{
				{
					"test-inner",
					map[string]interface{}{
						"resource.name": "test-inner",
						"error":         fmt.Errorf("something error"),
					},
				},
				{
					"test-outer",
					map[string]interface{}{
						"resource.name": "test-outer",
						"error":         fmt.Errorf("inner is error - something error"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tracer := mocktracer.Start()

			if err := tt.testcase(ctx); (err != nil) != tt.wantErr {
				t.Fatalf("returns = %+v, wantErr = %+v", err, tt.wantErr)
			}

			tracer.Stop()

			spans := tracer.FinishedSpans()

			if len(spans) != len(tt.wantSpan) {
				t.Fatalf("len(spans) = %d, len(wantSpan) = %d", len(spans), len(tt.wantSpan))
			}

			for i, span := range spans {
				wantSpan := tt.wantSpan[i]

				if operationName := span.OperationName(); operationName != wantSpan.operationName {
					t.Fatalf("invalid operation name, got = %+v, want = %+v", operationName, wantSpan.operationName)
				}

				for wantTagKey, wantTagValue := range wantSpan.tags {
					if tag := span.Tag(wantTagKey); !reflect.DeepEqual(tag, wantTagValue) {
						t.Fatalf("invalid tag, name = %s, got = %+v, want = %+v", wantTagKey, tag, wantTagValue)
					}
				}
			}
		})
	}
}
