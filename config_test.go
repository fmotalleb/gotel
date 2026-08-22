package gotel

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestParseConfig_JSON(t *testing.T) {
	data := []byte(`{
		"service_name": "test-service",
		"version": "1.0.0",
		"attributes": {"env": "test"},
		"tracing": {"url": "http://localhost:4318", "insecure": true},
		"metrics": {"url": "http://localhost:4319"},
		"logs": {"url": "http://localhost:4320"}
	}`)

	cfg, err := ParseConfig(data)
	assert.NoError(t, err)
	assert.Equal(t, "test-service", cfg.ServiceName)
	assert.Equal(t, "1.0.0", cfg.Version)
	assert.Equal(t, "test", cfg.Attributes["env"])
	assert.Equal(t, "http://localhost:4318", cfg.Tracing.URL)
	assert.Equal(t, true, cfg.Tracing.Insecure)
	assert.Equal(t, "http://localhost:4319", cfg.Metrics.URL)
	assert.Equal(t, "http://localhost:4320", cfg.Logs.URL)
}

func TestParseConfig_InvalidJSON(t *testing.T) {
	_, err := ParseConfig([]byte("{invalid}"))
	assert.Error(t, err)
}

func TestParseConfig_Defaults(t *testing.T) {
	data := []byte(`{"service_name": "test"}`)

	cfg, err := ParseConfig(data)
	assert.NoError(t, err)
	assert.Equal(t, "test", cfg.ServiceName)
	assert.NotZero(t, cfg.Tracing)
	assert.NotZero(t, cfg.Metrics)
	assert.NotZero(t, cfg.Logs)
	assert.Equal(t, false, cfg.Tracing.Insecure)
	assert.Equal(t, false, cfg.Metrics.Insecure)
	assert.Equal(t, 60*time.Second, cfg.Metrics.Interval)
	assert.Equal(t, false, cfg.Logs.Insecure)
}

func TestParseConfig_ExplicitInterval(t *testing.T) {
	data := []byte(`{
		"service_name": "test",
		"metrics": {"url": "http://localhost:4318", "interval": 30000000000}
	}`)

	cfg, err := ParseConfig(data)
	assert.NoError(t, err)
	assert.Equal(t, 30*time.Second, cfg.Metrics.Interval)
}

func TestParseConfig_EmptyJSON(t *testing.T) {
	cfg, err := ParseConfig([]byte(`{}`))
	assert.NoError(t, err)
	assert.NotZero(t, cfg.Tracing)
	assert.NotZero(t, cfg.Metrics)
	assert.NotZero(t, cfg.Logs)
}

func TestParseConfig_SignalHeaders(t *testing.T) {
	data := []byte(`{
		"tracing": {"url": "http://localhost:4318", "headers": {"Authorization": "Bearer token"}},
		"metrics": {"url": "http://localhost:4318", "headers": {"X-Custom": "value"}},
		"logs": {"url": "http://localhost:4318", "headers": {"X-Log-Token": "abc"}}
	}`)

	cfg, err := ParseConfig(data)
	assert.NoError(t, err)
	assert.Equal(t, "Bearer token", cfg.Tracing.Headers["Authorization"])
	assert.Equal(t, "value", cfg.Metrics.Headers["X-Custom"])
	assert.Equal(t, "abc", cfg.Logs.Headers["X-Log-Token"])
}

func TestConfig_HasSignal(t *testing.T) {
	cfg := &Config{}
	assert.False(t, cfg.hasSignal())

	cfg.Tracing = &TracingConfig{URL: "http://localhost:4318"}
	assert.True(t, cfg.hasSignal())

	cfg2 := &Config{Metrics: &MetricsConfig{URL: "http://localhost:4318"}}
	assert.True(t, cfg2.hasSignal())

	cfg3 := &Config{Logs: &LogsConfig{URL: "http://localhost:4318"}}
	assert.True(t, cfg3.hasSignal())
}
