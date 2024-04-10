package awslambdainstrumentation

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/trace/api"
	"github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters/instrumentations/aws-lambda/modules"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// InstrumentationIdentifier Identifier for this instrumentation library
// This identifier is based on the one used within the library using within the wild
// https://www.npmjs.com/package/@opentelemetry/instrumentation-aws-lambda
var InstrumentationIdentifier = "@opentelemetry/instrumentation-aws-lambda"

// InterpretBuilderForAwsLambdaInstrumentation Mapping for modules used within the Instrumentation library
// Modules are basically sub parts that have certain functionality. Based on that functionality different context
// values needs to be mapped to be part of the span response
func InterpretBuilderForAwsLambdaInstrumentation() string {
	lambdaInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, modules.OpenTelemetryLambdaEntryServiceIdentifier)
	log.Debugf("[OTEL] [INSTRUMENTATION-AWS-LAMBDA] Mapping service: %s", lambdaInterpreterIdentifier)
	return lambdaInterpreterIdentifier
}
