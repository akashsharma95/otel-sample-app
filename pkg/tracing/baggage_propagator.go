package tracing

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
)

type BaggagePropagator struct {
}

var _ propagation.TextMapPropagator = BaggagePropagator{}

// TODO: Implement jaeger baggage propagation

// Inject sets baggage key-values from ctx into the carrier.
func (b BaggagePropagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {

}

// Extract returns a copy of parent with the baggage from the carrier added.
func (b BaggagePropagator) Extract(parent context.Context, carrier propagation.TextMapCarrier) context.Context {
	return parent
}

// Fields returns the keys who's values are set with Inject.
func (b BaggagePropagator) Fields() []string {
	return []string{}
}
