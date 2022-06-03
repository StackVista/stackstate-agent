package handler

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	checkState "github.com/StackVista/stackstate-agent/pkg/collector/check/state"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
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
		CheckIdentifier:    check,
		CheckReloader:      checkReloader,
		config:             config,
		initConfig:         initConfig,
		shutdownChannel:    make(chan bool, 1),
		transactionChannel: make(chan StartTransaction, 1),
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
			// try closing the currentTransactionChannel to ensure we never accidentally leak a channel before
			// register a new one
			safeCloseTransactionChannel(ch.currentTransactionChannel)
			ch.currentTransactionChannel = make(chan interface{})

			// create a new transaction in the transaction manager and wait for responses
			transactionmanager.GetTransactionManager().StartTransaction(ch.ID(), ch.GetCurrentTransaction(), ch.currentTransactionChannel)

			// create a new batch transaction in the transaction batcher
			transactionbatcher.GetTransactionalBatcher().StartTransaction(ch.ID(), ch.GetCurrentTransaction())

			// this is a blocking function. Will continue when a transaction succeeds, fails or times out making it
			// ready to handle the next transaction in the ch.transactionChannel.
			ch.handleCurrentTransaction(ch.currentTransactionChannel)

			// close the ch.currentTransactionChannel, making use ready to start a new transaction
			close(ch.currentTransactionChannel)
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
currentTxHandler:
	for {
		select {
		case tx := <-txChan:
			switch txMsg := tx.(type) {
			case transactionmanager.RollbackTransaction, transactionmanager.EvictedTransaction:
				_ = log.Warnf("Reloading check %s after failed transaction", ch.ID())
				if err := ch.ReloadCheck(ch.ID(), ch.config, ch.initConfig, ch.ConfigSource()); err != nil {
					_ = log.Errorf("failed to reload check %s: %s", ch.ID(), err)
				}
				break currentTxHandler
			case transactionmanager.CompleteTransaction:
				log.Debugf("Completing transaction: %s for check %s", txMsg.TransactionID, ch.ID())

				if txMsg.State != nil {
					log.Debugf("Committing state for transaction: %s for check %s: %s", txMsg.TransactionID, ch.ID(),
						txMsg.State)
					err := checkState.GetCheckStateManager().SetState(txMsg.State.Key, txMsg.State.State)
					if err != nil {
						errorReason := fmt.Sprintf("Error while updating state for transaction: %s for check %s: %s. %s",
							txMsg.TransactionID, ch.ID(), txMsg.State, err)
						_ = log.Error(errorReason)
						txChan <- transactionmanager.RollbackTransaction{TransactionID: txMsg.TransactionID, Reason: errorReason}
					}

					log.Debugf("Successfully committed state for transaction: %s for check %s: %s", txMsg.TransactionID, ch.ID(),
						txMsg.State)
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

// StartTransaction is used to start a transaction to the input channel
type StartTransaction struct {
	CheckID       check.ID
	TransactionID string
}

// StopTransaction is used to stop a transaction to the input channel
type StopTransaction struct{}

// safeCloseTransactionChannel closes the tx channel that can potentially already be closed. It handles the panic and does a no-op.
func safeCloseTransactionChannel(ch chan interface{}) {
	defer func() {
		if recover() != nil {
		}
	}()

	close(ch) // panic if ch is closed
}
