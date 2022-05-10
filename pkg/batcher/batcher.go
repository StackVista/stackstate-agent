package batcher

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/serializer"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"sync"
)

var (
	batcherInstance Batcher
	batcherInit     sync.Once
)

// InitBatcher initializes the global transactionbatcher Instance
func InitBatcher(serializer serializer.AgentV1Serializer, hostname, agentName string, maxCapacity int) {
	batcherInit.Do(func() {
		batcherInstance = newAsynchronousBatcher(serializer, hostname, agentName, maxCapacity)
	})
}

func newAsynchronousBatcher(serializer serializer.AgentV1Serializer, hostname, agentName string, maxCapacity int) AsynchronousBatcher {
	batcher := AsynchronousBatcher{
		BatcherBase: MakeBatcherBase(hostname, agentName, maxCapacity),
		builder:     NewBatchBuilder(maxCapacity),
		serializer:  serializer,
	}
	go batcher.run()
	return batcher
}

// GetBatcher returns a handle on the global transactionbatcher Instance
func GetBatcher() Batcher {
	return batcherInstance
}

// NewMockBatcher initializes the global transactionbatcher with a mock version, intended for testing
func NewMockBatcher() MockBatcher {
	batcher := createMockBatcher()
	batcherInstance = batcher
	return batcher
}

// AsynchronousBatcher is the implementation of the transactionbatcher. Works asynchronous. Publishes data to the serializer
type AsynchronousBatcher struct {
	BatcherBase
	builder    BatchBuilder
	serializer serializer.AgentV1Serializer
}

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
			"internalHostname": batcher.Hostname,
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
		s := <-batcher.Input
		switch submission := s.(type) {
		case SubmitComponent:
			batcher.sendState(batcher.builder.AddComponent(submission.CheckID, submission.Instance, submission.Component))
		case SubmitRelation:
			batcher.sendState(batcher.builder.AddRelation(submission.CheckID, submission.Instance, submission.Relation))
		case SubmitStartSnapshot:
			batcher.sendState(batcher.builder.TopologyStartSnapshot(submission.CheckID, submission.Instance))
		case SubmitStopSnapshot:
			batcher.sendState(batcher.builder.TopologyStopSnapshot(submission.CheckID, submission.Instance))
		case SubmitDelete:
			batcher.sendState(batcher.builder.Delete(submission.CheckID, submission.Instance, submission.DeleteID))
		case SubmitHealthCheckData:
			batcher.sendState(batcher.builder.AddHealthCheckData(submission.CheckID, submission.Stream, submission.Data))
		case SubmitHealthStartSnapshot:
			batcher.sendState(batcher.builder.HealthStartSnapshot(submission.CheckID, submission.Stream, submission.IntervalSeconds, submission.ExpirySeconds))
		case SubmitHealthStopSnapshot:
			batcher.sendState(batcher.builder.HealthStopSnapshot(submission.CheckID, submission.Stream))

		case SubmitRawMetricsData:
			batcher.sendState(batcher.builder.AddRawMetricsData(submission.CheckID, submission.RawMetric))

		case SubmitComplete:
			batcher.sendState(batcher.builder.FlushIfDataProduced(submission.CheckID))
		case SubmitShutdown:
			return
		default:
			panic(fmt.Sprint("Unknown submission type"))
		}
	}
}
