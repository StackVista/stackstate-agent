package handler

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	checkState "github.com/DataDog/datadog-agent/pkg/collector/check/state"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/metrics/event"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/check"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/telemetry"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionmanager"
	"sync"
)

// TransactionalCheckHandler provides an interface between the Agent Check and the transactional components.
type TransactionalCheckHandler struct {
	CheckHandlerBase
	shutdownChannel           chan bool
	transactionChannel        chan StartTransaction
	currentTransactionChannel chan interface{}
	currentTransaction        string
	transactionalBatcher      transactionbatcher.TransactionalBatcher
	transactionalManager      transactionmanager.TransactionManager
	mux                       sync.RWMutex
}

// NewTransactionalCheckHandler creates a new check handler for a given check, check loader and configuration
func NewTransactionalCheckHandler(stateManager checkState.CheckStateAPI, transactionalBatcher transactionbatcher.TransactionalBatcher, transactionalManager transactionmanager.TransactionManager, check CheckIdentifier, config, initConfig integration.Data) CheckHandler {
	ch := &TransactionalCheckHandler{
		CheckHandlerBase: CheckHandlerBase{
			CheckIdentifier: check,
			config:          config,
			initConfig:      initConfig,
			stateManager:    stateManager,
		},
		shutdownChannel:           make(chan bool, 1),
		transactionChannel:        make(chan StartTransaction, 1),
		currentTransactionChannel: make(chan interface{}, 100),
		transactionalBatcher:      transactionalBatcher,
		transactionalManager:      transactionalManager,
	}

	go ch.Start()

	return ch
}

// Name returns TransactionalCheckHandler for the transactional check handler
func (ch *TransactionalCheckHandler) Name() string {
	return "TransactionalCheckHandler"
}

// Start starts the check handler to "listen" and handle check transactions or shutdown
func (ch *TransactionalCheckHandler) Start() {
txReceiverHandler:
	for {
		select {
		case transaction := <-ch.transactionChannel:
			// set the current transaction
			log.Infof("starting transaction for check %s: %s", transaction.CheckID, transaction.TransactionID)
			ch.mux.Lock()
			ch.currentTransaction = transaction.TransactionID
			ch.mux.Unlock()

			// create a new transaction in the transaction transactionalManager and wait for responses
			ch.transactionalManager.StartTransaction(check.CheckID(ch.ID()), ch.GetCurrentTransaction(), ch.currentTransactionChannel)

			// create a new batch transaction in the transaction transactionalBatcher
			ch.transactionalBatcher.StartTransaction(check.CheckID(ch.ID()), ch.GetCurrentTransaction())

			// this is a blocking function. Will continue when a transaction succeeds, fails or times out making it
			// ready to handle the next transaction in the ch.transactionChannel.
			ch.handleCurrentTransaction(ch.currentTransactionChannel)

		case <-ch.shutdownChannel:
			log.Debug("Shutting down check handler. Closing transaction channels.")
			// try closing currentTransactionChannel if the transaction is still in progress
			safeCloseTransactionChannel(ch.currentTransactionChannel)
			close(ch.transactionChannel)
			break txReceiverHandler
		}
	}
}

// Stop stops the check handler txReceiverHandler
func (ch *TransactionalCheckHandler) Stop() {
	ch.shutdownChannel <- true
}

// GetCurrentTransaction returns ch.currentTransaction with a mutex read lock
func (ch *TransactionalCheckHandler) GetCurrentTransaction() string {
	ch.mux.RLock()
	currentTransaction := ch.currentTransaction
	ch.mux.RUnlock()
	return currentTransaction
}

