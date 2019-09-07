module github.com/Omnition/internal-opentelemetry-service

go 1.12

require (
	contrib.go.opencensus.io/exporter/ocagent v0.6.0
	github.com/apache/thrift v0.0.0-20161221203622-b2a4d4ae21c7
	github.com/census-instrumentation/opencensus-proto v0.2.2
	github.com/client9/misspell v0.3.4
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.2
	github.com/google/addlicense v0.0.0-20190510175307-22550fa7c1b0
	github.com/grpc-ecosystem/grpc-gateway v1.9.4
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jaegertracing/jaeger v1.9.0
	github.com/jstemmer/go-junit-report v0.0.0-20190106144839-af01ea7f8024
	github.com/omnition/gogoproto-rewriter v0.0.0-20190723134119-239e2d24817f
	github.com/omnition/opencensus-go-exporter-kinesis v0.3.2
	github.com/open-telemetry/opentelemetry-service v0.0.0-20190801182910-a1ecc6f9489f
	github.com/openzipkin/zipkin-go v0.1.6
	github.com/rs/cors v1.6.0
	github.com/soheilhy/cmux v0.1.4
	github.com/stretchr/testify v1.3.0
	go.opencensus.io v0.22.0
	go.uber.org/zap v1.10.0
	golang.org/x/lint v0.0.0-20190409202823-959b441ac422
	golang.org/x/tools v0.0.0-20190730215328-ed3277de2799
	google.golang.org/grpc v1.22.0
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc
	k8s.io/api v0.0.0-20181213150558-05914d821849
	k8s.io/apimachinery v0.0.0-20181127025237-2b1284ed4c93
	k8s.io/client-go v2.0.0-alpha.0.0.20181121191925-a47917edff34+incompatible
)

replace contrib.go.opencensus.io/exporter/ocagent => github.com/omnition/opencensus-go-exporter-ocagent v0.4.8-gogoproto2-unary2

replace github.com/census-instrumentation/opencensus-proto => github.com/omnition/opencensus-proto v0.2.1-gogo-unary
