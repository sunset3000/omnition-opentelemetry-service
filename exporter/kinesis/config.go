package kinesis

import (
	"github.com/open-telemetry/opentelemetry-service/config/configmodels"
)

// AWSConfig contains AWS specific configuration such as kinesis stream, region, etc.
type AWSConfig struct {
	StreamName      string `mapstructure:"stream-name,omitempty"`
	KinesisEndpoint string `mapstructure:"kinesis-endpoint,omitempty"`
	Region          string `mapstructure:"region,omitempty"`
	Role            string `mapstructure:"role,omitempty"`
}

// KPLConfig contains kinesis producer library related config to controls things
// like aggregation, batching, connections, retries, etc.
type KPLConfig struct {
	AggregateBatchCount  int `mapstructure:"aggregate-batch-count,omitempty"`
	AggregateBatchSize   int `mapstructure:"aggregate-batch-size,omitempty"`
	BatchSize            int `mapstructure:"batch-size, omitempty"`
	BatchCount           int `mapstructure:"batch-count,omitempty"`
	BacklogCount         int `mapstructure:"backlog-count,omitempty"`
	FlushIntervalSeconds int `mapstructure:"flush-interval-seconds,omitempty"`
	MaxConnections       int `mapstructure:"max-connections,omitempty"`
	MaxRetries           int `mapstructure:"max-retries,omitempty"`
	MaxBackoffSeconds    int `mapstructure:"max-backoff-seconds,omitempty"`
}

// Config contains the main configuration options for the kinesis exporter
type Config struct {
	configmodels.ExporterSettings `mapstructure:",squash"`

	AWS AWSConfig `mapstructure:"aws,omitempty"`
	KPL KPLConfig `mapstructure:"kpl,omitempty"`

	QueueSize            int `mapstructure:"queue-size,omitempty"`
	NumWorkers           int `mapstructure:"num-workers,omitempty"`
	MaxBytesPerBatch     int `mapstructure:"max-bytes-per-batch,omitempty"`
	MaxBytesPerSpan      int `mapstructure:"max-bytes-per-span,omitempty"`
	FlushIntervalSeconds int `mapstructure:"flush-interval-seconds,omitempty"`
}