// handleCurrentTransaction handles the current transaction
func (ch *TransactionalCheckHandler) handleCurrentTransaction(txChan chan interface{}) {
	logPrefix := fmt.Sprintf("Check: %s, Transaction: %s.", ch.ID(), ch.GetCurrentTransaction())
currentTxHandler:
	for {
		select {
		case tx := <-txChan:
			switch msg := tx.(type) {
			// Transactional Operations for the current transaction
			case StopTransaction:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Stopping current transaction", logPrefix)
				}
				ch.transactionalBatcher.SubmitCompleteTransaction(check.CheckID(ch.ID()), ch.GetCurrentTransaction())

			case DiscardTransaction:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Discarding current transaction", logPrefix)
				}
				// trigger failed transaction
				ch.transactionalManager.DiscardTransaction(ch.GetCurrentTransaction(), msg.Reason)

			case SubmitSetTransactionState:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting set transaction state: %s -> %s", logPrefix, msg.Key, msg.State)
				}
				ch.transactionalManager.SetState(ch.GetCurrentTransaction(), msg.Key, msg.State)

			// Topology Operations for the current transaction
			case SubmitStartSnapshot:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting topology start snapshot for instance: %s", logPrefix, msg.Instance.GoString())
				}
				ch.transactionalBatcher.SubmitStartSnapshot(check.CheckID(ch.ID()), ch.GetCurrentTransaction(), msg.Instance)
			case SubmitComponent:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting topology component for instance: %s. %s", logPrefix, msg.Instance.GoString(), msg.Component.JSONString())
				}
				ch.transactionalBatcher.SubmitComponent(check.CheckID(ch.ID()), ch.GetCurrentTransaction(), msg.Instance, msg.Component)
			case SubmitRelation:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting topology relation for instance: %s. %s", logPrefix, msg.Instance.GoString(), msg.Relation.JSONString())
				}
				ch.transactionalBatcher.SubmitRelation(check.CheckID(ch.ID()), ch.GetCurrentTransaction(), msg.Instance, msg.Relation)
			case SubmitStopSnapshot:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting topology stop snapshot for instance: %s", logPrefix, msg.Instance.GoString())
				}
				ch.transactionalBatcher.SubmitStopSnapshot(check.CheckID(ch.ID()), ch.GetCurrentTransaction(), msg.Instance)
			case SubmitDelete:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting topology delete for instance: %s, externalID: %s", logPrefix, msg.Instance.GoString(), msg.TopologyElementID)
				}
				ch.transactionalBatcher.SubmitDelete(check.CheckID(ch.ID()), ch.GetCurrentTransaction(), msg.Instance, msg.TopologyElementID)

			// Health Operations for the current transaction
			case SubmitHealthStartSnapshot:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting health start snapshot for stream: %s with interval %ds and expiry %ds", logPrefix, msg.Stream.GoString(), msg.IntervalSeconds, msg.ExpirySeconds)
				}

				ch.transactionalBatcher.SubmitHealthStartSnapshot(check.CheckID(ch.ID()), ch.GetCurrentTransaction(), msg.Stream,
					msg.IntervalSeconds, msg.ExpirySeconds)

			case SubmitHealthCheckData:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting health check data for stream %s. %s", logPrefix, msg.Stream.GoString(), msg.Data.JSONString())
				}
				ch.transactionalBatcher.SubmitHealthCheckData(check.CheckID(ch.ID()), ch.GetCurrentTransaction(), msg.Stream, msg.Data)

			case SubmitHealthStopSnapshot:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting health stop snapshot for stream: %s", logPrefix, msg.Stream.GoString())
				}

				ch.transactionalBatcher.SubmitHealthStopSnapshot(check.CheckID(ch.ID()), ch.GetCurrentTransaction(), msg.Stream)

			// Telemetry Operations for the current transaction
			case SubmitRawMetric:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting raw metric: %s", logPrefix, msg.Value.JSONString())
				}

				ch.transactionalBatcher.SubmitRawMetricsData(check.CheckID(ch.ID()), ch.GetCurrentTransaction(), msg.Value)
			case SubmitEvent:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting event: %s", logPrefix, msg.Event.String())
				}

				ch.transactionalBatcher.SubmitEvent(check.CheckID(ch.ID()), ch.GetCurrentTransaction(), ConvertToStsEvent(msg.Event))
			// Lifecycle operations for the current transaction
			case SubmitComplete:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting complete for check run", logPrefix)
				}

			// Notifications from the transaction transactionalManager
			case transactionmanager.DiscardTransaction:
				if msg.TransactionID != ch.GetCurrentTransaction() {
					_ = log.Warnf("Attempting to discard transaction that is not the current transaction for this"+
						"check. Current transaction: %s, discarded transaction: %s",
						ch.GetCurrentTransaction(), msg.TransactionID)
					continue
				}

				log.Debugf("Discarding failed transaction %s for check %s. Reason %s", msg.TransactionID, ch.ID(),
					msg.Reason)

				// empty transactionalBatcher state
				ch.transactionalBatcher.SubmitClearState(check.CheckID(ch.ID()))

				// clear current transaction
				ch.clearCurrentTransaction()

				break currentTxHandler

			case transactionmanager.EvictedTransaction:
				if msg.TransactionID != ch.GetCurrentTransaction() {
					_ = log.Warnf("Attempting to evict transaction that is not the current transaction for this"+
						"check. Current transaction: %s, evicted transaction: %s",
						ch.GetCurrentTransaction(), msg.TransactionID)
					continue
				}

				log.Debugf("Evicted failed transaction %s for check %s", msg.TransactionID, ch.ID())

				// empty transactionalBatcher state
				ch.transactionalBatcher.SubmitClearState(check.CheckID(ch.ID()))

				// clear current transaction
				ch.clearCurrentTransaction()

				break currentTxHandler

			case transactionmanager.CompleteTransaction:
				if msg.TransactionID != ch.GetCurrentTransaction() {
					_ = log.Warnf("Attempting to complete transaction that is not the current transaction for this"+
						"check. Current transaction: %s, completed transaction: %s",
						ch.GetCurrentTransaction(), msg.TransactionID)
					continue
				}

				log.Infof("Completing transaction: %s for check %s", msg.TransactionID, ch.ID())

				if msg.State != nil {
					log.Debugf("Committing state for transaction: %s for check %s: %s", msg.TransactionID, ch.ID(),
						msg.State)
					err := ch.stateManager.SetState(msg.State.Key, msg.State.State)
					if err != nil {
						errorReason := fmt.Sprintf("Error while updating state for transaction: %s for check %s: %s. %s",
							msg.TransactionID, ch.ID(), msg.State, err)
						_ = log.Error(errorReason)
						txChan <- transactionmanager.DiscardTransaction{TransactionID: msg.TransactionID, Reason: errorReason}
					}

					log.Debugf("Successfully committed state for transaction: %s for check %s: %s", msg.TransactionID, ch.ID(),
						msg.State)
				}

				// clear current transaction
				ch.clearCurrentTransaction()

				break currentTxHandler
			}
		}
	}
}

