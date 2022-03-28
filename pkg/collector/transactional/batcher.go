package transactional

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"time"
)

// MakeCheckInstanceBatcher initializes the batcher instance for a check instance
func MakeCheckInstanceBatcher(checkId check.ID, hostname, agentName string, maxCapacity int, flushInterval time.Duration) *CheckTransactionalBatcher {
	checkFlushInterval := time.NewTicker(flushInterval)
	ctb := &CheckTransactionalBatcher{
		BatcherBase:   batcher.MakeBatcherBase(hostname, agentName, maxCapacity),
		CheckInstance: checkId,
		flushTicker:   checkFlushInterval,
	}

	go ctb.listenForFlushTicker()

	return ctb
}

// CheckTransactionalBatcher is a instance of a batcher for a specific check instance
type CheckTransactionalBatcher struct {
	batcher.BatcherBase
	CheckInstance check.ID
	builder       TransactionalBatchBuilder
	flushTicker   *time.Ticker
}

// GetCheckInstance returns the check instance for this batcher
func (ctb *CheckTransactionalBatcher) GetCheckInstance() check.ID {
	return ctb.CheckInstance
}

// FlushCheckInstance submits complete for this check instance
func (ctb *CheckTransactionalBatcher) FlushCheckInstance() {
	ctb.SubmitComplete(ctb.CheckInstance)
}

// listenForFlushTicker waits for messages on the ticker channel and submits a flush for this check
func (ctb *CheckTransactionalBatcher) listenForFlushTicker() {
	for _ = range ctb.flushTicker.C {
		ctb.FlushCheckInstance()
	}
}

// submitPayload submits the payload to the forwarder
func (ctb *CheckTransactionalBatcher) submitPayload(payload map[string]interface{}) {

}

// mapStateToPayload submits the payload to the forwarder
func (ctb *CheckTransactionalBatcher) mapStateToPayload(state *batcher.CheckInstanceBatchState) map[string]interface{} {
	// Create the topologies
	topologies := make([]topology.Topology, 0)
	if state.Topology != nil {
		topologies = append(topologies, *state.Topology)
	}

	// Create the healthData payload
	healthData := make([]health.Health, 0)
	for _, healthRecord := range state.Health {
		healthData = append(healthData, healthRecord)
	}

	// Create the rawMetricData payload
	rawMetrics := make([]interface{}, 0)
	if state.Metrics != nil {
		for _, metric := range *state.Metrics {
			rawMetrics = append(rawMetrics, metric.ConvertToIntakeMetric())
		}
	}

	payload := map[string]interface{}{
		"internalHostname": ctb.Hostname,
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

	return payload
}

// Start starts the transactional batcher
func (ctb *CheckTransactionalBatcher) Start() {
	for {
		var state *batcher.CheckInstanceBatchState
		s := <-ctb.Input
		switch submission := s.(type) {
		case batcher.SubmitComponent:
			state = ctb.builder.AddComponent(submission.Instance, submission.Component)
		case batcher.SubmitRelation:
			state = ctb.builder.AddRelation(submission.Instance, submission.Relation)
		case batcher.SubmitStartSnapshot:
			state = ctb.builder.StartSnapshot(submission.Instance)
		case batcher.SubmitStopSnapshot:
			state = ctb.builder.StopSnapshot(submission.Instance)
		case batcher.SubmitHealthCheckData:
			state = ctb.builder.AddHealthCheckData(submission.Stream, submission.Data)
		case batcher.SubmitHealthStartSnapshot:
			state = ctb.builder.HealthStartSnapshot(submission.Stream, submission.IntervalSeconds, submission.ExpirySeconds)
		case batcher.SubmitHealthStopSnapshot:
			state = ctb.builder.HealthStopSnapshot(submission.Stream)
		case batcher.SubmitRawMetricsData:
			state = ctb.builder.AddRawMetricsData(submission.RawMetric)
		case batcher.SubmitComplete:
			state = ctb.builder.FlushIfDataProduced()
		case batcher.SubmitShutdown:
			return
		default:
			panic(fmt.Sprint("Unknown submission type"))
		}

		ctb.mapStateToPayload(state)
	}
}

// Stop stops the transactional batcher
func (ctb *CheckTransactionalBatcher) Stop() {
	ctb.flushTicker.Stop()
	ctb.Input <- batcher.SubmitShutdown{}
}
