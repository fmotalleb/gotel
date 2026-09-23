package gotel

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fmotalleb/go-tools/defaulter"
)

type TracingConfig struct {
	URL      string            `json:"url"      yaml:"url"      mapstructure:"url"      env:"GOTEL_TRACING_URL"`
	Insecure bool              `json:"insecure" yaml:"insecure" mapstructure:"insecure" env:"GOTEL_TRACING_INSECURE"`
	Headers  map[string]string `json:"headers"  yaml:"headers"  mapstructure:"headers"  env:"GOTEL_TRACING_HEADERS"`
}

type MetricsConfig struct {
	URL      string            `json:"url"      yaml:"url"      mapstructure:"url"      env:"GOTEL_METRICS_URL"`
	Insecure bool              `json:"insecure" yaml:"insecure" mapstructure:"insecure" env:"GOTEL_METRICS_INSECURE"`
	Headers  map[string]string `json:"headers"  yaml:"headers"  mapstructure:"headers"  env:"GOTEL_METRICS_HEADERS"`
	Interval time.Duration     `json:"interval" yaml:"interval" mapstructure:"interval" env:"GOTEL_METRICS_INTERVAL"`
}

type LogsConfig struct {
	URL      string            `json:"url"      yaml:"url"      mapstructure:"url"      env:"GOTEL_LOGS_URL"`
	Insecure bool              `json:"insecure" yaml:"insecure" mapstructure:"insecure" env:"GOTEL_LOGS_INSECURE"`
	Headers  map[string]string `json:"headers"  yaml:"headers"  mapstructure:"headers"  env:"GOTEL_LOGS_HEADERS"`
}

type Config struct {
	ServiceName string            `json:"service_name" yaml:"service_name" mapstructure:"service_name" env:"GOTEL_SERVICE_NAME"`
	Version     string            `json:"version"      yaml:"version"      mapstructure:"version"     env:"GOTEL_VERSION"`
	Attributes  map[string]string `json:"attributes"   yaml:"attributes"   mapstructure:"attributes"  env:"GOTEL_ATTRIBUTES"`
	Tracing     *TracingConfig    `json:"tracing"      yaml:"tracing"      mapstructure:"tracing"     env:"GOTEL_TRACING"`
	Metrics     *MetricsConfig    `json:"metrics"      yaml:"metrics"      mapstructure:"metrics"     env:"GOTEL_METRICS"`
	Logs        *LogsConfig       `json:"logs"         yaml:"logs"         mapstructure:"logs"        env:"GOTEL_LOGS"`
}

func ParseConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.ApplyDefaults()
	return &cfg, nil
}

func (c *Config) ApplyDefaults() {
	if c.Tracing == nil {
		c.Tracing = &TracingConfig{}
	}
	if c.Metrics == nil {
		c.Metrics = &MetricsConfig{}
	}
	if c.Logs == nil {
		c.Logs = &LogsConfig{}
	}
	_ = defaulter.ApplyDefaults(c, nil)
}

func (c *Config) hasSignal() bool {
	return (c.Tracing != nil && c.Tracing.URL != "") ||
		(c.Metrics != nil && c.Metrics.URL != "") ||
		(c.Logs != nil && c.Logs.URL != "")
}
