package transactionbatcher

import (
	"encoding/json"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionalforwarder"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/google/uuid"
	"sync"
	"time"
)

var (
	batcherInstance TransactionalBatcher
	batcherInit     sync.Once
)

// InitTransactionalBatcher initializes the global transactional transactionbatcher Instance
func InitTransactionalBatcher(hostname, agentName string, maxCapacity int,
	flushInterval time.Duration, forwarder transactionalforwarder.TransactionalForwarder) {
	batcherInit.Do(func() {
		batcherInstance = newTransactionalBatcher(hostname, agentName, maxCapacity, flushInterval, forwarder)
	})
}

func newTransactionalBatcher(hostname, agentName string, maxCapacity int,
	flushInterval time.Duration, forwarder transactionalforwarder.TransactionalForwarder) *transactionalBatcher {
	checkFlushInterval := time.NewTicker(flushInterval)
	ctb := &transactionalBatcher{
		Hostname:    hostname,
		agentName:   agentName,
		Input:       make(chan interface{}, maxCapacity),
		builder:     NewTransactionalBatchBuilder(maxCapacity),
		flushTicker: checkFlushInterval,
		Forwarder:   forwarder,
		maxCapacity: maxCapacity,
	}

	go ctb.Start()
	go ctb.listenForFlushTicker()

	return ctb
}

// GetTransactionalBatcher returns a handle on the global transactionbatcher Instance
func GetTransactionalBatcher() TransactionalBatcher {
	return batcherInstance
}

// NewMockTransactionalBatcher initializes the global transactionbatcher with a mock version, intended for testing
func NewMockTransactionalBatcher() MockTransactionalBatcher {
	batcher := createMockTransactionalBatcher()
	batcherInstance = batcher
	return batcher
}

// transactionalBatcher is a instance of a transactionbatcher for a specific check instance
type transactionalBatcher struct {
	Hostname, agentName string
	Input               chan interface{}
	builder             TransactionBatchBuilder
	flushTicker         *time.Ticker
	Forwarder           transactionalforwarder.TransactionalForwarder
	maxCapacity         int
}

// listenForFlushTicker waits for messages on the ticker channel and submits a flush for this check
func (ctb *transactionalBatcher) listenForFlushTicker() {
	for _ = range ctb.flushTicker.C {
		ctb.SubmitState(ctb.builder.Flush())
	}
}

// submitPayload submits the payload to the forwarder
func (ctb *transactionalBatcher) submitPayload(payload []byte, transactionPayloadMap map[string]transactional.PayloadTransaction) {
	ctb.Forwarder.SubmitTransactionalIntake(transactionalforwarder.TransactionalPayload{
		Payload:              payload,
		TransactionActionMap: transactionPayloadMap,
	})
}

// marshallPayload submits the payload to the forwarder
func (ctb *transactionalBatcher) marshallPayload(data map[string]interface{}) ([]byte, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("could not serialize v1 payload: %s", err)
	}

	return payload, nil
}

