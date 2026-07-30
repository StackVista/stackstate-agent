module github.com/DataDog/datadog-agent/internal/tools/proto

go 1.25.0

require (
	github.com/favadi/protoc-go-inject-tag v1.4.0
	github.com/golang/mock v1.7.0-rc.1
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10
	github.com/tinylib/msgp v1.6.3
	google.golang.org/grpc v1.82.1
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.5.1
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/philhofer/fwd v1.2.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.44.0 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.39.0 // indirect
	golang.org/x/tools v0.47.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
)

replace github.com/go-openapi/testify/v2 => github.com/go-openapi/testify/v2 v2.4.1
