package gotel

import (
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

type Option func(*options)

type options struct {
	config      *Config
	logger      *zap.Logger
	serviceName string
	version     string
	attributes  map[string]string
	propagator  propagation.TextMapPropagator
}

func WithConfig(cfg *Config) Option {
	return func(o *options) {
		o.config = cfg
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func WithServiceName(name string) Option {
	return func(o *options) {
		o.serviceName = name
	}
}

func WithVersion(version string) Option {
	return func(o *options) {
		o.version = version
	}
}

func WithAttributes(attrs map[string]string) Option {
	return func(o *options) {
		o.attributes = attrs
	}
}

func WithTracing(cfg *TracingConfig) Option {
	return func(o *options) {
		if o.config == nil {
			o.config = &Config{}
		}
		o.config.Tracing = cfg
	}
}

func WithMetrics(cfg *MetricsConfig) Option {
	return func(o *options) {
		if o.config == nil {
			o.config = &Config{}
		}
		o.config.Metrics = cfg
	}
}

func WithLogs(cfg *LogsConfig) Option {
	return func(o *options) {
		if o.config == nil {
			o.config = &Config{}
		}
		o.config.Logs = cfg
	}
}

func WithPropagator(p propagation.TextMapPropagator) Option {
	return func(o *options) {
		o.propagator = p
	}
}

func buildConfig(opts []options) *Config {
	cfg := &Config{}
	if len(opts) > 0 && opts[0].config != nil {
		cfg = opts[0].config
	}

	for _, o := range opts {
		if o.serviceName != "" {
			cfg.ServiceName = o.serviceName
		}
		if o.version != "" {
			cfg.Version = o.version
		}
		if o.attributes != nil {
			if cfg.Attributes == nil {
				cfg.Attributes = make(map[string]string)
			}
			for k, v := range o.attributes {
				cfg.Attributes[k] = v
			}
		}
	}
	return cfg
}
