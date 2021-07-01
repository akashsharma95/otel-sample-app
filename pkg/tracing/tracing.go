package tracing

import (
	"context"
	"errors"
	"os"

	"go.opentelemetry.io/collector/translator/conventions"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	jaegerThriftEp = "http://localhost:14268/api/traces"

	service     = "notesapp"
	environment = "production"
	commitID    = "bad49c714a62da5493f2d1d9bafd7ebe8c8ce7eb"
)

// GetTracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func GetTracerProvider(ctx context.Context) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerThriftEp)))
	if err != nil {
		return nil, err
	}

	host, _ := os.Hostname()

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),

		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(service),
			attribute.String(conventions.AttributeHostName, host),
			attribute.String(conventions.AttributeDeploymentEnvironment, environment),
			attribute.String(conventions.AttributeHostName, host),
			attribute.String(conventions.AttributeServiceVersion, commitID),
		)),
	)
	return tp, nil
}

// GetSpan returns current span from context
func GetSpan(ctx context.Context) trace.Span {
	span := trace.SpanFromContext(ctx)
	return span
}

// AddAttribute adds attribute to span
func AddAttribute(ctx context.Context, kvPair ...string) error {
	if len(kvPair)%2 != 0 {
		return errors.New("all keys should have a value")
	}

	span := trace.SpanFromContext(ctx)
	for idx := 0; idx < len(kvPair)-1; idx++ {
		span.SetAttributes(attribute.String(kvPair[idx], kvPair[idx+1]))
	}

	return nil
}

// AddEvent adds event to span
func AddEvent(ctx context.Context, event string, attributes []attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(event, trace.WithAttributes(attributes...))
}

// RecordError sets error
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, "internal error")
}

// CreateChildSpan creates a child of parent span
func CreateChildSpan(ctx context.Context, name string, tracer trace.Tracer, opts ...trace.SpanStartOption) (trace.Span, context.Context) {
	ctx, childSpan := tracer.Start(ctx, name, opts...)
	return childSpan, ctx
}
