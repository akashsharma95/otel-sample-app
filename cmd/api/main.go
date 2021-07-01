package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"notesapp/internal/controllers"
	"notesapp/pkg/metrics"
	"notesapp/pkg/tracing"
)

func main() {
	r := gin.Default()
	gin.SetMode("debug")

	ctx := context.Background()

	tp, err := tracing.GetTracerProvider(ctx)
	if err != nil {
		log.Fatal(err)
	}

	ctrl := metrics.GetMetricsController(ctx)
	if err := ctrl.Start(ctx); err != nil {
		log.Fatalf("could not start metric controller: %v", err)
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(jaeger.Jaeger{}, propagation.Baggage{}))

	r.Use(otelgin.Middleware("notes"), metrics.HttpServerRequestMetrics())

	r.GET("/notes/:id", controllers.GetNote)
	r.GET("/notes", controllers.GetNotes)
	r.POST("/notes", controllers.CreateNote)
	r.PUT("/notes/:id", controllers.UpdateNote)
	r.DELETE("/notes/:id", controllers.DeleteNote)

	err = r.Run(":8081")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := ctrl.Stop(ctx); err != nil {
			otel.Handle(err)
		}
	}()

	err = tp.Shutdown(ctx)
	if err != nil {
		log.Fatalf("failed to shutdown with err %v", err)
	}
}
