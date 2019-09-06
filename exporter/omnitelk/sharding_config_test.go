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

package omnitelk

import (
	"testing"

	omnitelk "github.com/Omnition/omnition-opentelemetry-service/exporter/omnitelk/gen"
	"github.com/stretchr/testify/assert"
)

func TestNewShardingInMemConfig(t *testing.T) {
	pbConf := &omnitelk.ShardingConfig{
		ShardDefinitions: []*omnitelk.ShardDefinition{
			{
				ShardId:         []byte{1, 2, 3},
				StartingHashKey: []byte{4, 5, 6},
				EndingHashKey:   []byte{7, 8, 9},
			},
			{
				ShardId:         []byte{0, 1, 2},
				StartingHashKey: []byte{3, 4, 5},
				EndingHashKey:   []byte{6, 7, 8},
			},
		},
	}

	sc := NewShardingInMemConfig(pbConf)

	want := &ShardingInMemConfig{
		shards: []ShardInMemConfig{
			{
				shardID: []byte{0, 1, 2},
			},
			{
				shardID: []byte{1, 2, 3},
			},
		},
	}
	want.shards[0].startingHashKey.SetUint64(197637)
	want.shards[0].endingHashKey.SetUint64(395016)

	want.shards[1].startingHashKey.SetUint64(263430)
	want.shards[1].endingHashKey.SetUint64(460809)

	assert.Equal(t, sc, want)
}
