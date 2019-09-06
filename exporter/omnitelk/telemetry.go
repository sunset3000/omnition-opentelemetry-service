// Copyright 2019, Omnition
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package omnitelk

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var latencyDistributionAggregation = view.Distribution(
	10, 25, 50, 75, 100, 250, 500, 750, 1000, 2000, 3000, 4000, 5000, 10000, 20000, 30000, 50000,
)

// Keys and stats for telemetry.
var (
	tagExporterName, _ = tag.NewKey("exporter")
	tagShardID, _      = tag.NewKey("shard_id")
	tagFlushReason, _  = tag.NewKey("flush_reason")
	tagErrCode, _      = tag.NewKey("err_code")

	statXLSpansBytes = stats.Int64("kinesis_xl_span_size", "size of spans bigger than max support size", stats.UnitBytes)
	statXLSpans      = stats.Int64("kinesis_xl_spans", "number of spans found to be bigger than the max support size", stats.UnitDimensionless)

	statFlushedSpans     = stats.Int64("kinesis_flushed_spans", "number of total spans flushed to kinesis exporter", stats.UnitDimensionless)
	statFlushedSpanLists = stats.Int64("kinesis_flushed_spanlists", "number of spanlists flushed to kinesis exporter", stats.UnitDimensionless)
	statSpanListBytes    = stats.Int64("kinesis_spanlist_bytes", "size of span list in bytes", stats.UnitBytes)

	statEnqueuedSpans = stats.Int64("kinesis_enqueued_spans", "spans received and put in a queue to be processed by kinesis exporter", stats.UnitDimensionless)
	statDequeuedSpans = stats.Int64("kinesis_dequeued_spans", "spans taken out of queue and processed by kinesis exporter", stats.UnitDimensionless)

	statCompressFactor = stats.Int64("kinesis_compress_factor", "compression factor achieved by spanlists", stats.UnitDimensionless)
)

// TODO: support telemetry level

// MetricViews return the metrics views according to given telemetry level.
func metricViews() []*view.View {
	tagKeys := []tag.Key{tagExporterName, tagShardID, tagFlushReason}

	// There are some metricskinesis enabled, return the views.

	spansFlushedView := &view.View{
		Name:        statFlushedSpans.Name(),
		Measure:     statFlushedSpans,
		Description: "Number of spans flushed.",
		TagKeys:     tagKeys,
		Aggregation: view.Sum(),
	}

	spanListsFlushedView := &view.View{
		Name:        statFlushedSpanLists.Name(),
		Measure:     statFlushedSpanLists,
		Description: "Number of spanlists flushed.",
		TagKeys:     tagKeys,
		Aggregation: view.Sum(),
	}

	spanListBytesView := &view.View{
		Name:        statSpanListBytes.Name(),
		Measure:     statSpanListBytes,
		Description: "Size of a spanlist in bytes",
		TagKeys:     tagKeys,
		Aggregation: view.LastValue(),
	}

	xlSpansBytesView := &view.View{
		Name:        statXLSpansBytes.Name(),
		Measure:     statXLSpansBytes,
		Description: "size of spans found that were larger than max supported size",
		TagKeys:     tagKeys,
		Aggregation: view.Sum(),
	}

	xlSpansView := &view.View{
		Name:        statXLSpans.Name(),
		Measure:     statXLSpans,
		Description: "number of spans found that were larger than max supported size",
		TagKeys:     tagKeys,
		Aggregation: view.Sum(),
	}

	enqueuedSpansView := &view.View{
		Name:        statEnqueuedSpans.Name(),
		Measure:     statEnqueuedSpans,
		Description: "spans received and put in a queue to be processed by kinesis exporter",
		TagKeys:     tagKeys,
		Aggregation: view.Sum(),
	}

	dequeuedSpansView := &view.View{
		Name:        statDequeuedSpans.Name(),
		Measure:     statDequeuedSpans,
		Description: "spans taken out of queue and processed by kinesis exporter",
		TagKeys:     tagKeys,
		Aggregation: view.Sum(),
	}

	compressFactorView := &view.View{
		Name:        statCompressFactor.Name(),
		Measure:     statCompressFactor,
		Description: "compression factor achieved per spanlist",
		TagKeys:     tagKeys,
		Aggregation: view.LastValue(),
	}

	return []*view.View{
		compressFactorView,
		spansFlushedView,
		spanListsFlushedView,
		spanListBytesView,
		xlSpansBytesView,
		xlSpansView,
		enqueuedSpansView,
		dequeuedSpansView,
	}
}
