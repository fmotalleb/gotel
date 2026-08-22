package gotel

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

type transportKind int

const (
	transportHTTP transportKind = iota
	transportGRPC
)

type signalEndpoint struct {
	transport  transportKind
	endpoint   string
	path       string
	tls        bool
	skipVerify bool
}

func parseEndpoint(rawURL string, insecure bool) (signalEndpoint, error) {
	ep := signalEndpoint{transport: transportHTTP, skipVerify: insecure}
	if !strings.Contains(rawURL, "://") {
		ep.endpoint = rawURL
		return ep, nil
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return signalEndpoint{}, fmt.Errorf("parse OTLP endpoint %q: %w", rawURL, err)
	}
	switch strings.ToLower(u.Scheme) {
	case "http":
		ep.transport, ep.tls = transportHTTP, false
	case "https":
		ep.transport, ep.tls = transportHTTP, true
	case "grpc":
		ep.transport, ep.tls = transportGRPC, false
	case "grpcs":
		ep.transport, ep.tls = transportGRPC, true
	default:
		return signalEndpoint{}, fmt.Errorf("unsupported OTLP URL scheme %q (use http, https, grpc or grpcs)", u.Scheme)
	}
	if u.Host != "" {
		ep.endpoint = u.Host
	} else {
		ep.endpoint = strings.TrimPrefix(rawURL, u.Scheme+":")
	}
	ep.path = u.Path
	return ep, nil
}

func skipVerifyTLS() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // explicit user opt-in
		MinVersion:         tls.VersionTLS12,
	}
}

func pathOrDefault(path, def string) string {
	if path == "" {
		return def
	}
	return path
}

func newTraceExporter(ctx context.Context, ep signalEndpoint, headers map[string]string) (sdktrace.SpanExporter, error) {
	if ep.transport == transportGRPC {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(ep.endpoint),
		}
		if len(headers) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(headers))
		}
		switch {
		case !ep.tls:
			opts = append(opts, otlptracegrpc.WithInsecure())
		case ep.skipVerify:
			opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(skipVerifyTLS())))
		default:
			opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(nil)))
		}
		return otlptracegrpc.New(ctx, opts...)
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(ep.endpoint),
		otlptracehttp.WithURLPath(pathOrDefault(ep.path, "/v1/traces")),
	}
	if len(headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}
	switch {
	case !ep.tls:
		opts = append(opts, otlptracehttp.WithInsecure())
	case ep.skipVerify:
		opts = append(opts, otlptracehttp.WithTLSClientConfig(skipVerifyTLS()))
	}
	return otlptracehttp.New(ctx, opts...)
}

func newMetricExporter(ctx context.Context, ep signalEndpoint, headers map[string]string) (metric.Exporter, error) {
	if ep.transport == transportGRPC {
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(ep.endpoint),
		}
		if len(headers) > 0 {
			opts = append(opts, otlpmetricgrpc.WithHeaders(headers))
		}
		switch {
		case !ep.tls:
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		case ep.skipVerify:
			opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(skipVerifyTLS())))
		default:
			opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(nil)))
		}
		return otlpmetricgrpc.New(ctx, opts...)
	}

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(ep.endpoint),
		otlpmetrichttp.WithURLPath(pathOrDefault(ep.path, "/v1/metrics")),
	}
	if len(headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(headers))
	}
	switch {
	case !ep.tls:
		opts = append(opts, otlpmetrichttp.WithInsecure())
	case ep.skipVerify:
		opts = append(opts, otlpmetrichttp.WithTLSClientConfig(skipVerifyTLS()))
	}
	return otlpmetrichttp.New(ctx, opts...)
}

func newLogExporter(ctx context.Context, ep signalEndpoint, headers map[string]string) (log.Exporter, error) {
	if ep.transport == transportGRPC {
		opts := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(ep.endpoint),
		}
		if len(headers) > 0 {
			opts = append(opts, otlploggrpc.WithHeaders(headers))
		}
		switch {
		case !ep.tls:
			opts = append(opts, otlploggrpc.WithInsecure())
		case ep.skipVerify:
			opts = append(opts, otlploggrpc.WithTLSCredentials(credentials.NewTLS(skipVerifyTLS())))
		default:
			opts = append(opts, otlploggrpc.WithTLSCredentials(credentials.NewTLS(nil)))
		}
		return otlploggrpc.New(ctx, opts...)
	}

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(ep.endpoint),
		otlploghttp.WithURLPath(pathOrDefault(ep.path, "/v1/logs")),
	}
	if len(headers) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(headers))
	}
	switch {
	case !ep.tls:
		opts = append(opts, otlploghttp.WithInsecure())
	case ep.skipVerify:
		opts = append(opts, otlploghttp.WithTLSClientConfig(skipVerifyTLS()))
	}
	return otlploghttp.New(ctx, opts...)
}