func (ch *TransactionalCheckHandler) clearCurrentTransaction() {
	ch.mux.Lock()
	ch.currentTransaction = ""
	ch.mux.Unlock()
}

//	Convert datadog event type to STS event type. These are pretty much equal but we want to start
//
// decoupling from the datadog api, this is the way we do. We copied their datastructures that we are compatible with,
// brought them to the receiver-go-client and transform in our agent fork.
func ConvertToStsEvent(event event.Event) telemetry.Event {
	//  Convert datadog event type to STS event type. These are pretty much equal but we want to start
	// decoupling from the datadog api, this is the way we do. We copied their datastructures that we are compatible with,
	// brought them to the receiver-go-client and transform in our agent fork.
	var stsEventCtx *telemetry.EventContext

	if event.EventContext != nil {
		var stsSourceLinks = make([]telemetry.SourceLink, 0, 0)
		for _, v := range event.EventContext.SourceLinks {
			stsSourceLinks = append(stsSourceLinks, telemetry.SourceLink{
				Title: v.Title,
				URL:   v.URL,
			})
		}

		stsEventCtx = &telemetry.EventContext{
			SourceIdentifier:   event.EventContext.SourceIdentifier,
			ElementIdentifiers: event.EventContext.ElementIdentifiers,
			Source:             event.EventContext.Source,
			Category:           event.EventContext.Category,
			Data:               event.EventContext.Data,
			SourceLinks:        stsSourceLinks,
		}
	}

	return telemetry.Event{
		Title:          event.Title,
		Text:           event.Text,
		Ts:             event.Ts,
		Priority:       telemetry.EventPriority(event.Priority),
		Host:           event.Host,
		Tags:           event.Tags,
		AlertType:      telemetry.EventAlertType(event.AlertType),
		AggregationKey: event.AggregationKey,
		SourceTypeName: event.SourceTypeName,
		EventType:      event.EventType,
		OriginID:       "",
		K8sOriginID:    "",
		Cardinality:    event.Cardinality,
		EventContext:   stsEventCtx,
	}
}

// safeCloseTransactionChannel closes the tx channel that can potentially already be closed. It handles the panic and does a no-op.
func safeCloseTransactionChannel(ch chan interface{}) {
	defer func() {
		if recover() != nil {
		}
	}()

	close(ch) // panic if ch is closed
}
