package obs

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var tp *sdktrace.TracerProvider

func InitTracing(serviceName string) func(context.Context) error {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		otel.SetTracerProvider(sdktrace.NewTracerProvider()) // no-op
		otel.SetTextMapPropagator(propagation.TraceContext{})
		return func(context.Context) error { return nil }
	}
	exp, _ := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	res, _ := sdkresource.New(context.Background(),
		sdkresource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp.Shutdown
}
