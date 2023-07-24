package tracing

import (
	"fmt"
	"net"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/zipkin"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

const (
	defaultJaegerPort = "14268"
	defaultZipkinPort = "9411"
)

type traceExporter interface {
	createExporter() (tracesdk.SpanExporter, error)
}

type jaegerExporter struct {
	jaegerIp   string
	jaegerPort string
}

func (j *jaegerExporter) createExporter() (tracesdk.SpanExporter, error) {
	jaegerPort := j.jaegerPort
	if jaegerPort == "" {
		jaegerPort = defaultJaegerPort
	}

	endpoint := fmt.Sprintf("http://%s/api/traces", net.JoinHostPort(j.jaegerIp, jaegerPort))
	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
}

type zipkinExporter struct {
	zipkinIp   string
	zipkinPort string
}

func (z *zipkinExporter) createExporter() (tracesdk.SpanExporter, error) {
	zipkinPort := z.zipkinPort
	if zipkinPort == "" {
		zipkinPort = defaultZipkinPort
	}

	endpoint := fmt.Sprintf("http://%s/api/v2/spans", net.JoinHostPort(z.zipkinIp, zipkinPort))
	return zipkin.New(endpoint)
}
