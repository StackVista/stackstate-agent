package batcher

import (
	"fmt"
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/serializer"
	"github.com/DataDog/datadog-agent/pkg/util"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/health"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/telemetry"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
)

// Batcher interface can receive data for sending to the intake and will accumulate the data in batches. This does
// not work on a fixed schedule like the aggregator but flushes either when data exceeds a threshold, when
// data is complete.
type Component interface {
	// Topology
	SubmitComponent(checkID checkid.ID, instance topology.Instance, component topology.Component)
	SubmitRelation(checkID checkid.ID, instance topology.Instance, relation topology.Relation)
	SubmitStartSnapshot(checkID checkid.ID, instance topology.Instance)
	SubmitStopSnapshot(checkID checkid.ID, instance topology.Instance)
	SubmitDelete(checkID checkid.ID, instance topology.Instance, topologyElementID string)

	// Health
	SubmitHealthCheckData(checkID checkid.ID, stream health.Stream, data health.CheckData)
	SubmitHealthStartSnapshot(checkID checkid.ID, stream health.Stream, intervalSeconds int, expirySeconds int)
	SubmitHealthStopSnapshot(checkID checkid.ID, stream health.Stream)

	// Raw Metrics
	SubmitRawMetricsData(checkID checkid.ID, data telemetry.RawMetric)

	// lifecycle
	SubmitComplete(checkID checkid.ID)
	Shutdown()
}


func MakeAsynchronousBatcher(serializer serializer.AgentV1Serializer, hostname string, maxCapacity int) Component {
	batcher := AsynchronousBatcher{
		builder:    NewBatchBuilder(maxCapacity),
		hostname:   hostname,
		input:      make(chan interface{}),
		serializer: serializer,
	}
	go batcher.run()
	return batcher
}

// NewMockBatcher initializes the global batcher with a mock version, intended for testing
func NewMockBatcher() *MockBatcher {
	return createMockBatcher()
}

// AsynchronousBatcher is the implementation of the batcher. Works asynchronous. Publishes data to the serializer
type AsynchronousBatcher struct {
	builder             BatchBuilder
	hostname, agentName string
	input               chan interface{}
	serializer          serializer.AgentV1Serializer
}

type submitComponent struct {
	checkID   checkid.ID
	instance  topology.Instance
	component topology.Component
}

type submitRelation struct {
	checkID  checkid.ID
	instance topology.Instance
	relation topology.Relation
}

type submitStartSnapshot struct {
	checkID  checkid.ID
	instance topology.Instance
}

type submitStopSnapshot struct {
	checkID  checkid.ID
	instance topology.Instance
}

type submitDelete struct {
	checkID  checkid.ID
	instance topology.Instance
	deleteID string
}

type submitHealthCheckData struct {
	checkID checkid.ID
	stream  health.Stream
	data    health.CheckData
}

type submitHealthStartSnapshot struct {
	checkID         checkid.ID
	stream          health.Stream
	intervalSeconds int
	expirySeconds   int
}

type submitHealthStopSnapshot struct {
	checkID checkid.ID
	stream  health.Stream
}

type submitRawMetricsData struct {
	checkID   checkid.ID
	rawMetric telemetry.RawMetric
}

type submitComplete struct {
	checkID checkid.ID
}

type submitShutdown struct{}

func (batcher *AsynchronousBatcher) sendState(states CheckInstanceBatchStates) {
	if states != nil {

		// Create the topologies
		topologies := make([]topology.Topology, 0)
		for _, state := range states {
			if state.Topology != nil {
				topologies = append(topologies, *state.Topology)
			}
		}

		// Create the healthData payload
		healthData := make([]health.Health, 0)
		for _, state := range states {
			for _, healthRecord := range state.Health {
				healthData = append(healthData, healthRecord)
			}
		}

		// Create the rawMetricData payload
		rawMetrics := make([]interface{}, 0)
		for _, state := range states {
			if state.Metrics != nil {
				for _, metric := range *state.Metrics {
					rawMetrics = append(rawMetrics, metric.ConvertToIntakeMetric())
				}
			}
		}

		payload := map[string]interface{}{
			"internalHostname": batcher.hostname,
			"topologies":       topologies,
			"health":           healthData,
			"metrics":          rawMetrics,
		}

		// For debug purposes print out all topologies payload
		if config.Datadog.GetBool("log_payloads") {
			log.Debug("Flushing the following topologies:")
			for _, topo := range topologies {
				log.Debugf("%v", topo)
			}

			log.Debug("Flushing the following health data:")
			for _, health := range healthData {
				log.Debugf("%v", health)
			}

			log.Debug("Flushing the following raw metric data:")
			for _, rawMetric := range rawMetrics {
				log.Debugf("%v", rawMetric)
			}
		}

		// For debug purposes print out all topologies payload
		if config.Datadog.GetBool("log_payloads") {
			log.Debug("Flushing the following topologies:")
			for _, topo := range topologies {
				log.Debugf("%v", topo)
			}
		}

		if err := batcher.serializer.SendJSONToV1Intake(payload); err != nil {
			_ = log.Errorf("error in SendJSONToV1Intake: %s", err)
		}
	}
}

