module github.com/Omnition/omnition-opentelemetry-service

go 1.12

require (
	contrib.go.opencensus.io/exporter/ocagent v0.5.1
	github.com/census-instrumentation/opencensus-proto v0.2.2
	github.com/client9/misspell v0.3.4
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.1
	github.com/google/addlicense v0.0.0-20190510175307-22550fa7c1b0
	github.com/grpc-ecosystem/grpc-gateway v1.9.0
	github.com/jstemmer/go-junit-report v0.0.0-20190106144839-af01ea7f8024
	github.com/omnition/gogoproto-rewriter v0.0.0-20190723134119-239e2d24817f
	github.com/omnition/opencensus-go-exporter-kinesis v0.3.2
	github.com/open-telemetry/opentelemetry-service v0.0.0-20190731175920-831d805e2d8e
	github.com/rs/cors v1.6.0
	github.com/soheilhy/cmux v0.1.4
	github.com/stretchr/testify v1.3.0
	go.opencensus.io v0.22.0
	go.uber.org/zap v1.10.0
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	golang.org/x/tools v0.0.0-20190710184609-286818132824
	google.golang.org/grpc v1.21.0
	honnef.co/go/tools v0.0.0-20190106161140-3f1c8253044a
)

replace contrib.go.opencensus.io/exporter/ocagent => github.com/omnition/opencensus-go-exporter-ocagent v0.4.8-gogoproto2-unary2

replace github.com/census-instrumentation/opencensus-proto => github.com/omnition/opencensus-proto v0.2.1-gogo-unary
