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
	"math/big"
	"sort"

	omnitelk "github.com/Omnition/omnition-opentelemetry-service/exporter/omnitelk/gen"
)

// ShardingInMemConfig is an immutable in-memory representation of sharding
// configuration.
type ShardingInMemConfig struct {
	// List of shards sorted by startingHashKey.
	shards []ShardInMemConfig
}

// ShardInMemConfig  is an immutable in-memory representation of one shard
// configuration.
type ShardInMemConfig struct {
	shardID         []byte
	startingHashKey big.Int
	endingHashKey   big.Int
}

// byStartingHashKey implements sort.Interface for []ShardInMemConfig based on
// the startingHashKey field.
type byStartingHashKey []ShardInMemConfig

func (a byStartingHashKey) Len() int      { return len(a) }
func (a byStartingHashKey) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byStartingHashKey) Less(i, j int) bool {
	return a[i].startingHashKey.Cmp(&a[j].startingHashKey) < 0
}

// NewShardingInMemConfig creates a ShardingInMemConfig from omnitelk.ShardingConfig.
func NewShardingInMemConfig(pbConf *omnitelk.ShardingConfig) *ShardingInMemConfig {
	sc := &ShardingInMemConfig{}
	for _, s := range pbConf.ShardDefinitions {
		shard := ShardInMemConfig{
			shardID: s.ShardId,
		}
		shard.startingHashKey.SetBytes(s.StartingHashKey)
		shard.endingHashKey.SetBytes(s.EndingHashKey)

		sc.shards = append(sc.shards, shard)
	}

	sort.Sort(byStartingHashKey(sc.shards))

	return sc
}
