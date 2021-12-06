package interpreter

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	"github.com/StackVista/stackstate-agent/pkg/trace/config"
	interpreterConfig "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	openTelemetry "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open_telemetry"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/model"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
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
	sourceIns[openTelemetry.OpenTelemetryLambdaInterpreterSpan] = openTelemetry.MakeOpenTelemetryLambdaInterpreter(interpreterConf)
	sourceIns[openTelemetry.OpenTelemetrySQSInterpreterSpan] = openTelemetry.MakeOpenTelemetrySQSInterpreter(interpreterConf)

	return MakeSpanInterpreterEngine(interpreterConf, typeIns, sourceIns)
}

func UpdateOpenTelemetrySpanSource(source string, span *pb.Span) string {
	if source == "openTelemetry" {
		switch span.Meta["instrumentation_library"] {
		// TODO: instrumentation-aws-sdk is NodeJS support need to add support for more languages

		case "@opentelemetry/instrumentation-aws-lambda":
			return "openTelemetryLambda"

		case "@opentelemetry/instrumentation-aws-sdk":
			switch span.Meta["aws.service.identifier"] {
			case "sqs":
				return "openTelemetrySQS"

			default:
				fmt.Printf("Unknown AWS identifier for Open Telemetry: %v", span.Meta["aws.service.identifier"])
			}
			break

		default:
			fmt.Printf("Unknown Open Telemetry instrumentation library: %v", span.Meta["instrumentation_library"])
		}
	}

	return source
}

// Interpret interprets the trace using the configured SpanInterpreterEngine
func (se *SpanInterpreterEngine) Interpret(origTrace pb.Trace) pb.Trace {
	// we do not mutate the original trace
	var interpretedTrace = make(pb.Trace, 0)
	groupedSourceSpans := make(map[string][]*pb.Span)

	for _, _span := range origTrace {
		// we do not mutate the original span
		span := proto.Clone(_span).(*pb.Span)

		// check if span is pre-interpreted by the trace client
		if _, found := span.Meta["span.serviceURN"]; found {
			interpretedTrace = append(interpretedTrace, span)
		} else {
			se.DefaultSpanInterpreter.Interpret(span)

			meta, err := se.extractSpanMetadata(span)
			// no metadata, let's look for the span's source.
			if err != nil {
				if source, found := span.Meta["source"]; found {
					// Special mapping for open telemetry, Multiple interpreters required for spans
					if source == api.OpenTelemetrySource {
						source = UpdateOpenTelemetrySpanSource(source, span)
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
			interpretedTrace = append(interpretedTrace, interpreter.Interpret(spans)...)
		}
	}

	return interpretedTrace
}
