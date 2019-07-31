package opencensusexporter

import (
	"time"

	"github.com/open-telemetry/opentelemetry-service/exporter/opencensusexporter"
)

// Config defines configuration for OpenCensus exporter.
type Config struct {
	opencensusexporter.Config `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.

	UseUnaryExporter     bool          `mapstructure:"unary-exporter,omitempty"`
	UnaryExporterTimeout time.Duration `mapstructure:"unary-exporter-timeout,omitempty"`
}
