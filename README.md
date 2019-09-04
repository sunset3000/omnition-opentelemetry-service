# internal-opentelemetry-service
Internal build of OpenTelemetry Service (not a fork)


## Local development

We leverage Go's vendor feature to automatically modify source code of some dependencies before compiling them. This is mainly done to auto translate code that uses Go's default protobuf implementation to use the faster gogoproto one. This means one must run `go mod vendor` after adding, removing or updating dependencies during local development. `make install` automatically runs this and makes sure other tool dependencies are installed as well. 

