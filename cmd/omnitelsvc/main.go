// Program omnitelsvc is the Omnition Telemetry Service built on top of
// OpenTelemetry Service.
package main

import (
	"log"

	"github.com/Omnition/internal-opentelemetry-service/exporters/kinesis"
	"github.com/open-telemetry/opentelemetry-service/defaults"
	"github.com/open-telemetry/opentelemetry-service/service"
)

func main() {
	handleErr := func(err error) {
		if err != nil {
			log.Fatalf("Failed to run the service: %v", err)
		}
	}

	receivers, processors, exporters, err := defaults.Components()
	handleErr(err)

	kinesisFactory := &kinesis.Factory{}
	exporters[kinesisFactory.Type()] = kinesisFactory
	svc := service.New(receivers, processors, exporters)
	err = svc.StartUnified()
	handleErr(err)
}
