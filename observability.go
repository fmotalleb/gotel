package gotel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ShutdownFunc func(context.Context) error

type SetupResult struct {
	Shutdown ShutdownFunc
	LogCore  zapcore.Core
}

func Setup(ctx context.Context, opts ...Option) (SetupResult, error) {
	var parsedOpts []options
	for _, opt := range opts {
		o := options{}
		opt(&o)
		parsedOpts = append(parsedOpts, o)
	}

	cfg := buildConfig(parsedOpts)

	var logger *zap.Logger
	for _, o := range parsedOpts {
		if o.logger != nil {
			logger = o.logger
			break
		}
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	if cfg == nil || !cfg.hasSignal() {
		return SetupResult{Shutdown: noopShutdown}, nil
	}

	res, err := buildResource(cfg)
	if err != nil {
		return SetupResult{Shutdown: noopShutdown}, fmt.Errorf("create otel resource: %w", err)
	}

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	for _, o := range parsedOpts {
		if o.propagator != nil {
			propagator = o.propagator
			break
		}
	}
	otel.SetTextMapPropagator(propagator)

	var (
		shutdownFuncs []ShutdownFunc
		logCore       zapcore.Core
	)

	if cfg.Tracing != nil && cfg.Tracing.URL != "" {
		sd, setupErr := setupTracing(ctx, cfg.Tracing, res, logger)
		if setupErr != nil {
			logger.Warn("tracing setup failed, tracing disabled", zap.Error(setupErr))
		} else {
			shutdownFuncs = append(shutdownFuncs, sd)
		}
	}

	if cfg.Metrics != nil && cfg.Metrics.URL != "" {
		sd, setupErr := setupMetrics(ctx, cfg.Metrics, res, logger)
		if setupErr != nil {
			logger.Warn("metrics setup failed, metrics disabled", zap.Error(setupErr))
		} else {
			shutdownFuncs = append(shutdownFuncs, sd)
		}
	}

	if cfg.Logs != nil && cfg.Logs.URL != "" {
		sd, core, setupErr := setupLogs(ctx, cfg.Logs, res, resourceServiceName(cfg), logger)
		if setupErr != nil {
			logger.Warn("log export setup failed, OTLP logging disabled", zap.Error(setupErr))
		} else {
			shutdownFuncs = append(shutdownFuncs, sd)
			logCore = core
		}
	}

	return SetupResult{
		Shutdown: combinedShutdown(shutdownFuncs),
		LogCore:  logCore,
	}, nil
}

func combinedShutdown(funcs []ShutdownFunc) ShutdownFunc {
	return func(ctx context.Context) error {
		var firstErr error
		for _, fn := range funcs {
			if err := fn(ctx); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}
}

func noopShutdown(_ context.Context) error { return nil }

func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("gotel").Start(ctx, name, opts...)
}

func setupTracing(ctx context.Context, cfg *TracingConfig, res *resource.Resource, logger *zap.Logger) (ShutdownFunc, error) {
	ep, err := parseEndpoint(cfg.URL, cfg.Insecure)
	if err != nil {
		return nil, err
	}
	exp, err := newTraceExporter(ctx, ep, cfg.Headers)
	if err != nil {
		return nil, fmt.Errorf("trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	logger.Info("OTLP tracing enabled", zap.String("url", cfg.URL))
	return tp.Shutdown, nil
}

func setupMetrics(ctx context.Context, cfg *MetricsConfig, res *resource.Resource, logger *zap.Logger) (ShutdownFunc, error) {
	ep, err := parseEndpoint(cfg.URL, cfg.Insecure)
	if err != nil {
		return nil, err
	}
	exp, err := newMetricExporter(ctx, ep, cfg.Headers)
	if err != nil {
		return nil, fmt.Errorf("metric exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(cfg.Interval))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	logger.Info("OTLP metrics enabled", zap.String("url", cfg.URL))
	return mp.Shutdown, nil
}

func setupLogs(ctx context.Context, cfg *LogsConfig, res *resource.Resource, svcName string, logger *zap.Logger) (ShutdownFunc, zapcore.Core, error) {
	ep, err := parseEndpoint(cfg.URL, cfg.Insecure)
	if err != nil {
		return nil, nil, err
	}
	exp, err := newLogExporter(ctx, ep, cfg.Headers)
	if err != nil {
		return nil, nil, fmt.Errorf("log exporter: %w", err)
	}

	lp := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exp)),
		log.WithResource(res),
	)

	otelzapCore := otelzap.NewCore(svcName,
		otelzap.WithLoggerProvider(lp),
	)
	logger.Info("OTLP logging enabled", zap.String("url", cfg.URL))
	return lp.Shutdown, otelzapCore, nil
}
