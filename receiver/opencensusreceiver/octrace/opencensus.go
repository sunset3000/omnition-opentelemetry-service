package octrace

import (
	"context"
	"errors"
	"io"
	"sync/atomic"

	commonpb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	agenttracepb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/trace/v1"
	resourcepb "github.com/census-instrumentation/opencensus-proto/gen-go/resource/v1"
	"github.com/open-telemetry/opentelemetry-service/consumer"
	"github.com/open-telemetry/opentelemetry-service/consumer/consumerdata"
	"github.com/open-telemetry/opentelemetry-service/observability"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultNumWorkers             = 4
	messageChannelSize            = 64
	receiverUnaryTagValue         = "oc_trace_unary"
	receiverBiDirectionalTagValue = "oc_trace"
)

// Receiver is the type used to handle spans from OpenCensus exporters.
type Receiver struct {
	numWorkers       int
	backPressureOn   bool
	maxServerStreams int64

	nextConsumer       consumer.TraceConsumer
	workers            []*receiverWorker
	messageChan        chan *traceDataWithCtx
	serverStreamsCount int64
}

type traceDataWithCtx struct {
	data *consumerdata.TraceData
	ctx  context.Context
}

// New creates a new opencensus.Receiver reference.
func New(nextConsumer consumer.TraceConsumer, opts ...Option) (*Receiver, error) {
	if nextConsumer == nil {
		return nil, errors.New("needs a non-nil consumer.TraceConsumer")
	}

	messageChan := make(chan *traceDataWithCtx, messageChannelSize)
	ocr := &Receiver{
		nextConsumer: nextConsumer,
		numWorkers:   defaultNumWorkers,
		messageChan:  messageChan,
	}
	for _, opt := range opts {
		opt(ocr)
	}

	// Setup and startup worker pool
	workers := make([]*receiverWorker, 0, ocr.numWorkers)
	for index := 0; index < ocr.numWorkers; index++ {
		worker := newReceiverWorker(ocr)
		go worker.listenOn(messageChan)
		workers = append(workers, worker)
	}
	ocr.workers = workers

	return ocr, nil
}

var _ agenttracepb.TraceServiceServer = (*Receiver)(nil)

var errUnimplemented = errors.New("unimplemented")

// Config handles configuration messages.
func (ocr *Receiver) Config(tcs agenttracepb.TraceService_ConfigServer) error {
	// TODO: Implement when we define the config receiver/sender.
	return errUnimplemented
}

var errTraceExportProtocolViolation = errors.New("protocol violation: Export's first message must have a Node")

// ExportOne handles unary export calls made by grpc clients
func (ocr *Receiver) ExportOne(ctx context.Context, req *agenttracepb.ExportTraceServiceRequest) (*agenttracepb.ExportTraceServiceResponse, error) {
	// Every batch must have node information when exported using unary rpc
	if req.Node == nil {
		return nil, status.Error(codes.InvalidArgument, "Node must be specified")
	}

	// We need to ensure that it propagates the receiver name as a tag
	ctxWithReceiverName := observability.ContextWithReceiverName(ctx, receiverUnaryTagValue)
	_, _ = ocr.dispatchReceivedMsg(ctxWithReceiverName, nil, nil, req)
	return &agenttracepb.ExportTraceServiceResponse{}, nil
}

