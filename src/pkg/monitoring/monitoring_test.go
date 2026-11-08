package monitoring

import (
	"context"
	"errors"
	"testing"

	"github.com/getsentry/sentry-go"
)

func TestStartOperationContextCarriesSpan(t *testing.T) {
	op, ctx := StartOperation(context.Background(), "test", "test.op")
	defer op.Finish()

	if got := sentry.SpanFromContext(ctx); got != op.span {
		t.Fatalf("SpanFromContext = %v, want operation span %v", got, op.span)
	}

	child := sentry.StartSpan(ctx, "child")
	defer child.Finish()
	if child.ParentSpanID != op.span.SpanID {
		t.Errorf("child ParentSpanID = %v, want %v", child.ParentSpanID, op.span.SpanID)
	}
}

func TestMonitorFuncReturnsError(t *testing.T) {
	want := errors.New("boom")
	err := MonitorFunc(context.Background(), "test", "test.op", func(context.Context) error {
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("MonitorFunc error = %v, want %v", err, want)
	}
}