// mapStateToPayload submits the payload to the forwarder
func (ctb *transactionalBatcher) mapStateToPayload(states TransactionCheckInstanceBatchStates) map[string]interface{} {
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

// Start starts the transactional transactionbatcher
func (ctb *transactionalBatcher) Start() {
	for {
		s := <-ctb.Input
		switch submission := s.(type) {
		case SubmitComponent:
			ctb.SubmitState(ctb.builder.AddComponent(submission.CheckID, submission.TransactionID, submission.Instance, submission.Component))
		case SubmitRelation:
			ctb.SubmitState(ctb.builder.AddRelation(submission.CheckID, submission.TransactionID, submission.Instance, submission.Relation))
		case SubmitStartSnapshot:
			ctb.SubmitState(ctb.builder.TopologyStartSnapshot(submission.CheckID, submission.TransactionID, submission.Instance))
		case SubmitStopSnapshot:
			ctb.SubmitState(ctb.builder.TopologyStopSnapshot(submission.CheckID, submission.TransactionID, submission.Instance))
		case SubmitDelete:
			ctb.SubmitState(ctb.builder.Delete(submission.CheckID, submission.TransactionID, submission.Instance, submission.DeleteID))
		case SubmitHealthCheckData:
			ctb.SubmitState(ctb.builder.AddHealthCheckData(submission.CheckID, submission.TransactionID, submission.Stream, submission.Data))
		case SubmitHealthStartSnapshot:
			ctb.SubmitState(ctb.builder.HealthStartSnapshot(submission.CheckID, submission.TransactionID, submission.Stream, submission.IntervalSeconds, submission.ExpirySeconds))
		case SubmitHealthStopSnapshot:
			ctb.SubmitState(ctb.builder.HealthStopSnapshot(submission.CheckID, submission.TransactionID, submission.Stream))
		case SubmitRawMetricsData:
			ctb.SubmitState(ctb.builder.AddRawMetricsData(submission.CheckID, submission.TransactionID, submission.RawMetric))
		case SubmitComplete:
			ctb.SubmitState(ctb.builder.FlushOnComplete(submission.CheckID))
		case SubmitShutdown:
			return
		default:
			panic(fmt.Sprint("Unknown submission type"))
		}
	}
}

// SubmitState submits the transactional check instance batch state and commits an action for this payload
func (ctb *transactionalBatcher) SubmitState(states TransactionCheckInstanceBatchStates) {
	if states != nil {

		data := ctb.mapStateToPayload(states)
		payload, err := ctb.marshallPayload(data)
		if err != nil {
			// rollback all the transactions in the transactionbatcher states
			for _, state := range states {
				transactionmanager.GetTransactionManager().RollbackTransaction(state.Transaction.TransactionID, fmt.Sprintf("Marshall error in payload: %v", data))
			}
		}
		// create a transaction -> action map that can be used to acknowledge / reject actions
		transactionPayloadMap := make(map[string]transactional.PayloadTransaction, len(states))
		for _, state := range states {
			actionID := uuid.New().String()
			transactionPayloadMap[state.Transaction.TransactionID] = transactional.PayloadTransaction{
				ActionID:             actionID,
				CompletedTransaction: state.Transaction.CompletedTransaction,
			}
			// commit an action for each of the transactions in this transactionbatcher state
			transactionmanager.GetTransactionManager().CommitAction(state.Transaction.TransactionID, actionID)
		}

		ctb.submitPayload(payload, transactionPayloadMap)
	}
}

// SubmitComponent submits a component to the batch
func (ctb *transactionalBatcher) SubmitComponent(checkID check.ID, transactionID string, instance topology.Instance, component topology.Component) {
	ctb.Input <- SubmitComponent{
		CheckID:       checkID,
		TransactionID: transactionID,
		Instance:      instance,
		Component:     component,
	}
}

// SubmitRelation submits a relation to the batch
func (ctb *transactionalBatcher) SubmitRelation(checkID check.ID, transactionID string, instance topology.Instance, relation topology.Relation) {
	ctb.Input <- SubmitRelation{
		CheckID:       checkID,
		TransactionID: transactionID,
		Instance:      instance,
		Relation:      relation,
	}
}

// SubmitStartSnapshot submits start of a snapshot
func (ctb *transactionalBatcher) SubmitStartSnapshot(checkID check.ID, transactionID string, instance topology.Instance) {
	ctb.Input <- SubmitStartSnapshot{
		CheckID:       checkID,
		TransactionID: transactionID,
		Instance:      instance,
	}
}

// SubmitStopSnapshot submits a stop of a snapshot. This always causes a flush of the data downstream
func (ctb *transactionalBatcher) SubmitStopSnapshot(checkID check.ID, transactionID string, instance topology.Instance) {
	ctb.Input <- SubmitStopSnapshot{
		CheckID:       checkID,
		TransactionID: transactionID,
		Instance:      instance,
	}
}

// SubmitDelete submits a deletion of topology element.
func (ctb *transactionalBatcher) SubmitDelete(checkID check.ID, transactionID string, instance topology.Instance, topologyElementID string) {
	ctb.Input <- SubmitDelete{
		CheckID:       checkID,
		TransactionID: transactionID,
		Instance:      instance,
		DeleteID:      topologyElementID,
	}
}

// SubmitHealthCheckData submits a Health check data record to the batch
func (ctb *transactionalBatcher) SubmitHealthCheckData(checkID check.ID, transactionID string, stream health.Stream, data health.CheckData) {
	log.Debugf("Submitting Health check data for check [%s] stream [%s]: %s", checkID, stream.GoString(), data.JSONString())
	ctb.Input <- SubmitHealthCheckData{
		CheckID:       checkID,
		TransactionID: transactionID,
		Stream:        stream,
		Data:          data,
	}
}

// SubmitHealthStartSnapshot submits start of a Health snapshot
func (ctb *transactionalBatcher) SubmitHealthStartSnapshot(checkID check.ID, transactionID string, stream health.Stream, intervalSeconds int, expirySeconds int) {
	ctb.Input <- SubmitHealthStartSnapshot{
		CheckID:         checkID,
		TransactionID:   transactionID,
		Stream:          stream,
		IntervalSeconds: intervalSeconds,
		ExpirySeconds:   expirySeconds,
	}
}

// SubmitHealthStopSnapshot submits a stop of a Health snapshot. This always causes a flush of the data downstream
func (ctb *transactionalBatcher) SubmitHealthStopSnapshot(checkID check.ID, transactionID string, stream health.Stream) {
	ctb.Input <- SubmitHealthStopSnapshot{
		CheckID:       checkID,
		TransactionID: transactionID,
		Stream:        stream,
	}
}

// SubmitRawMetricsData submits a raw metrics data record to the batch
func (ctb *transactionalBatcher) SubmitRawMetricsData(checkID check.ID, transactionID string, rawMetric telemetry.RawMetrics) {
	if rawMetric.HostName == "" {
		rawMetric.HostName = ctb.Hostname
	}

	ctb.Input <- SubmitRawMetricsData{
		CheckID:       checkID,
		TransactionID: transactionID,
		RawMetric:     rawMetric,
	}
}

// SubmitComplete signals completion of a check. May trigger a flush only if the check produced data
func (ctb *transactionalBatcher) SubmitComplete(checkID check.ID) {
	log.Debugf("Submitting complete for check [%s]", checkID)
	ctb.Input <- SubmitComplete{
		CheckID: checkID,
	}
}

// Shutdown shuts down the transactionbatcher
func (ctb *transactionalBatcher) Shutdown() {
	ctb.Input <- SubmitShutdown{}
}

// Stop stops the transactional transactionbatcher
func (ctb *transactionalBatcher) Stop() {
	ctb.flushTicker.Stop()
	ctb.Input <- SubmitShutdown{}
}
