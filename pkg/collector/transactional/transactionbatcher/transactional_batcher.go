package transactionbatcher

import (
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional/transactionforwarder"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional/transactionmanager"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/metrics/event"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/google/uuid"
	"sync"
)

var (
	batcherInstance TransactionalBatcher
	batcherInit     *sync.Once
)

func init() {
	batcherInit = new(sync.Once)
}

// InitTransactionalBatcher initializes the global transactional transactionbatcher Instance
func InitTransactionalBatcher(hostname, agentName string, maxCapacity int) {
	batcherInit.Do(func() {
		batcherInstance = newTransactionalBatcher(hostname, agentName, maxCapacity)
	})
}

// GetTransactionalBatcher returns a handle on the global transactionbatcher Instance
func GetTransactionalBatcher() TransactionalBatcher {
	return batcherInstance
}

// NewMockTransactionalBatcher initializes the global transactionbatcher with a mock version, intended for testing
func NewMockTransactionalBatcher() *MockTransactionalBatcher {
	batcherInit.Do(func() {
		batcherInstance = newMockTransactionalBatcher()
	})
	return batcherInstance.(*MockTransactionalBatcher)
}

// newTransactionalBatcher returns an instance of the transactionalBatcher and starts listening for submissions
func newTransactionalBatcher(hostname, agentName string, maxCapacity int) *transactionalBatcher {
	ctb := &transactionalBatcher{
		Hostname:    hostname,
		agentName:   agentName,
		Input:       make(chan interface{}, maxCapacity),
		builder:     NewTransactionalBatchBuilder(maxCapacity),
		maxCapacity: maxCapacity,
	}

	go ctb.Start()

	return ctb
}

// transactionalBatcher is a instance of a transactionbatcher for a specific check instance
type transactionalBatcher struct {
	Hostname, agentName string
	Input               chan interface{}
	builder             TransactionBatchBuilder
	maxCapacity         int
}

// Start starts the transactional transactionbatcher
func (ctb *transactionalBatcher) Start() {
BatcherReceiver:
	for {
		select {
		case s := <-ctb.Input:
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
			case SubmitEvent:
				ctb.SubmitState(ctb.builder.AddEvent(submission.CheckID, submission.TransactionID, submission.Event))
			case StartTransaction:
				ctb.SubmitState(ctb.builder.StartTransaction(submission.CheckID, submission.TransactionID))
			case SubmitCompleteTransaction:
				ctb.SubmitState(ctb.builder.MarkTransactionComplete(submission.CheckID, submission.TransactionID))
			case SubmitClearState:
				ctb.SubmitState(ctb.builder.ClearState(submission.CheckID))
			case SubmitShutdown:
				break BatcherReceiver
			default:
				panic(fmt.Sprint("Unknown submission type"))
			}
		}
	}
}

// Stop stops the transactional transactionbatcher
func (ctb *transactionalBatcher) Stop() {
	ctb.Input <- SubmitShutdown{}

	// reset the batcher init to re-init the batcher
	batcherInit = new(sync.Once)
}

// SubmitState submits the transactional check instance batch state and commits an action for this payload
func (ctb *transactionalBatcher) SubmitState(states TransactionCheckInstanceBatchStates) {
	if len(states) > 0 {
		data := ctb.mapStateToPayload(states)
		payload, err := ctb.marshallPayload(data)
		if err != nil {
			// discard all the transactions in the transactionbatcher states
			for _, state := range states {
				transactionmanager.GetTransactionManager().DiscardTransaction(state.Transaction.TransactionID, fmt.Sprintf("Marshall error in payload: %v", data))
			}
		}

		// Catering for the edge case where the data produced for a given check and transaction was published in a
		// previous payload. There is a very likely possibility that this payload is still in the forwarder being sent
		// to StackState. The best way to guarantee that we're not prematurely marking a transaction as complete is to
		// forward an empty payload to the forwarder, thereby committing one final (fake) action, acknowledging it in the
		// forwarder and then marking the transaction as complete.
		// In the event of the same happening, but ending up in a state where there is already data in the payload, we
		// don't have to do anything special. The action will be committed and the transaction completed as part of that
		// payload.
		progressTransactions := false
		emptyPayload := data.EqualDataPayload(transactional.NewIntakePayload())
		if emptyPayload {
			for _, state := range states {
				if state.Transaction.CompletedTransaction {
					progressTransactions = true
				}
			}
		}

		// if this is an empty payload, and we have no transactions to mark, return
		if !progressTransactions && emptyPayload {
			return
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

		log.Debugf("Marshalled payload for transactions: %v, payload: %s", transactionPayloadMap, string(payload))

		ctb.submitPayload(payload, transactionPayloadMap)
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
	allEvents := &event.IntakeEvents{}
	for _, state := range states {
		if state.Topology != nil {
			intake.Topologies = append(intake.Topologies, *state.Topology)
		}

		for _, healthRecord := range state.Health {
			intake.Health = append(intake.Health, healthRecord)
		}

		if state.Metrics != nil {
			for _, metric := range state.Metrics.Values {
				intake.Metrics = append(intake.Metrics, metric.ConvertToIntakeMetric())
			}
		}

		if state.Events != nil {
			allEvents.Events = append(allEvents.Events, state.Events.Events...)
		}
	}

	intake.Events = allEvents.IntakeFormat()

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

		log.Debug("Flushing the following event data:")
		for _, e := range intake.Events {
			log.Debugf("%v", e)
		}
	}

	return intake
}
