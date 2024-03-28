package interpreter

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	"github.com/StackVista/stackstate-agent/pkg/trace/config"
	interpreterConfig "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/instrumentations"
	awsLambdaInstrumentationModules "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/instrumentations/aws-lambda/modules"
	awsSdkInstrumentationModules "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/instrumentations/aws-sdk/modules"
	httpInstrumentationModules "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/instrumentations/http/modules"
	stackStateInstrumentationModules "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/instrumentations/stackstate/modules"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/model"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/golang/protobuf/proto"
)

// SpanInterpreterEngine type is used to setup the span interpreters
type SpanInterpreterEngine struct {
	SpanInterpreterEngineContext
	DefaultSpanInterpreter *interpreters.DefaultSpanInterpreter
	SourceInterpreters     map[string]interpreters.SourceInterpreter
	TypeInterpreters       map[string]interpreters.TypeInterpreter
}

// MakeSpanInterpreterEngine creates a SpanInterpreterEngine given the config and interpreters
func MakeSpanInterpreterEngine(config *interpreterConfig.Config, typeIns map[string]interpreters.TypeInterpreter, sourceIns map[string]interpreters.SourceInterpreter) *SpanInterpreterEngine {
	return &SpanInterpreterEngine{
		DefaultSpanInterpreter:       interpreters.MakeDefaultSpanInterpreter(config),
		SpanInterpreterEngineContext: MakeSpanInterpreterEngineContext(config),
		SourceInterpreters:           sourceIns,
		TypeInterpreters:             typeIns,
	}
}

// NewSpanInterpreterEngine creates a SpanInterpreterEngine given the config and the default interpreters
func NewSpanInterpreterEngine(agentConfig *config.AgentConfig) *SpanInterpreterEngine {
	interpreterConf := agentConfig.InterpreterConfig

	typeIns := make(map[string]interpreters.TypeInterpreter, 0)
	typeIns[interpreters.ProcessSpanInterpreterName] = interpreters.MakeProcessSpanInterpreter(interpreterConf)
	typeIns[interpreters.SQLSpanInterpreterName] = interpreters.MakeSQLSpanInterpreter(interpreterConf)

	sourceIns := make(map[string]interpreters.SourceInterpreter, 0)
	sourceIns[interpreters.TraefikSpanInterpreterSpan] = interpreters.MakeTraefikInterpreter(interpreterConf)

	// Open Telemetry - AWS Lambda Instrumentation - Modules
	sourceIns[awsLambdaInstrumentationModules.OpenTelemetryLambdaEntryInterpreterSpan] = awsLambdaInstrumentationModules.MakeOpenTelemetryLambdaEntryInterpreter(interpreterConf)

	// Open Telemetry - AWS SDK Instrumentation - Modules
	sourceIns[awsSdkInstrumentationModules.OpenTelemetryLambdaInterpreterSpan] = awsSdkInstrumentationModules.MakeOpenTelemetryLambdaInterpreter(interpreterConf)
	sourceIns[awsSdkInstrumentationModules.OpenTelemetrySQSInterpreterSpan] = awsSdkInstrumentationModules.MakeOpenTelemetrySQSInterpreter(interpreterConf)
	sourceIns[awsSdkInstrumentationModules.OpenTelemetryS3InterpreterSpan] = awsSdkInstrumentationModules.MakeOpenTelemetryS3Interpreter(interpreterConf)
	sourceIns[awsSdkInstrumentationModules.OpenTelemetrySFNInterpreterSpan] = awsSdkInstrumentationModules.MakeOpenTelemetryStepFunctionsInterpreter(interpreterConf)
	sourceIns[awsSdkInstrumentationModules.OpenTelemetrySNSInterpreterSpan] = awsSdkInstrumentationModules.MakeOpenTelemetrySNSInterpreter(interpreterConf)

	// Open Telemetry - HTTP Instrumentation - Modules
	sourceIns[httpInstrumentationModules.OpenTelemetryHTTPInterpreterSpan] = httpInstrumentationModules.MakeOpenTelemetryHTTPInterpreter(interpreterConf)

	// Open Telemetry - StackState Instrumentation - Modules
	sourceIns[stackStateInstrumentationModules.OpenTelemetryStackStateInterpreterSpan] = stackStateInstrumentationModules.MakeOpenTelemetryStackStateInterpreter(interpreterConf)

	return MakeSpanInterpreterEngine(interpreterConf, typeIns, sourceIns)
}

// Interpret interprets the trace using the configured SpanInterpreterEngine
func (se *SpanInterpreterEngine) Interpret(origTrace pb.Trace) pb.Trace {
	log.Infof("[sts] Interpreting %d spans", len(origTrace))
	// we do not mutate the original trace
	var interpretedTrace = make(pb.Trace, 0)
	groupedSourceSpans := make(map[string][]*pb.Span)

	for _, _span := range origTrace {
		// we do not mutate the original span
		span := proto.Clone(_span).(*pb.Span)

		// check if span is pre-interpreted by the trace client
		if _, found := span.Meta["span.serviceURN"]; found {
			interpretedTrace = append(interpretedTrace, span)
			log.Info("[sts] Append interpretedTrace for SpanID %v", span.SpanID)
		} else {
			log.Info("[sts] Interpreting span %+v", span)
			se.DefaultSpanInterpreter.Interpret(span)

			meta, err := se.extractSpanMetadata(span)
			// no metadata, let's look for the span's source.
			if err != nil {
				if source, found := span.Meta["source"]; found {
					// Unique routing for OpenTelemetry
					if source == api.OpenTelemetrySource {
						source = instrumentations.InterpretBasedOnInstrumentationLibrary(span, source)
					}

					groupedSourceSpans[source] = append(groupedSourceSpans[source], span)
				} else {
					interpretedTrace = append(interpretedTrace, span)
				}
			} else {
				// process different span types
				spanWithMeta := &model.SpanWithMeta{Span: span, SpanMetadata: meta}

				// interpret the type if we have a interpreter, otherwise run it through the process interpreter.
				if interpreter, found := se.TypeInterpreters[meta.Type]; found {
					interpretedTrace = append(interpretedTrace, interpreter.Interpret(spanWithMeta))
				} else {
					//defaults to a process interpreter
					processInterpreter := se.TypeInterpreters[interpreters.ProcessSpanInterpreterName]
					interpretedTrace = append(interpretedTrace, processInterpreter.Interpret(spanWithMeta))
				}
			}
		}
	}

	for source, spans := range groupedSourceSpans {
		if interpreter, found := se.SourceInterpreters[source]; found {
			log.Infof("[sts] interpreted source '%s' span", source)
			interpretedTrace = append(interpretedTrace, interpreter.Interpret(spans)...)
		}
	}

	log.Infof("[sts] Interpreted %d spans", len(interpretedTrace))

	return interpretedTrace
}
