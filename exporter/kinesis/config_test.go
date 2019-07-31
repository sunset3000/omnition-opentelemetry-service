package kinesis

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-service/config"
	"github.com/open-telemetry/opentelemetry-service/config/configmodels"
)

func TestDefaultConfig(t *testing.T) {
	receiverFactories, processorsFactories, exporterFactories, err := config.ExampleComponents()
	assert.Nil(t, err)

	factory := &Factory{}
	exporterFactories[factory.Type()] = factory
	cfg, err := config.LoadConfigFile(
		t, path.Join(".", "testdata", "default.yaml"), receiverFactories, processorsFactories, exporterFactories,
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	e := cfg.Exporters["kinesis"]

	assert.Equal(t, e,
		&Config{
			ExporterSettings: configmodels.ExporterSettings{
				TypeVal: "kinesis",
				NameVal: "kinesis",
			},
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
		},
	)
}

func TestConfig(t *testing.T) {
	receiverFactories, processorsFactories, exporterFactories, err := config.ExampleComponents()
	assert.Nil(t, err)

	factory := &Factory{}
	exporterFactories[factory.Type()] = factory
	cfg, err := config.LoadConfigFile(
		t, path.Join(".", "testdata", "config.yaml"), receiverFactories, processorsFactories, exporterFactories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	e := cfg.Exporters["kinesis"]

	assert.Equal(t, e,
		&Config{
			ExporterSettings: configmodels.ExporterSettings{
				TypeVal: "kinesis",
				NameVal: "kinesis",
			},
			AWS: AWSConfig{
				StreamName:      "test-stream",
				KinesisEndpoint: "kinesis.mars-1.aws.galactic",
				Region:          "mars-1",
				Role:            "arn:test-role",
			},
			KPL: KPLConfig{
				AggregateBatchCount:  10,
				AggregateBatchSize:   11,
				BatchSize:            12,
				BatchCount:           13,
				BacklogCount:         14,
				FlushIntervalSeconds: 15,
				MaxConnections:       16,
				MaxRetries:           17,
				MaxBackoffSeconds:    18,
			},

			QueueSize:            1,
			NumWorkers:           2,
			FlushIntervalSeconds: 3,
			MaxBytesPerBatch:     4,
			MaxBytesPerSpan:      5,
		},
	)
}
