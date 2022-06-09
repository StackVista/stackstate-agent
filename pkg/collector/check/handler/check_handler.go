package handler

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	checkState "github.com/StackVista/stackstate-agent/pkg/collector/check/state"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"sync"
)

// CheckReloader is a interface wrapper around the Collector which controls all checks.
type CheckReloader interface {
	ReloadCheck(id check.ID, config, initConfig integration.Data, newSource string) error
}

// CheckIdentifier encapsulates all the functionality needed to describe and configure an agent check.
type CheckIdentifier interface {
	String() string       // provide a printable version of the check name
	ID() check.ID         // provide a unique identifier for every check instance
	ConfigSource() string // return the configuration source of the check
}

// CheckHandler represents a wrapper around an Agent Check that allows us track data produced by an agent check, as well
// as handle transactions for it.
type CheckHandler interface {
	CheckIdentifier
	CheckAPI
	GetConfig() (config, initConfig integration.Data)
	GetCheckReloader() CheckReloader
	Reload()
}

// checkHandler provides an interface between the Agent Check and the transactional components.
type checkHandler struct {
	CheckIdentifier
	CheckReloader
	config, initConfig        integration.Data
	shutdownChannel           chan bool
	transactionChannel        chan StartTransaction
	currentTransactionChannel chan interface{}
	currentTransaction        string
	mux                       sync.RWMutex
}

// Reload uses the Check Reloader (Collector) to reload the check: pkg/collector/collector.go:126
func (ch *checkHandler) Reload() {
	config, initConfig := ch.GetConfig()
	if err := ch.ReloadCheck(ch.ID(), config, initConfig, ch.ConfigSource()); err != nil {
		_ = log.Errorf("could not reload check %s", string(ch.ID()))
	}
}

// NewCheckHandler creates a new check handler for a given check, check loader and configuration
func NewCheckHandler(check check.Check, checkReloader CheckReloader, config, initConfig integration.Data) CheckHandler {
	ch := &checkHandler{
		CheckIdentifier:           check,
		CheckReloader:             checkReloader,
		config:                    config,
		initConfig:                initConfig,
		shutdownChannel:           make(chan bool, 1),
		transactionChannel:        make(chan StartTransaction, 1),
		currentTransactionChannel: make(chan interface{}, 100),
	}

	go ch.Start()

	return ch
}

// Start starts the check handler to "listen" and handle check transactions or shutdown
func (ch *checkHandler) Start() {
txReceiverHandler:
	for {
		select {
		case transaction := <-ch.transactionChannel:
			// set the current transaction
			log.Infof("starting transaction for check %s: %s", transaction.CheckID, transaction.TransactionID)
			ch.mux.Lock()
			ch.currentTransaction = transaction.TransactionID
			ch.mux.Unlock()

			// create a new transaction in the transaction manager and wait for responses
			transactionmanager.GetTransactionManager().StartTransaction(ch.ID(), ch.GetCurrentTransaction(), ch.currentTransactionChannel)

			// create a new batch transaction in the transaction batcher
			transactionbatcher.GetTransactionalBatcher().StartTransaction(ch.ID(), ch.GetCurrentTransaction())

			// this is a blocking function. Will continue when a transaction succeeds, fails or times out making it
			// ready to handle the next transaction in the ch.transactionChannel.
			ch.handleCurrentTransaction(ch.currentTransactionChannel)

		case <-ch.shutdownChannel:
			log.Debug("Shutting down check handler. Closing transaction channels.")
			// try closing currentTransactionChannel if the transaction is still in progress
			safeCloseTransactionChannel(ch.currentTransactionChannel)
			close(ch.transactionChannel)
			break txReceiverHandler
		default:
		}
	}
}

// Stop stops the check handler txReceiverHandler
func (ch *checkHandler) Stop() {
	ch.shutdownChannel <- true
}

// GetCurrentTransaction returns ch.currentTransaction with a mutex read lock
func (ch *checkHandler) GetCurrentTransaction() string {
	ch.mux.RLock()
	currentTransaction := ch.currentTransaction
	ch.mux.RUnlock()
	return currentTransaction
}

