package gotel

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.opentelemetry.io/otel/propagation"
)

func TestParseEndpoint_SchemeSelectsTransportAndTLS(t *testing.T) {
	cases := []struct {
		name      string
		url       string
		transport transportKind
		endpoint  string
		path      string
		tls       bool
	}{
		{name: "http plaintext", url: "http://collector:4318", transport: transportHTTP, endpoint: "collector:4318", path: "", tls: false},
		{name: "https tls", url: "https://collector:4318/v1/traces", transport: transportHTTP, endpoint: "collector:4318", path: "/v1/traces", tls: true},
		{name: "grpc plaintext", url: "grpc://collector:4317", transport: transportGRPC, endpoint: "collector:4317", path: "", tls: false},
		{name: "grpcs tls", url: "grpcs://collector:4317", transport: transportGRPC, endpoint: "collector:4317", path: "", tls: true},
		{name: "bare host defaults to http plaintext", url: "collector:4318", transport: transportHTTP, endpoint: "collector:4318", path: "", tls: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ep, err := parseEndpoint(tc.url, false)
			assert.NoError(t, err)
			assert.Equal(t, tc.transport, ep.transport)
			assert.Equal(t, tc.endpoint, ep.endpoint)
			assert.Equal(t, tc.path, ep.path)
			assert.Equal(t, tc.tls, ep.tls)
		})
	}
}

func TestParseEndpoint_UnsupportedScheme(t *testing.T) {
	_, err := parseEndpoint("udp://collector:4318", false)
	assert.Error(t, err)
}

func TestParseEndpoint_InsecureSkipsVerify(t *testing.T) {
	ep, err := parseEndpoint("https://collector:4318", true)
	assert.NoError(t, err)
	assert.True(t, ep.tls)
	assert.True(t, ep.skipVerify)
}

func TestSetup_WithNoSignals(t *testing.T) {
	result, err := Setup(t.Context())
	assert.NoError(t, err)
	assert.NotZero(t, result.Shutdown)
	assert.Zero(t, result.LogCore)
}

func TestSetup_WithEmptyConfig(t *testing.T) {
	result, err := Setup(t.Context(), WithConfig(&Config{}))
	assert.NoError(t, err)
	assert.NotZero(t, result.Shutdown)
	assert.Zero(t, result.LogCore)
}

func TestOptions_WithServiceName(t *testing.T) {
	cfg := &Config{}
	opts := []Option{WithConfig(cfg), WithServiceName("test-service")}
	var parsedOpts []options
	for _, opt := range opts {
		o := options{}
		opt(&o)
		parsedOpts = append(parsedOpts, o)
	}
	result := buildConfig(parsedOpts)
	assert.Equal(t, "test-service", result.ServiceName)
}

func TestOptions_WithVersion(t *testing.T) {
	cfg := &Config{}
	opts := []Option{WithConfig(cfg), WithVersion("1.0.0")}
	var parsedOpts []options
	for _, opt := range opts {
		o := options{}
		opt(&o)
		parsedOpts = append(parsedOpts, o)
	}
	result := buildConfig(parsedOpts)
	assert.Equal(t, "1.0.0", result.Version)
}

func TestOptions_WithAttributes(t *testing.T) {
	cfg := &Config{}
	opts := []Option{WithConfig(cfg), WithAttributes(map[string]string{"env": "test"})}
	var parsedOpts []options
	for _, opt := range opts {
		o := options{}
		opt(&o)
		parsedOpts = append(parsedOpts, o)
	}
	result := buildConfig(parsedOpts)
	assert.Equal(t, "test", result.Attributes["env"])
}

func TestOptions_WithTracing(t *testing.T) {
	opts := []Option{WithTracing(&TracingConfig{URL: "http://localhost:4318"})}
	var parsedOpts []options
	for _, opt := range opts {
		o := options{}
		opt(&o)
		parsedOpts = append(parsedOpts, o)
	}
	result := buildConfig(parsedOpts)
	assert.NotZero(t, result.Tracing)
	assert.Equal(t, "http://localhost:4318", result.Tracing.URL)
}

func TestOptions_WithMetrics(t *testing.T) {
	opts := []Option{WithMetrics(&MetricsConfig{URL: "http://localhost:4318"})}
	var parsedOpts []options
	for _, opt := range opts {
		o := options{}
		opt(&o)
		parsedOpts = append(parsedOpts, o)
	}
	result := buildConfig(parsedOpts)
	assert.NotZero(t, result.Metrics)
	assert.Equal(t, "http://localhost:4318", result.Metrics.URL)
}

func TestOptions_WithLogs(t *testing.T) {
	opts := []Option{WithLogs(&LogsConfig{URL: "http://localhost:4318"})}
	var parsedOpts []options
	for _, opt := range opts {
		o := options{}
		opt(&o)
		parsedOpts = append(parsedOpts, o)
	}
	result := buildConfig(parsedOpts)
	assert.NotZero(t, result.Logs)
	assert.Equal(t, "http://localhost:4318", result.Logs.URL)
}

func TestOptions_WithPropagator(t *testing.T) {
	p := propagation.TraceContext{}
	opts := []Option{WithPropagator(p)}
	var parsedOpts []options
	for _, opt := range opts {
		o := options{}
		opt(&o)
		parsedOpts = append(parsedOpts, o)
	}
	assert.NotZero(t, parsedOpts[0].propagator)
}
