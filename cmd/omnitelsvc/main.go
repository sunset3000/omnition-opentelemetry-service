// Program omnitelsvc is the Omnition Telemetry Service built on top of
// OpenTelemetry Service.
package main

import (
	"github.com/open-telemetry/opentelemetry-service/otelsvc"

	_ "github.com/Omnition/internal-opentelemetry-service/exporters/kinesis"
)

func main() {
	otelsvc.Run()
}
