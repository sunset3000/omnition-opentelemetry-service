package kinesis

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-service/configv2"
	"github.com/open-telemetry/opentelemetry-service/models"
)

var _ = configv2.RegisterTestFactories()

func TestDefaultConfig(t *testing.T) {
	cfg, err := configv2.LoadConfigFile(t, path.Join(".", "testdata", "default.yaml"))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	e := cfg.Exporters["kinesis"]

	assert.Equal(t, e,
		&config{
			ExporterSettings: models.ExporterSettings{
				TypeVal: "kinesis",
				NameVal: "kinesis",
				Enabled: true,
			},
			AWS: awsConfig{
				Region: "us-west-2",
			},
			KPL: kplConfig{
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
	cfg, err := configv2.LoadConfigFile(t, path.Join(".", "testdata", "config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	e := cfg.Exporters["kinesis"]

	assert.Equal(t, e,
		&config{
			ExporterSettings: models.ExporterSettings{
				TypeVal: "kinesis",
				NameVal: "kinesis",
				Enabled: true,
			},
			AWS: awsConfig{
				StreamName:      "test-stream",
				KinesisEndpoint: "kinesis.mars-1.aws.galactic",
				Region:          "mars-1",
				Role:            "arn:test-role",
			},
			KPL: kplConfig{
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
