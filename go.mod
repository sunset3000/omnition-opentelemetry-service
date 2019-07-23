module github.com/Omnition/internal-opentelemetry-service

go 1.12

require (
	github.com/jstemmer/go-junit-report v0.0.0-20190106144839-af01ea7f8024
	github.com/omnition/gogoproto-rewriter v0.0.0-20190723134119-239e2d24817f
	github.com/omnition/opencensus-go-exporter-kinesis v0.3.2
	github.com/open-telemetry/opentelemetry-service v0.0.0-20190717165254-8905c13995e4
	github.com/stretchr/testify v1.3.0
	go.uber.org/zap v1.10.0
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3
)

replace github.com/census-instrumentation/opencensus-proto => github.com/omnition/opencensus-proto v0.3.0-omnition-gogo
