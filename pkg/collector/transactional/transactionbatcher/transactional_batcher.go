package transactionbatcher

import (
	"encoding/json"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionforwarder"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/config"
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
func InitTransactionalBatcher(hostname, agentName string, maxCapacity int, flushInterval time.Duration) {
	batcherInit.Do(func() {
		batcherInstance = newTransactionalBatcher(hostname, agentName, maxCapacity, flushInterval)
	})
}

func newTransactionalBatcher(hostname, agentName string, maxCapacity int, flushInterval time.Duration) *transactionalBatcher {
	checkFlushInterval := time.NewTicker(flushInterval)
	ctb := &transactionalBatcher{
		Hostname:    hostname,
		agentName:   agentName,
		Input:       make(chan interface{}, maxCapacity),
		builder:     NewTransactionalBatchBuilder(maxCapacity),
		flushTicker: checkFlushInterval,
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
	maxCapacity         int
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
		case SubmitCompleteTransaction:
			ctb.SubmitState(ctb.builder.MarkTransactionComplete(submission.CheckID, submission.TransactionID))
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

// listenForFlushTicker waits for messages on the ticker channel and submits a flush for this check
func (ctb *transactionalBatcher) listenForFlushTicker() {
	for _ = range ctb.flushTicker.C {
		ctb.SubmitState(ctb.builder.Flush())
	}
}

// submitPayload submits the payload to the forwarder
func (ctb *transactionalBatcher) submitPayload(payload []byte, transactionPayloadMap map[string]transactional.PayloadTransaction) {
	transactionforwarder.GetTransactionalForwarder().SubmitTransactionalIntake(transactionforwarder.TransactionalPayload{
		Body:                 payload,
		Path:                 transactional.IntakePath,
		TransactionActionMap: transactionPayloadMap,
	})
}

// marshallPayload submits the payload to the forwarder
func (ctb *transactionalBatcher) marshallPayload(intake transactional.IntakePayload) ([]byte, error) {
	payload, err := json.Marshal(intake)
	if err != nil {
		return nil, fmt.Errorf("could not serialize intake payload: %s", err)
	}

	return payload, nil
}

// mapStateToPayload submits the payload to the forwarder
func (ctb *transactionalBatcher) mapStateToPayload(states TransactionCheckInstanceBatchStates) transactional.IntakePayload {
	intake := transactional.NewIntakePayload()
	intake.InternalHostname = ctb.Hostname

	// Create the topologies payload
	for _, state := range states {
		if state.Topology != nil {
			intake.Topologies = append(intake.Topologies, *state.Topology)
		}
	}

	// Create the health payload
	for _, state := range states {
		for _, healthRecord := range state.Health {
			intake.Health = append(intake.Health, healthRecord)
		}
	}

	// Create the metric payload
	for _, state := range states {
		if state.Metrics != nil {
			for _, metric := range state.Metrics.Values {
				intake.Metrics = append(intake.Metrics, metric.ConvertToIntakeMetric())
			}
		}
	}

	// For debug purposes print out all topologies payload
	if config.Datadog.GetBool("log_payloads") {
		log.Debug("Flushing the following topologies:")
		for _, topo := range intake.Topologies {
			log.Debugf("%v", topo)
		}

		log.Debug("Flushing the following health data:")
		for _, h := range intake.Health {
			log.Debugf("%v", h)
		}

		log.Debug("Flushing the following raw metric data:")
		for _, rawMetric := range intake.Metrics {
			log.Debugf("%v", rawMetric)
		}
	}

	return intake
}