// Export is the gRPC method that receives streamed traces from
// OpenCensus-traceproto compatible libraries/applications.
func (ocr *Receiver) Export(tes agenttracepb.TraceService_ExportServer) error {
	if ocr.maxServerStreams > 0 {
		count := atomic.AddInt64(&ocr.serverStreamsCount, 1)
		defer atomic.AddInt64(&ocr.serverStreamsCount, -1)
		if count > ocr.maxServerStreams {
			return status.Errorf(codes.ResourceExhausted, "max-server-streams %d rechead", ocr.maxServerStreams)
		}
	}

	// We need to ensure that it propagates the receiver name as a tag
	ctxWithReceiverName := observability.ContextWithReceiverName(tes.Context(), receiverBiDirectionalTagValue)

	// The first message MUST have a non-nil Node.
	recv, err := tes.Recv()
	if err != nil {
		return err
	}

	// Check the condition that the first message has a non-nil Node.
	if recv.Node == nil {
		return errTraceExportProtocolViolation
	}

	var lastNonNilNode *commonpb.Node
	var resource *resourcepb.Resource
	// Now that we've got the first message with a Node, we can start to receive streamed up spans.
	for {
		lastNonNilNode, resource = ocr.dispatchReceivedMsg(ctxWithReceiverName, lastNonNilNode, resource, recv)

		recv, err = tes.Recv()
		if err != nil {
			if err == io.EOF {
				// Do not return EOF as an error so that grpc-gateway calls get an empty
				// response with HTTP status code 200 rather than a 500 error with EOF.
				return nil
			}
			return err
		}
	}
}

// Stop the receiver and its workers
func (ocr *Receiver) Stop() {
	for _, worker := range ocr.workers {
		worker.stopListening()
	}
}

func (ocr *Receiver) dispatchReceivedMsg(
	ctx context.Context,
	lastNonNilNode *commonpb.Node,
	resource *resourcepb.Resource,
	recv *agenttracepb.ExportTraceServiceRequest,
) (*commonpb.Node, *resourcepb.Resource) {
	// If a Node has been sent from downstream, save and use it.
	if recv.Node != nil {
		lastNonNilNode = recv.Node
	}

	// TODO(songya): differentiate between unset and nil resource. See
	// https://github.com/census-instrumentation/opencensus-proto/issues/146.
	if recv.Resource != nil {
		resource = recv.Resource
	}

	td := &consumerdata.TraceData{
		Node:         lastNonNilNode,
		Resource:     resource,
		Spans:        recv.Spans,
		SourceFormat: "oc_trace",
	}

	ocr.messageChan <- &traceDataWithCtx{data: td, ctx: ctx}
	return lastNonNilNode, resource
}

type receiverWorker struct {
	receiver *Receiver
	tes      agenttracepb.TraceService_ExportServer
	cancel   chan struct{}
}

func newReceiverWorker(receiver *Receiver) *receiverWorker {
	return &receiverWorker{
		receiver: receiver,
		cancel:   make(chan struct{}),
	}
}

func (rw *receiverWorker) listenOn(cn <-chan *traceDataWithCtx) {
	for {
		select {
		case tdWithCtx := <-cn:
			rw.export(tdWithCtx.ctx, tdWithCtx.data)
		case <-rw.cancel:
			return
		}
	}
}

func (rw *receiverWorker) stopListening() {
	close(rw.cancel)
}

func (rw *receiverWorker) export(longLivedCtx context.Context, tracedata *consumerdata.TraceData) {
	if tracedata == nil {
		return
	}

	if len(tracedata.Spans) == 0 {
		observability.RecordTraceReceiverMetrics(longLivedCtx, 0, 0)
		return
	}

	// Trace this method
	ctx, span := trace.StartSpan(context.Background(), "OpenCensusTraceReceiver.Export")
	defer span.End()

	// TODO: (@odeke-em) investigate if it is necessary
	// to group nodes with their respective spans during
	// spansAndNode list unfurling then send spans grouped per node

	// If the starting RPC has a parent span, then add it as a parent link.
	observability.SetParentLink(longLivedCtx, span)

	// TODO: propagate err back somehow to enable ACK
	err := rw.receiver.nextConsumer.ConsumeTraceData(ctx, *tracedata)
	if err != nil {
		observability.RecordTraceReceiverMetrics(longLivedCtx, 0, len(tracedata.Spans))
		span.Annotate([]trace.Attribute{
			trace.Int64Attribute("dropped_spans", int64(len(tracedata.Spans))),
		}, "")
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeDataLoss,
			Message: err.Error(),
		})
	} else {
		observability.RecordTraceReceiverMetrics(longLivedCtx, len(tracedata.Spans), 0)
		span.Annotate([]trace.Attribute{
			trace.Int64Attribute("num_spans", int64(len(tracedata.Spans))),
		}, "")
	}

}
