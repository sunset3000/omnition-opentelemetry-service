package kinesis

import (
	kinesis "github.com/omnition/opencensus-go-exporter-kinesis"
	"github.com/open-telemetry/opentelemetry-service/config/configerror"
	"github.com/open-telemetry/opentelemetry-service/config/configmodels"
	"github.com/open-telemetry/opentelemetry-service/consumer"
	"github.com/open-telemetry/opentelemetry-service/exporter"
	"go.uber.org/zap"
)

const (
	// The value of "type" key in configuration.
	typeStr      = "kinesis"
	exportFormat = "jaeger-proto"
)

// Factory is the factory for Kinesis exporter.
type Factory struct {
}

// Type gets the type of the Exporter config created by this factory.
func (f *Factory) Type() string {
	return typeStr
}

// CreateDefaultConfig creates the default configuration for exporter.
func (f *Factory) CreateDefaultConfig() configmodels.Exporter {
	return &Config{
		AWS: AWSConfig{
			Region: "us-west-2",
		},
		KPL: KPLConfig{
			BatchSize:            5242880,
			BatchCount:           1000,
			BacklogCount:         2000,
			FlushIntervalSeconds: 5,
			MaxConnections:       24,
		},

		QueueSize:            100000,
		NumWorkers:           8,
		FlushIntervalSeconds: 5,
		MaxBytesPerBatch:     100000,
		MaxBytesPerSpan:      900000,
	}
}

// CreateTraceExporter initializes and returns a new trace exporter
func (f *Factory) CreateTraceExporter(logger *zap.Logger, cfg configmodels.Exporter) (consumer.TraceConsumer, exporter.StopFunc, error) {
	c := cfg.(*Config)
	k, err := kinesis.NewExporter(kinesis.Options{
		Name:               c.Name(),
		StreamName:         c.AWS.StreamName,
		AWSRegion:          c.AWS.Region,
		AWSRole:            c.AWS.Role,
		AWSKinesisEndpoint: c.AWS.KinesisEndpoint,

		KPLAggregateBatchSize:   c.KPL.AggregateBatchSize,
		KPLAggregateBatchCount:  c.KPL.AggregateBatchCount,
		KPLBatchSize:            c.KPL.BatchSize,
		KPLBatchCount:           c.KPL.BatchCount,
		KPLBacklogCount:         c.KPL.BacklogCount,
		KPLFlushIntervalSeconds: c.KPL.FlushIntervalSeconds,
		KPLMaxConnections:       c.KPL.MaxConnections,
		KPLMaxRetries:           c.KPL.MaxRetries,
		KPLMaxBackoffSeconds:    c.KPL.MaxBackoffSeconds,

		QueueSize:             c.QueueSize,
		NumWorkers:            c.NumWorkers,
		MaxAllowedSizePerSpan: c.MaxBytesPerSpan,
		MaxListSize:           c.MaxBytesPerBatch,
		ListFlushInterval:     c.FlushIntervalSeconds,
		Encoding:              exportFormat,
	}, logger)
	if err != nil {
		return nil, nil, err
	}
	stopFunc := func() error {
		k.Flush()
		return nil
	}
	return Exporter{k, logger}, stopFunc, nil
}

// CreateMetricsExporter creates a metrics exporter based on this config.
func (f *Factory) CreateMetricsExporter(logger *zap.Logger, cfg configmodels.Exporter) (consumer.MetricsConsumer, exporter.StopFunc, error) {
	return nil, nil, configerror.ErrDataTypeIsNotSupported
}
