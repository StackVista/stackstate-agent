module github.com/StackVista/stackstate-agent/pkg/otlp/model

go 1.21

replace github.com/StackVista/stackstate-agent/pkg/quantile => ../../quantile

require (
	github.com/StackVista/stackstate-agent/pkg/quantile v0.19.0-rc.4
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector/model v0.38.0
	go.uber.org/zap v1.19.1
)
