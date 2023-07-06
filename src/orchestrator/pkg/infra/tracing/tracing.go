// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package tracing

import (
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/utils"
)

type Config struct {
	ZipkinIp   string `json:"zipkin-ip"`
	ZipkinPort string `json:"zipkin-port"`
	JaegerIp   string `json:"jaeger-ip"`
	JaegerPort string `json:"jaeger-port"`
}

func parseConfig(cfgMap map[string]interface{}) (*Config, error) {
	cfg := &Config{}
	return cfg, utils.ConvertType(&cfgMap, cfg)
}

func createTracerProvider(exporters ...tracesdk.SpanExporter) (*tracesdk.TracerProvider, error) {
	var opts []tracesdk.TracerProviderOption
	for _, exp := range exporters {
		opts = append(opts, tracesdk.WithBatcher(exp))
	}
	var ok bool
	var name, namespace string
	if name, ok = os.LookupEnv("APP_NAME"); !ok {
		name = "unknown"
	}
	if namespace, ok = os.LookupEnv("POD_NAMESPACE"); !ok {
		namespace = "default"
	}
	opts = append(opts, tracesdk.WithResource(resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(name+"."+namespace),
	)))

	return tracesdk.NewTracerProvider(opts...), nil
}

func InitializeTracer() error {
	cfgMap, err := utils.StructToMap(config.GetConfiguration())
	if err != nil {
		return err
	}

	cfg, err := parseConfig(cfgMap)
	if err != nil {
		return err
	}

	var traceExporters []traceExporter
	if cfg.JaegerIp != "" {
		traceExporters = append(traceExporters, &jaegerExporter{jaegerIp: cfg.JaegerIp, jaegerPort: cfg.JaegerPort})
	}

	if cfg.ZipkinIp != "" {
		traceExporters = append(traceExporters, &zipkinExporter{zipkinIp: cfg.ZipkinIp, zipkinPort: cfg.ZipkinPort})
	}

	var exporters []tracesdk.SpanExporter
	for _, traceExporter := range traceExporters {
		exp, err := traceExporter.createExporter()
		if err != nil {
			return err
		}

		exporters = append(exporters, exp)
	}

	tp, err := createTracerProvider(exporters...)
	if err != nil {
		return err
	}

	otel.SetTracerProvider(tp)
	// Istio uses b3 propagation
	otel.SetTextMapPropagator(b3.New())
	return nil
}

func Middleware(next http.Handler) http.Handler {
	var ok bool
	var name string
	if name, ok = os.LookupEnv("APP_NAME"); !ok {
		name = "unknown"
	}
	tracer := otel.Tracer(name)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path)
		defer span.End()
		next.ServeHTTP(w, r.Clone(ctx))
	})
}
