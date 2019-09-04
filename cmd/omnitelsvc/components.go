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

package main

import (
	"github.com/open-telemetry/opentelemetry-service/exporter"
	"github.com/open-telemetry/opentelemetry-service/exporter/loggingexporter"
	"github.com/open-telemetry/opentelemetry-service/exporter/prometheusexporter"
	"github.com/open-telemetry/opentelemetry-service/oterr"
	"github.com/open-telemetry/opentelemetry-service/processor"
	"github.com/open-telemetry/opentelemetry-service/processor/addattributesprocessor"
	"github.com/open-telemetry/opentelemetry-service/processor/attributekeyprocessor"
	"github.com/open-telemetry/opentelemetry-service/processor/probabilisticsampler"

	"github.com/open-telemetry/opentelemetry-service/processor/nodebatcher"
	"github.com/open-telemetry/opentelemetry-service/processor/queued"
	"github.com/open-telemetry/opentelemetry-service/receiver"
	"github.com/open-telemetry/opentelemetry-service/receiver/jaegerreceiver"
	"github.com/open-telemetry/opentelemetry-service/receiver/zipkinreceiver"

	"github.com/Omnition/omnition-opentelemetry-service/exporter/kinesis"
	"github.com/Omnition/omnition-opentelemetry-service/exporter/opencensusexporter"
	"github.com/Omnition/omnition-opentelemetry-service/processor/memorylimiter"
	"github.com/Omnition/omnition-opentelemetry-service/receiver/opencensusreceiver"
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
		&memorylimiter.Factory{},
		&probabilisticsampler.Factory{},
	)
	if err != nil {
		errs = append(errs, err)
	}
	return receivers, processors, exporters, oterr.CombineErrors(errs)
}
