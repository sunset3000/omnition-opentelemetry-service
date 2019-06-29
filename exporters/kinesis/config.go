package kinesis

import (
	"github.com/open-telemetry/opentelemetry-service/models"
)

type awsConfig struct {
	StreamName      string `mapstructure:"stream-name,omitempty"`
	KinesisEndpoint string `mapstructure:"kinesis-endpoint,omitempty"`
	Region          string `mapstructure:"region,omitempty"`
	Role            string `mapstructure:"role,omitempty"`
}

type kplConfig struct {
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

type config struct {
	models.ExporterSettings `mapstructure:",squash"`

	AWS awsConfig `mapstructure:"aws,omitempty"`
	KPL kplConfig `mapstructure:"kpl,omitempty"`

	QueueSize            int `mapstructure:"queue-size,omitempty"`
	NumWorkers           int `mapstructure:"num-workers,omitempty"`
	MaxBytesPerBatch     int `mapstructure:"max-bytes-per-batch,omitempty"`
	MaxBytesPerSpan      int `mapstructure:"max-bytes-per-span,omitempty"`
	FlushIntervalSeconds int `mapstructure:"flush-interval-seconds,omitempty"`
}
