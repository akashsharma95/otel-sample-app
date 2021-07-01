package metrics

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
)

func GetMetricsController(ctx context.Context) *controller.Controller {
	client := otlpmetricgrpc.NewClient(otlpmetricgrpc.WithInsecure(), otlpmetricgrpc.WithDialOption(grpc.WithBlock()))
	exp, err := otlpmetric.New(ctx, client)
	if err != nil {
		log.Fatalf("failed to create the collector exporter: %v", err)
	}

	res, err := resource.New(
		context.Background(),
		resource.WithBuiltinDetectors(),
	)

	ctrl := controller.New(
		processor.New(
			selector.NewWithExactDistribution(),
			exp,
			processor.WithMemory(true),
		),
		controller.WithExporter(exp),
		controller.WithCollectPeriod(1*time.Second),
		controller.WithResource(res),
	)
	global.SetMeterProvider(ctrl.MeterProvider())

	return ctrl
}
