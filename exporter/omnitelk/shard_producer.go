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
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"sync"
	"time"

	omnitelk "github.com/Omnition/omnition-opentelemetry-service/exporter/omnitelk/gen"

	"github.com/gogo/protobuf/proto"
	jaeger "github.com/jaegertracing/jaeger/model"
	model "github.com/omnition/opencensus-go-exporter-kinesis/models/gen"
)

const avgBatchSize = 1000

var compressedMagicByte = [8]byte{111, 109, 58, 106, 115, 112, 108, 122}

type shardProducer struct {
	sync.RWMutex

	// pr            *producer.Producer
	shard         ShardInMemConfig
	hooks         *kinesisHooks
	maxSize       uint64
	flushInterval time.Duration
	partitionKey  string

	gzipWriter *gzip.Writer
	spans      *model.SpanList
	size       uint64

	// Called when encoded record is ready.
	produce func(record *omnitelk.EncodedRecord) error
}

func (sp *shardProducer) start() {
	sp.gzipWriter = gzip.NewWriter(&bytes.Buffer{})
	sp.spans = &model.SpanList{Spans: make([]*jaeger.Span, 0, avgBatchSize)}
	sp.size = 0

	// sp.pr.Start()
	go sp.flushPeriodically()
}

func (sp *shardProducer) stop() {
	// TODO: implement stop.
}

func (sp *shardProducer) currentSize() uint64 {
	sp.RLock()
	defer sp.RUnlock()
	return sp.size
}

func (sp *shardProducer) put(span *jaeger.Span, size uint64) error {
	// flush the queue and enqueue new span
	if sp.currentSize()+size >= sp.maxSize {
		sp.flush()
	}

	sp.Lock()
	sp.spans.Spans = append(sp.spans.Spans, span)
	sp.size += size
	sp.Unlock()
	// atomic.AddUint64(&sp.size, size)
	return nil
}

func (sp *shardProducer) flush() {
	sp.Lock()
	defer sp.Unlock()

	numSpans := len(sp.spans.Spans)
	if numSpans == 0 {
		return
	}
	encoded, err := proto.Marshal(sp.spans)
	if err != nil {
		fmt.Println("failed to marshal: ", err)
		return
	}

	compressed := sp.compress(encoded)

	record := &omnitelk.EncodedRecord{
		Data:              compressed,
		PartitionKey:      sp.spans.Spans[0].TraceID.String(),
		SpanCount:         uint64(numSpans),
		UncompressedBytes: uint64(len(encoded)),
	}
	sp.produce(record)

	sp.hooks.OnCompressed(int64(len(encoded)), int64(len(compressed)))
	sp.hooks.OnPutSpanListFlushed(int64(len(sp.spans.Spans)), int64(len(compressed)))

	// TODO: iterate over and set items to nil to enable GC on them?
	// Re-slicing to zero re-uses the same underlying array insead of re-allocating it.
	// This saves us a huge number of allocations but the downside is that spans from the
	// underlying array are never GC'ed. This should be okay as they'll be overwritten
	// anyway as newer spans arrive. This should allow us to make the spanlist consume
	// a static amount of memory throughout the life of the process.
	sp.spans.Spans = sp.spans.Spans[:0]
	sp.size = 0
}

// compress is unsafe for concurrent usage. caller must protect calls with mutexes
func (sp *shardProducer) compress(in []byte) []byte {
	var buf bytes.Buffer
	buf.Write(compressedMagicByte[:])
	sp.gzipWriter.Reset(&buf)

	_, err := sp.gzipWriter.Write(in)
	if err != nil {
		log.Fatal(err)
	}

	if err := sp.gzipWriter.Close(); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

func (sp *shardProducer) flushPeriodically() {
	ticker := time.NewTicker(sp.flushInterval)
	for {
		// add heuristics to not send very small batches unless
		// with too recent records
		<-ticker.C
		size := sp.currentSize()
		if size > 0 {
			sp.flush()
		}
	}
}
