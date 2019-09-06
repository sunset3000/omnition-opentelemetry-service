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
	"context"
	"sync"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

type tagCache struct {
	sync.RWMutex
	tags map[tag.Key]map[string]tag.Mutator
}

func (c *tagCache) get(name tag.Key, value string) tag.Mutator {
	c.RLock()
	values, ok := c.tags[name]
	c.RUnlock()

	if !ok {
		t := tag.Upsert(name, value)
		c.Lock()
		c.tags[name] = map[string]tag.Mutator{
			value: t,
		}
		c.Unlock()
		return t
	}

	c.RLock()
	t, ok := values[value]
	c.RUnlock()
	if ok {
		return t
	}
	t = tag.Upsert(name, value)
	c.Lock()
	values[value] = t
	c.Unlock()
	return t
}

type kinesisHooks struct {
	exporterName string
	shardID      string
	commonTags   []tag.Mutator
	tagCache     tagCache
}

func newKinesisHooks(name, shardID string) *kinesisHooks {
	tags := []tag.Mutator{
		tag.Upsert(tagExporterName, name),
	}
	if shardID != "" {
		tags = append(tags, tag.Upsert(tagShardID, shardID))
	}
	return &kinesisHooks{
		exporterName: name,
		commonTags:   tags,
		tagCache: tagCache{
			tags: map[tag.Key]map[string]tag.Mutator{},
		},
	}
}

func (h *kinesisHooks) tags(name tag.Key, value string) []tag.Mutator {
	tags := make([]tag.Mutator, 0, len(h.commonTags)+1)
	copy(tags, h.commonTags)
	tags = append(tags, h.tagCache.get(name, value))
	return tags
}

func (h *kinesisHooks) OnDrain(bytes, length int64) {
	_ = stats.RecordWithTags(
		context.Background(),
		h.commonTags,
		statDrainBytes.M(bytes),
		statDrainLength.M(length),
	)
}

func (h *kinesisHooks) OnPutRecords(batches, spanlists, bytes, putLatencyMS int64, reason string) {
	_ = stats.RecordWithTags(
		context.Background(),
		h.tags(tagFlushReason, reason),
		statPutRequests.M(1),
		statPutBatches.M(batches),
		statPutSpanLists.M(spanlists),
		statPutBytes.M(bytes),
		statPutLatency.M(putLatencyMS),
	)
}

func (h *kinesisHooks) OnPutErr(errCode string) {
	_ = stats.RecordWithTags(
		context.Background(),
		h.tags(tagErrCode, errCode),
		statPutErrors.M(1),
	)
}

func (h *kinesisHooks) OnDropped(batches, spanlists, bytes int64) {
	_ = stats.RecordWithTags(
		context.Background(),
		h.commonTags,
		statDroppedBatches.M(batches),
		statDroppedSpanLists.M(spanlists),
		statDroppedBytes.M(bytes),
	)
}

func (h *kinesisHooks) OnSpanEnqueued() {
	_ = stats.RecordWithTags(
		context.Background(),
		h.commonTags,
		statEnqueuedSpans.M(1),
	)
}

func (h *kinesisHooks) OnSpanDequeued() {
	_ = stats.RecordWithTags(
		context.Background(),
		h.commonTags,
		statDequeuedSpans.M(1),
	)
}

func (h *kinesisHooks) OnXLSpanDropped(size int) {
	_ = stats.RecordWithTags(
		context.Background(),
		h.commonTags,
		statXLSpansBytes.M(int64(size)),
		statXLSpans.M(1),
	)
}

func (h *kinesisHooks) OnPutSpanListFlushed(spans, bytes int64) {
	_ = stats.RecordWithTags(
		context.Background(),
		h.commonTags,
		statFlushedSpans.M(spans),
		statSpanListBytes.M(bytes),
	)
}

func (h *kinesisHooks) OnCompressed(original, compressed int64) {
	_ = stats.RecordWithTags(
		context.Background(),
		h.commonTags,
		statCompressFactor.M(original/compressed),
	)
}
