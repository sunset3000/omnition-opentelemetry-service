package kinesis

import (
	"context"

	kinesis "github.com/omnition/opencensus-go-exporter-kinesis"
	"github.com/open-telemetry/opentelemetry-service/consumer/consumerdata"
	jaegertranslator "github.com/open-telemetry/opentelemetry-service/translator/trace/jaeger"
	"go.uber.org/zap"
)

// Exporter implements an OpenTelemetry trace exporter that exports all spans to AWS Kinesis
type Exporter struct {
	kinesis *kinesis.Exporter
	logger  *zap.Logger
}

// ConsumeTraceData receives a span batch and exports it to AWS Kinesis
func (e Exporter) ConsumeTraceData(c context.Context, td consumerdata.TraceData) error {
	pBatch, err := jaegertranslator.OCProtoToJaegerProto(td)
	if err != nil {
		e.logger.Error("error translating span batch", zap.Error(err))
		return err
	}
	// TODO: Use a multi error type
	var exportErr error
	for _, span := range pBatch.GetSpans() {
		if span.Process == nil {
			span.Process = pBatch.Process
		}
		err := e.kinesis.ExportSpan(span)
		if err != nil {
			e.logger.Error("error exporting span to kinesis", zap.Error(err))
			exportErr = err
		}
	}
	return exportErr
}
