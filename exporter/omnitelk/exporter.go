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

	"github.com/open-telemetry/opentelemetry-service/consumer/consumerdata"
	jaegertranslator "github.com/open-telemetry/opentelemetry-service/translator/trace/jaeger"
	"go.uber.org/zap"
)

// Exporter implements an OpenTelemetry trace exporter that exports all spans via
// OmnitelK protocol.
type Exporter struct {
	logger *zap.Logger
}

// ConsumeTraceData receives a span batch and exports it to OmnitelK server.
func (e Exporter) ConsumeTraceData(c context.Context, td consumerdata.TraceData) error {
	pBatch, err := jaegertranslator.OCProtoToJaegerProto(td)
	if err != nil {
		e.logger.Error("error translating span batch", zap.Error(err))
		return err
	}
	var exportErr error
	for _, span := range pBatch.GetSpans() {
		if span.Process == nil {
			span.Process = pBatch.Process
		}
		// TODO: export spans.
	}
	return exportErr
}
