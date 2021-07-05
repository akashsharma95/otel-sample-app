package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
)

// HttpServerRequestMetrics - Middleware to push http R.E.D metrics from http server
func HttpServerRequestMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer InstrumentHttpRequest(c, start)

		c.Next()
	}
}

func InstrumentHttpRequest(c *gin.Context, start time.Time) {
	// labels extracted from request
	status := strconv.Itoa(c.Writer.Status())
	method := c.Request.Method
	route := c.FullPath()
	elapsed := float64(time.Since(start).Milliseconds())

	meter := global.Meter("notes")

	counter := metric.Must(meter).NewInt64Counter("random_counter", metric.WithDescription("random counter"))

	reqHist := metric.Must(meter).NewFloat64ValueRecorder(
		"request_histogram",
		metric.WithUnit(unit.Milliseconds),
		metric.WithDescription("HTTP server request duration histogram"),
	)

	reqHist.Record(c, elapsed, attribute.String("http_method", method),
		attribute.String("route", route),
		attribute.String("http_status_code", status))

	counter.Add(c, 1, attribute.String("key", method))

	// meter.RecordBatch(c, attrs, reqHist.Measurement(elapsed))
}
