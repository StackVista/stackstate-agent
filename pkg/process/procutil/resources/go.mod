// This empty go.mod is a hack to protect against gopls (the go LSP server) trying to work in this directory.
// This solves the issue of gopls crashing since https://github.com/DataDog/datadog-agent/pull/15843 was merged.

// Issue suggesting this solution: https://github.com/golang/go/issues/42965

replace github.com/go-openapi/testify/v2 => github.com/go-openapi/testify/v2 v2.0.2
