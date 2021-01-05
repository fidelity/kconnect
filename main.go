/*
Copyright 2020 The kconnect Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"

	"github.com/fidelity/kconnect/internal/commands"
	intver "github.com/fidelity/kconnect/internal/version"
	"github.com/fidelity/kconnect/pkg/flags"
	"github.com/fidelity/kconnect/pkg/logging"
	_ "github.com/fidelity/kconnect/pkg/plugins" // Import all the plugins
)

func main() {
	if err := setupLogging(); err != nil {
		log.Fatalf("failed to configure logging %v", err)
	}
	traceShutdown, err := setupTracing()
	if err != nil {
		log.Fatalf("failed to setup tracing %v", err)
	}
	defer traceShutdown()

	v := intver.Get()

	commonLabels := []label.KeyValue{
		label.String("version", v.Version),
		label.String("platform", v.Platform),
	}

	tracer := otel.Tracer("kconnect")
	ctx, span := tracer.Start(context.Background(),
		"operation",
		trace.WithAttributes(commonLabels...))
	defer span.End()

	span.AddEvent("kconnect started")

	zap.S().Infow("kconnect - the Kubernetes Connection Manager CLI", "version", v.Version)
	zap.S().Debugw("build information", "date", v.BuildDate, "commit", v.CommitHash, "gover", v.GoVersion)

	rootCmd, err := commands.RootCmd()
	if err != nil {
		zap.S().Fatalw("failed getting root command", "error", err.Error())
	}

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		zap.S().Fatalw("failed executing root command", "error", err.Error())
	}
}

func setupLogging() error {
	verbosity, err := flags.GetFlagValueDirect(os.Args, "verbosity", "v")
	if err != nil {
		return fmt.Errorf("getting verbosity flag: %w", err)
	}

	logVerbosity := 0
	if verbosity != "" {
		logVerbosity, err = strconv.Atoi(verbosity)
		if err != nil {
			return fmt.Errorf("parsing verbosity level: %w", err)
		}
	}

	if err := logging.Configure(logVerbosity); err != nil {
		log.Fatalf("failed to configure logging %v", err)
	}

	return nil
}

func setupTracing() (func(), error) {
	ctx := context.Background()

	exporter, err := stdout.NewExporter([]stdout.Option{
		//stdout.WithQuantiles([]float64{0.5, 0.9, 0.99}),
		stdout.WithWriter(os.Stderr),
		stdout.WithPrettyPrint(),
	}...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize stdout export pipeline: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("kconnect"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resoure: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	pusher := push.New(
		basic.New(
			simple.NewWithExactDistribution(),
			exporter,
		),
		exporter,
		push.WithPeriod(1*time.Second),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(pusher.MeterProvider())
	pusher.Start()

	return func() {
		handleErr(tp.Shutdown(ctx), "failed to shutdown trace provider")
		handleErr(exporter.Shutdown(ctx), "failed to stop exporter")
		pusher.Stop()
	}, nil
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
