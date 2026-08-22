package gotel

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func buildResource(cfg *Config) (*resource.Resource, error) {
	svcName := cfg.ServiceName
	if svcName == "" {
		svcName = "unknown"
	}

	resAttrs := make([]attribute.KeyValue, 0, len(cfg.Attributes)+1)
	resAttrs = append(resAttrs, semconv.ServiceNameKey.String(svcName))
	if cfg.Version != "" {
		resAttrs = append(resAttrs, semconv.ServiceVersion(cfg.Version))
	}
	for key, value := range cfg.Attributes {
		resAttrs = append(resAttrs, attribute.String(key, value))
	}

	defRes := resource.Default()
	return resource.Merge(
		defRes,
		resource.NewWithAttributes(
			defRes.SchemaURL(),
			resAttrs...,
		),
	)
}

func resourceServiceName(cfg *Config) string {
	if cfg.ServiceName != "" {
		return cfg.ServiceName
	}
	return "unknown"
}
