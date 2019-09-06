// Copyright 2019 Omnition Inc.
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
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/jaegertracing/jaeger/model"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"

	omnitelk "github.com/Omnition/omnition-opentelemetry-service/exporter/omnitelk/gen"
)

type sharder struct {
	config    atomic.Value
	options   *Options
	producers []*shardProducer
	logger    *zap.Logger
	hooks     *kinesisHooks
	semaphore chan struct{}
}

// Options are the options to be used when initializing a Jaeger exporter.
type Options struct {
	Name                  string
	QueueSize             int
	NumWorkers            int
	MaxListSize           int
	ListFlushInterval     int
	MaxAllowedSizePerSpan int
}

// newSharder returns a sharder implementation that exports
// the collected spans to Jaeger.
func newSharder(o *Options, logger *zap.Logger) *sharder {

	if o.MaxListSize == 0 {
		o.MaxListSize = 100000
	}

	if o.ListFlushInterval == 0 {
		o.ListFlushInterval = 5
	}

	if o.MaxAllowedSizePerSpan == 0 {
		o.MaxAllowedSizePerSpan = 900000
	}

	if o.QueueSize == 0 {
		o.QueueSize = 100000
	}
	if o.NumWorkers == 0 {
		o.NumWorkers = 8
	}

	s := &sharder{
		options: o,
		logger:  logger,
	}

	return s
}

// SetConfig sets the sharding configuration. Normally called first during startup
// and can be called multiple times later when new config is available. Safe to call
// while processing is in progress.
func (s *sharder) SetConfig(config *omnitelk.ShardingConfig) {
	s.stopShardProducers()

	cfg := NewShardingInMemConfig(config)
	s.config.Store(cfg)

	s.startShardProducers(cfg)
}

// Stops shard producers. Waits until producers are stopped. Any currently processing
// spans that are not consumed will ????
func (s *sharder) stopShardProducers() {
	for _, sp := range s.producers {
		sp.stop()
	}

	// TODO: return unprocessed Spans to processing queue.
}

// startShardProducers is starts shard producers. Safe to call after calling
// stopShardProducers.
func (s *sharder) startShardProducers(cfg *ShardingInMemConfig) {
	o := s.options
	producers := make([]*shardProducer, 0, len(cfg.shards))
	for _, shard := range cfg.shards {
		hooks := newKinesisHooks(o.Name, shard.shardID)
		producers = append(producers, &shardProducer{
			shard:         shard,
			hooks:         hooks,
			maxSize:       uint64(o.MaxListSize),
			flushInterval: time.Duration(o.ListFlushInterval) * time.Second,
			partitionKey:  shard.startingHashKey.String(),
		})
	}

	s.producers = producers
	s.hooks = newKinesisHooks(o.Name, "")
	s.semaphore = nil

	maxReceivers, _ := strconv.Atoi(os.Getenv("MAX_KINESIS_RECEIVERS"))
	if maxReceivers > 0 {
		s.semaphore = make(chan struct{}, maxReceivers)
	}

	v := metricViews()
	if err := view.Register(v...); err != nil {
		s.logger.Error("Cannot register metric views", zap.Error(err))
		return
	}

	for _, sp := range s.producers {
		sp.start()
	}
}

// Stop flushes queues and stops exporters.
func (s *sharder) Stop() {
	for _, sp := range s.producers {
		sp.stop()
	}
	close(s.semaphore)
}

func (s *sharder) acquire() {
	if s.semaphore != nil {
		s.semaphore <- struct{}{}
	}
}

func (s *sharder) release() {
	if s.semaphore != nil {
		<-s.semaphore
	}
}

// ExportSpan exports a Jaeger protobuf span to Kinesis.
func (s *sharder) ExportSpan(span *model.Span) error {
	s.hooks.OnSpanEnqueued()
	s.acquire()
	go s.processSpan(span)
	return nil
}

func (s *sharder) processSpan(span *model.Span) {
	defer s.release()
	s.hooks.OnSpanDequeued()
	sp, err := s.getShardProducer(span.TraceID.String())
	if err != nil {
		fmt.Println("failed to get producer/shard for traceID: ", err)
		return
	}
	encoded, err := proto.Marshal(span)
	if err != nil {
		fmt.Println("failed to marshal: ", err)
		return
	}
	size := len(encoded)
	if size > s.options.MaxAllowedSizePerSpan {
		sp.hooks.OnXLSpanDropped(size)
		span.Tags = []model.KeyValue{
			{Key: "omnition.dropped", VBool: true, VType: model.ValueType_BOOL},
			{Key: "omnition.dropped.reason", VStr: "unsupported size", VType: model.ValueType_STRING},
			{Key: "omnition.dropped.size", VInt64: int64(size), VType: model.ValueType_INT64},
		}
		span.Logs = []model.Log{}
		encoded, err = proto.Marshal(span)
		if err != nil {
			fmt.Println("failed to modified span: ", err)
			return
		}
		size = len(encoded)
	}
	// TODO: See if we can encode only once and put encoded span on the shard producer.
	// shard producer will have to arrange the bytes exactly as protobuf marshaller would
	// encode a SpanList object.
	// err = sp.pr.Put(encoded, traceID)
	err = sp.put(span, uint64(size))
	if err != nil {
		fmt.Println("error putting span: ", err)
	}
}

func (s *sharder) getShardProducer(partitionKey string) (*shardProducer, error) {
	for _, sp := range s.producers {
		ok, err := sp.shard.belongsToShard(partitionKey)
		if err != nil {
			return nil, err
		}
		if ok {
			return sp, nil
		}
	}
	return nil, fmt.Errorf("no shard found for parition key %s", partitionKey)
}
