package providers

import (
	"context"

	"github.com/maisiq/go-auth-service/cmd/wire"
	"github.com/maisiq/go-auth-service/internal/configs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.19.0"
	"go.opentelemetry.io/otel/trace"
)

// Trace

func JaegerExporterProvider(c *wire.DIContainer) *otlptrace.Exporter {
	ctx := context.TODO()

	cfg := wire.Get[*configs.Config](c)
	_ = cfg

	jaegerClient := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint("jaeger:4318"),
		otlptracehttp.WithInsecure(),
	)

	jaegerExporter, err := otlptrace.New(ctx, jaegerClient)
	if err != nil {
		panic(err)
	}

	return jaegerExporter
}

func TracerProvider(c *wire.DIContainer) trace.TracerProvider {
	exporter := wire.Get[*otlptrace.Exporter](c)

	ctx := context.TODO()

	res, _ := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("go-auth-service"),
		),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	return tp
}

func UserServiceTracerProvider(c *wire.DIContainer) trace.Tracer {
	provider := wire.Get[trace.TracerProvider](c)
	return provider.Tracer("user-service")
}

func HTTPTracerProvider(c *wire.DIContainer) trace.Tracer {
	provider := wire.Get[trace.TracerProvider](c)
	return provider.Tracer("http-server")
}
