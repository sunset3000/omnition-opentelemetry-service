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
	"context"

	"github.com/cenkalti/backoff"
	jaeger "github.com/jaegertracing/jaeger/model"
	"github.com/open-telemetry/opentelemetry-service/consumer/consumerdata"
	jaegertranslator "github.com/open-telemetry/opentelemetry-service/translator/trace/jaeger"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	omnitelk "github.com/Omnition/omnition-opentelemetry-service/exporter/omnitelk/gen"
)

// Maximum number of batches allowed in the input queue.
const inputQueueSize = 1000

// Exporter implements an OpenTelemetry trace exporter that exports spans via
// OmnitelK protocol.
type Exporter struct {
	// User-specified config.
	cfg *Config

	// Input queue to processInputQueue.
	toProcess chan *jaeger.Batch

	// Sharder that performs processing.
	sharder *sharder

	// gRPC client.
	client omnitelk.OmnitelKClient

	// Channel for "done" signalling.
	done chan interface{}

	logger *zap.Logger
}

// NewExporter creates a new Exporter based on user config.
func NewExporter(cfg *Config, logger *zap.Logger) *Exporter {
	return &Exporter{
		cfg:       cfg,
		toProcess: make(chan *jaeger.Batch, inputQueueSize),
		logger:    logger,
		sharder:   newSharder(&Options{}, logger),
		done:      make(chan interface{}),
	}
}

// Start exporter. Begins by connecting to the specified endpoint.
func (e *Exporter) Start() {
	go e.connect()
}

// Stop exporter. Exporter cannot be started after it is stopped. Safe to call
// concurrently with other public functions.
func (e *Exporter) Stop() error {
	close(e.toProcess)
	close(e.done)
	e.sharder.Stop()
	return nil
}

func (e *Exporter) connect() {
	// Try connecting to the server. Use retries with exponential backoff strategy.
	b := backoff.NewExponentialBackOff()

	// Try infinitely (or until Stopped, see "done" channel).
	b.MaxElapsedTime = 0

	// Create ticker that ticks with desired backoff policy.
	ticker := backoff.NewTicker(b)

	connected := false
	for !connected {
		select {
		case <-ticker.C:
			// Got a tick, time to try to connect.
			err := e.tryConnect()
			if err != nil {
				e.logger.Error("Error connecting",
					zap.Error(err),
					zap.String("endpoint", e.cfg.Endpoint))
				continue
			}
			// Connection successful.
			ticker.Stop()
			connected = true

		case <-e.done:
			// Stopped while connecting.
			ticker.Stop()
			return
		}
	}

	// Connected. Now get the sharding config.
	shardingConfig, err := e.client.GetShardingConfig(context.Background(), &omnitelk.ConfigRequest{})
	if err != nil {
		e.logger.Error("Error fetching sharding config",
			zap.Error(err))
		return
	}

	// Initialize sharder.
	e.sharder.SetConfig(shardingConfig)

	// Initialization is done. Begin processing spans.
	go e.processInputQueue()
}

func (e *Exporter) tryConnect() error {
	// Set up a connection to the server. TODO: add secure connection support.
	conn, err := grpc.Dial(e.cfg.Endpoint, grpc.WithInsecure())
	if err != nil {
		return err
	}
	e.client = omnitelk.NewOmnitelKClient(conn)
	return nil
}

// ConsumeTraceData receives a span batch and exports it to OmnitelK server.
func (e *Exporter) ConsumeTraceData(c context.Context, td consumerdata.TraceData) error {
	// Translate to Jaeger format. This is the only operation that does not depend
	// on sharding and which we can perform regardless of sharding configuration.
	// Subsequent processing operations depend on sharding and may need to be repeated
	// when sharding configuration changes, so they will go through toProcess pipeline
	// again in that case.
	pBatch, err := jaegertranslator.OCProtoToJaegerProto(td)
	if err != nil {
		e.logger.Error("error translating span batch", zap.Error(err))
		return err
	}
	// Populate Process field in each Span for easier processing.
	var exportErr error
	for _, span := range pBatch.GetSpans() {
		if span.Process == nil {
			span.Process = pBatch.Process
		}
	}

	// Add the batch to processing queue.
	e.toProcess <- pBatch
	return exportErr
}

func (e *Exporter) processInputQueue() {
	// Here we don't need to watch "done" signal since Stop() will close toProcess
	// channel when stopping.
	for batch := range e.toProcess {
		for _, span := range batch.GetSpans() {
			e.sharder.ExportSpan(span)
		}
	}
}