// handleCurrentTransaction handles the current transaction
func (ch *checkHandler) handleCurrentTransaction(txChan chan interface{}) {
	logPrefix := fmt.Sprintf("Check: %s, Transaction: %s.", ch.GetCurrentTransaction(), ch.ID())
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
				transactionbatcher.GetTransactionalBatcher().SubmitCompleteTransaction(ch.ID(), ch.GetCurrentTransaction())

			case CancelTransaction:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Cancelling current transaction", logPrefix)
				}
				// empty batcher state
				transactionbatcher.GetTransactionalBatcher().SubmitClearState(ch.ID())
				// trigger failed transaction
				transactionmanager.GetTransactionManager().RollbackTransaction(ch.GetCurrentTransaction(), msg.Reason)

			case SubmitSetTransactionState:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting set transaction state: %s -> %s", logPrefix, msg.Key, msg.State)
				}
				transactionmanager.GetTransactionManager().SetState(ch.GetCurrentTransaction(), msg.Key, msg.State)

			// Topology Operations for the current transaction
			case SubmitStartSnapshot:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting topology start snapshot for instance: %s", logPrefix, msg.Instance.GoString())
				}
				transactionbatcher.GetTransactionalBatcher().SubmitStartSnapshot(ch.ID(), ch.GetCurrentTransaction(), msg.Instance)
			case SubmitComponent:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting topology component for instance: %s. %s", logPrefix, msg.Instance.GoString(), msg.Component.JSONString())
				}
				transactionbatcher.GetTransactionalBatcher().SubmitComponent(ch.ID(), ch.GetCurrentTransaction(), msg.Instance, msg.Component)
			case SubmitRelation:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting topology relation for instance: %s. %s", logPrefix, msg.Instance.GoString(), msg.Relation.JSONString())
				}
				transactionbatcher.GetTransactionalBatcher().SubmitRelation(ch.ID(), ch.GetCurrentTransaction(), msg.Instance, msg.Relation)
			case SubmitStopSnapshot:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting topology stop snapshot for instance: %s", logPrefix, msg.Instance.GoString())
				}
				transactionbatcher.GetTransactionalBatcher().SubmitStopSnapshot(ch.ID(), ch.GetCurrentTransaction(), msg.Instance)
			case SubmitDelete:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting topology delete for instance: %s, externalID: %s", logPrefix, msg.Instance.GoString(), msg.TopologyElementID)
				}
				transactionbatcher.GetTransactionalBatcher().SubmitDelete(ch.ID(), ch.GetCurrentTransaction(), msg.Instance, msg.TopologyElementID)

			// Health Operations for the current transaction
			case SubmitHealthStartSnapshot:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting health start snapshot for stream: %s with interval %ds and expiry %ds", logPrefix, msg.Stream.GoString(), msg.IntervalSeconds, msg.ExpirySeconds)
				}

				transactionbatcher.GetTransactionalBatcher().SubmitHealthStartSnapshot(ch.ID(), ch.GetCurrentTransaction(), msg.Stream,
					msg.IntervalSeconds, msg.ExpirySeconds)

			case SubmitHealthCheckData:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting health check data for stream %s. %s", logPrefix, msg.Stream.GoString(), msg.Data.JSONString())
				}
				transactionbatcher.GetTransactionalBatcher().SubmitHealthCheckData(ch.ID(), ch.GetCurrentTransaction(), msg.Stream, msg.Data)

			case SubmitHealthStopSnapshot:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting health stop snapshot for stream: %s", logPrefix, msg.Stream.GoString())
				}

				transactionbatcher.GetTransactionalBatcher().SubmitHealthStopSnapshot(ch.ID(), ch.GetCurrentTransaction(), msg.Stream)

			// Telemetry Operations for the current transaction
			case SubmitRawMetric:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting raw metric: %s", logPrefix, msg.Value.JSONString())
				}

				transactionbatcher.GetTransactionalBatcher().SubmitRawMetricsData(ch.ID(), ch.GetCurrentTransaction(), msg.Value)

			// Lifecycle operations for the current transaction
			case SubmitComplete:
				if config.Datadog.GetBool("log_payloads") {
					log.Debugf("%s. Submitting complete for check run", logPrefix)
				}

				transactionbatcher.GetTransactionalBatcher().SubmitComplete(ch.ID())

			// Notifications from the transaction manager
			case transactionmanager.RollbackTransaction, transactionmanager.EvictedTransaction:
				_ = log.Warnf("Reloading check %s after failed transaction", ch.ID())
				if err := ch.ReloadCheck(ch.ID(), ch.config, ch.initConfig, ch.ConfigSource()); err != nil {
					_ = log.Errorf("failed to reload check %s: %s", ch.ID(), err)
				}
				break currentTxHandler
			case transactionmanager.CompleteTransaction:
				log.Debugf("Completing transaction: %s for check %s", msg.TransactionID, ch.ID())

				if msg.State != nil {
					log.Debugf("Committing state for transaction: %s for check %s: %s", msg.TransactionID, ch.ID(),
						msg.State)
					err := checkState.GetCheckStateManager().SetState(msg.State.Key, msg.State.State)
					if err != nil {
						errorReason := fmt.Sprintf("Error while updating state for transaction: %s for check %s: %s. %s",
							msg.TransactionID, ch.ID(), msg.State, err)
						_ = log.Error(errorReason)
						txChan <- transactionmanager.RollbackTransaction{TransactionID: msg.TransactionID, Reason: errorReason}
					}

					log.Debugf("Successfully committed state for transaction: %s for check %s: %s", msg.TransactionID, ch.ID(),
						msg.State)
				}
				break currentTxHandler
			}
		}
	}
}

// GetConfig returns the config and the init config of the check
func (ch *checkHandler) GetConfig() (integration.Data, integration.Data) {
	return ch.config, ch.initConfig
}

// GetCheckReloader returns the configured CheckReloader.
func (ch *checkHandler) GetCheckReloader() CheckReloader {
	return ch.CheckReloader
}

// safeCloseTransactionChannel closes the tx channel that can potentially already be closed. It handles the panic and does a no-op.
func safeCloseTransactionChannel(ch chan interface{}) {
	defer func() {
		if recover() != nil {
		}
	}()

	close(ch) // panic if ch is closed
}
