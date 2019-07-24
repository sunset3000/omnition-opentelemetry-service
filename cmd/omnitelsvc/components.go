package main

import (
	"github.com/open-telemetry/opentelemetry-service/exporter"
	"github.com/open-telemetry/opentelemetry-service/exporter/loggingexporter"
	"github.com/open-telemetry/opentelemetry-service/exporter/opencensusexporter"
	"github.com/open-telemetry/opentelemetry-service/exporter/prometheusexporter"
	"github.com/open-telemetry/opentelemetry-service/oterr"
	"github.com/open-telemetry/opentelemetry-service/processor"
	"github.com/open-telemetry/opentelemetry-service/processor/addattributesprocessor"
	"github.com/open-telemetry/opentelemetry-service/processor/attributekeyprocessor"
	"github.com/open-telemetry/opentelemetry-service/processor/nodebatcher"
	"github.com/open-telemetry/opentelemetry-service/processor/queued"
	"github.com/open-telemetry/opentelemetry-service/receiver"
	"github.com/open-telemetry/opentelemetry-service/receiver/jaegerreceiver"
	"github.com/open-telemetry/opentelemetry-service/receiver/zipkinreceiver"

	"github.com/Omnition/internal-opentelemetry-service/exporters/kinesis"
	"github.com/Omnition/internal-opentelemetry-service/receiver/opencensusreceiver"
)

func components() (
	map[string]receiver.Factory,
	map[string]processor.Factory,
	map[string]exporter.Factory,
	error,
) {
	errs := []error{}
	receivers, err := receiver.Build(
		&jaegerreceiver.Factory{},
		&zipkinreceiver.Factory{},
		&opencensusreceiver.Factory{},
	)
	if err != nil {
		errs = append(errs, err)
	}

	exporters, err := exporter.Build(
		&opencensusexporter.Factory{},
		&prometheusexporter.Factory{},
		&loggingexporter.Factory{},
		&kinesis.Factory{},
	)
	if err != nil {
		errs = append(errs, err)
	}

	processors, err := processor.Build(
		&addattributesprocessor.Factory{},
		&attributekeyprocessor.Factory{},
		&queued.Factory{},
		&nodebatcher.Factory{},
	)
	if err != nil {
		errs = append(errs, err)
	}
	return receivers, processors, exporters, oterr.CombineErrors(errs)
}