func (batcher *AsynchronousBatcher) run() {
	for {
		s := <-batcher.input
		switch submission := s.(type) {
		case submitComponent:
			batcher.sendState(batcher.builder.AddComponent(submission.checkID, submission.instance, submission.component))
		case submitRelation:
			batcher.sendState(batcher.builder.AddRelation(submission.checkID, submission.instance, submission.relation))
		case submitStartSnapshot:
			batcher.sendState(batcher.builder.TopologyStartSnapshot(submission.checkID, submission.instance))
		case submitStopSnapshot:
			batcher.sendState(batcher.builder.TopologyStopSnapshot(submission.checkID, submission.instance))
		case submitDelete:
			batcher.sendState(batcher.builder.Delete(submission.checkID, submission.instance, submission.deleteID))

		case submitHealthCheckData:
			batcher.sendState(batcher.builder.AddHealthCheckData(submission.checkID, submission.stream, submission.data))
		case submitHealthStartSnapshot:
			batcher.sendState(batcher.builder.HealthStartSnapshot(submission.checkID, submission.stream, submission.intervalSeconds, submission.expirySeconds))
		case submitHealthStopSnapshot:
			batcher.sendState(batcher.builder.HealthStopSnapshot(submission.checkID, submission.stream))

		case submitRawMetricsData:
			batcher.sendState(batcher.builder.AddRawMetricsData(submission.checkID, submission.rawMetric))

		case submitComplete:
			batcher.sendState(batcher.builder.FlushIfDataProduced(submission.checkID))
		case submitShutdown:
			return
		default:
			panic(fmt.Sprint("Unknown submission type"))
		}
	}
}

// SubmitComponent submits a component to the batch
func (batcher AsynchronousBatcher) SubmitComponent(checkID checkid.ID, instance topology.Instance, component topology.Component) {
	batcher.input <- submitComponent{
		checkID:   checkID,
		instance:  instance,
		component: component,
	}
}

// SubmitRelation submits a relation to the batch
func (batcher AsynchronousBatcher) SubmitRelation(checkID checkid.ID, instance topology.Instance, relation topology.Relation) {
	batcher.input <- submitRelation{
		checkID:  checkID,
		instance: instance,
		relation: relation,
	}
}

// SubmitStartSnapshot submits start of a snapshot
func (batcher AsynchronousBatcher) SubmitStartSnapshot(checkID checkid.ID, instance topology.Instance) {
	batcher.input <- submitStartSnapshot{
		checkID:  checkID,
		instance: instance,
	}
}

// SubmitStopSnapshot submits a stop of a snapshot. This always causes a flush of the data downstream
func (batcher AsynchronousBatcher) SubmitStopSnapshot(checkID checkid.ID, instance topology.Instance) {
	batcher.input <- submitStopSnapshot{
		checkID:  checkID,
		instance: instance,
	}
}

// SubmitDelete submits a deletion of topology element.
func (batcher AsynchronousBatcher) SubmitDelete(checkID checkid.ID, instance topology.Instance, topologyElementID string) {
	batcher.input <- submitDelete{
		checkID:  checkID,
		instance: instance,
		deleteID: topologyElementID,
	}
}

// SubmitHealthCheckData submits a Health check data record to the batch
func (batcher AsynchronousBatcher) SubmitHealthCheckData(checkID checkid.ID, stream health.Stream, data health.CheckData) {
	log.Debugf("Submitting Health check data for check [%s] stream [%s]: %s", checkID, stream.GoString(), util.JSONString(data))
	batcher.input <- submitHealthCheckData{
		checkID: checkID,
		stream:  stream,
		data:    data,
	}
}

// SubmitHealthStartSnapshot submits start of a Health snapshot
func (batcher AsynchronousBatcher) SubmitHealthStartSnapshot(checkID checkid.ID, stream health.Stream, intervalSeconds int, expirySeconds int) {
	batcher.input <- submitHealthStartSnapshot{
		checkID:         checkID,
		stream:          stream,
		intervalSeconds: intervalSeconds,
		expirySeconds:   expirySeconds,
	}
}

// SubmitHealthStopSnapshot submits a stop of a Health snapshot. This always causes a flush of the data downstream
func (batcher AsynchronousBatcher) SubmitHealthStopSnapshot(checkID checkid.ID, stream health.Stream) {
	batcher.input <- submitHealthStopSnapshot{
		checkID: checkID,
		stream:  stream,
	}
}

// SubmitRawMetricsData submits a raw metrics data record to the batch
func (batcher AsynchronousBatcher) SubmitRawMetricsData(checkID checkid.ID, rawMetric telemetry.RawMetric) {
	if rawMetric.HostName == "" {
		rawMetric.HostName = batcher.hostname
	}

	batcher.input <- submitRawMetricsData{
		checkID:   checkID,
		rawMetric: rawMetric,
	}
}

// SubmitComplete signals completion of a check. May trigger a flush only if the check produced data
func (batcher AsynchronousBatcher) SubmitComplete(checkID checkid.ID) {
	log.Debugf("Submitting complete for check [%s]", checkID)
	batcher.input <- submitComplete{
		checkID: checkID,
	}
}

// Shutdown shuts down the batcher
func (batcher AsynchronousBatcher) Shutdown() {
	batcher.input <- submitShutdown{}
}
