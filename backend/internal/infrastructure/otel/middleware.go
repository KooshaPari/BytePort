// Package otel provides OpenTelemetry instrumentation for the BytePort backend.
package otel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/gin-gonic/gin"
)

// Tracer is the package-level tracer for the BytePort service.
var Tracer trace.Tracer

// InitOpenTelemetry configures the global OTel tracer provider.
// Reads endpoint, service name, and environment from env vars.
// Defaults to http://localhost:4318 if OTLP_ENDPOINT is unset.
// Sets global Tracer and returns a shutdown function.
func InitOpenTelemetry() func() {
	endpoint := os.Getenv("OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4318"
	}

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "byteport-backend"
	}

	env := os.Getenv("BYTEPORT_ENV")
	if env == "" {
		env = "development"
	}

	exp, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Printf("OTel: failed to create exporter: %v (traces disabled)", err)
		return func() {}
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.DeploymentEnvironment(env),
		attribute.String("service.version", os.Getenv("BYTEPORT_VERSION")),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	Tracer = tp.Tracer(serviceName)

	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("OTel: tracer provider shutdown error: %v", err)
		}
	}
}

// HTTPMiddleware returns an HTTP handler that wraps the next handler with
// OTel tracing. It extracts the incoming trace context, creates a span
// named after the HTTP method + path, and records request attributes.
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Tracer == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

		ctx, span := Tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(r.Method),
				semconv.HTTPURL(r.URL.String()),
				semconv.NetHostName(r.Host),
				attribute.String("http.target", r.URL.Path),
				attribute.String("http.user_agent", r.UserAgent()),
			),
		)
		defer span.End()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// GinMiddleware returns a Gin-compatible OTel tracing middleware.
// It wraps each request in an OTel span with HTTP semantic attributes.
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if Tracer == nil {
			c.Next()
			return
		}

		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)

		ctx, span := Tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPURL(c.Request.URL.String()),
				semconv.NetHostName(c.Request.Host),
				attribute.String("http.target", c.Request.URL.Path),
				attribute.String("http.user_agent", c.Request.UserAgent()),
			),
		)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
