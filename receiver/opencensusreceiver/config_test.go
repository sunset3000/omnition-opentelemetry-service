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

package opencensusreceiver

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-service/config"
	"github.com/open-telemetry/opentelemetry-service/config/configmodels"
)

func TestLoadConfig(t *testing.T) {
	receivers, processors, exporters, err := config.ExampleComponents()
	assert.Nil(t, err)

	factory := &Factory{}
	receivers[typeStr] = factory
	cfg, err := config.LoadConfigFile(
		t, path.Join(".", "testdata", "config.yaml"), receivers, processors, exporters,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Receivers), 4)

	r0 := cfg.Receivers["opencensus"]
	assert.Equal(t, r0, factory.CreateDefaultConfig())

	r1 := cfg.Receivers["opencensus/customname"].(*Config)
	assert.Equal(t, r1.ReceiverSettings,
		configmodels.ReceiverSettings{
			TypeVal:  typeStr,
			NameVal:  "opencensus/customname",
			Endpoint: "0.0.0.0:9090",
		})

	r2 := cfg.Receivers["opencensus/keepalive"].(*Config)
	assert.Equal(t, r2,
		&Config{
			ReceiverSettings: configmodels.ReceiverSettings{
				TypeVal:  typeStr,
				NameVal:  "opencensus/keepalive",
				Endpoint: "127.0.0.1:55678",
			},
			Keepalive: &serverParametersAndEnforcementPolicy{
				ServerParameters: &keepaliveServerParameters{
					Time:    30 * time.Second,
					Timeout: 5 * time.Second,
				},
				EnforcementPolicy: &keepaliveEnforcementPolicy{
					MinTime:             10 * time.Second,
					PermitWithoutStream: true,
				},
			},
		})

	r3 := cfg.Receivers["opencensus/nobackpressure"].(*Config)
	assert.Equal(t, r3,
		&Config{
			ReceiverSettings: configmodels.ReceiverSettings{
				TypeVal:             typeStr,
				NameVal:             "opencensus/nobackpressure",
				Endpoint:            "127.0.0.1:55678",
				DisableBackPressure: true,
			},
			MaxRecvMsgSizeMiB:    32,
			MaxConcurrentStreams: 16,
		})
}
