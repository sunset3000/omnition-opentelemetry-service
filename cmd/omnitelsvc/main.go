// Program omnitelsvc is the Omnition Telemetry Service built on top of
// OpenTelemetry Service.
package main

import "github.com/open-telemetry/opentelemetry-service/otelsvc"

func main() {
	otelsvc.Run()
}
