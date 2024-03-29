// Copyright 2019 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opencensusexporter

import (
	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/open-telemetry/opentelemetry-service/config/configmodels"
	"github.com/open-telemetry/opentelemetry-service/consumer"
	"github.com/open-telemetry/opentelemetry-service/exporter"
	"github.com/open-telemetry/opentelemetry-service/exporter/opencensusexporter"
	"go.uber.org/zap"
)

const (
	// The value of "type" key in configuration.
	typeStr = "opencensus"
)

// Factory is the factory for OpenCensus exporter.
type Factory struct {
	factory opencensusexporter.Factory
}

// Type gets the type of the Exporter config created by this factory.
func (f *Factory) Type() string {
	return typeStr
}

// CreateDefaultConfig creates the default configuration for exporter.
func (f *Factory) CreateDefaultConfig() configmodels.Exporter {
	cfg := f.factory.CreateDefaultConfig()
	c := cfg.(*opencensusexporter.Config)
	return &Config{Config: *c, UseUnaryExporter: true}
}

// CreateTraceExporter creates a trace exporter based on this config.
func (f *Factory) CreateTraceExporter(logger *zap.Logger, config configmodels.Exporter) (consumer.TraceConsumer, exporter.StopFunc, error) {
	ocac := config.(*Config)
	opts, err := f.factory.OCAgentOptions(logger, &ocac.Config)
	if err != nil {
		return nil, nil, err
	}

	if ocac.UseUnaryExporter {
		opts = append(opts, ocagent.WithUnaryBatchExporter(ocagent.UnaryExporterParams{
			Timeout: ocac.UnaryExporterTimeout,
		}))
	}
	return f.factory.CreateOCAgent(logger, &ocac.Config, opts)
}

// CreateMetricsExporter creates a metrics exporter based on this config.
func (f *Factory) CreateMetricsExporter(logger *zap.Logger, config configmodels.Exporter) (consumer.MetricsConsumer, exporter.StopFunc, error) {
	return f.factory.CreateMetricsExporter(logger, config)
}
