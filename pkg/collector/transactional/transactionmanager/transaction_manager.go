package transactionmanager

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"sync"
	"time"
)

var (
	tmInstance TransactionManager
	tmInit     sync.Once
)

// TransactionManager encapsulates all the functionality of the transaction manager to keep track of transactions
type TransactionManager interface {
	Start()
	StartTransaction(CheckID check.ID, TransactionID string, NotifyChannel chan interface{})
	CompleteTransaction(transactionID string)
	RollbackTransaction(transactionID, reason string)
	CommitAction(transactionID, actionID string)
	AcknowledgeAction(transactionID, actionID string)
	RejectAction(transactionID, actionID, reason string)
	Stop()
}

// InitTransactionManager ...
func InitTransactionManager(transactionChannelBufferSize int, tickerInterval, transactionTimeoutDuration,
	transactionEvictionDuration time.Duration) {
	tmInit.Do(func() {
		tmInstance = newTransactionManager(transactionChannelBufferSize, tickerInterval, transactionTimeoutDuration,
			transactionEvictionDuration)
	})
}

// GetTransactionManager returns a handle on the global transactionbatcher Instance
func GetTransactionManager() TransactionManager {
	return tmInstance
}

// NewMockTransactionManager returns a handle on the global transactionbatcher Instance
func NewMockTransactionManager() *MockTransactionManager {
	tm := newTestTransactionManager()
	tmInstance = tm
	return tm
}

// newTransactionManager returns an instance of a TransactionManager
func newTransactionManager(transactionChannelBufferSize int, tickerInterval, transactionTimeoutDuration,
	transactionEvictionDuration time.Duration) TransactionManager {
	return &transactionManager{
		transactionChannel:          make(chan interface{}, transactionChannelBufferSize),
		transactionTicker:           time.NewTicker(tickerInterval),
		transactions:                make(map[string]*IntakeTransaction),
		transactionTimeoutDuration:  transactionTimeoutDuration,
		transactionEvictionDuration: transactionEvictionDuration,
	}
}

// TransactionManager keeps track of all transactions for agent checks
type transactionManager struct {
	transactionChannel          chan interface{}
	transactionTicker           *time.Ticker
	transactions                map[string]*IntakeTransaction
	transactionTimeoutDuration  time.Duration
	transactionEvictionDuration time.Duration
	running                     bool
}

// CompleteTransaction completes a transaction for a given transactionID
func (txm *transactionManager) CompleteTransaction(transactionID string) {
	txm.transactionChannel <- CompleteTransaction{
		TransactionID: transactionID,
	}
}

// StartTransaction begins a transaction for a given check
func (txm *transactionManager) StartTransaction(checkID check.ID, transactionID string, notifyChannel chan interface{}) {
	txm.transactionChannel <- StartTransaction{
		CheckID:       checkID,
		TransactionID: transactionID,
		NotifyChannel: notifyChannel,
	}
}

// RollbackTransaction rolls back a transaction for a given transactionID and a reason for the rollback
func (txm *transactionManager) RollbackTransaction(transactionID, reason string) {
	txm.transactionChannel <- RollbackTransaction{
		TransactionID: transactionID,
		Reason:        reason,
	}
}

// CommitAction commits an action for a given transaction. All actions must be acknowledged for a given transaction
func (txm *transactionManager) CommitAction(transactionID, actionID string) {
	txm.transactionChannel <- CommitAction{
		TransactionID: transactionID,
		ActionID:      actionID,
	}
}

// AcknowledgeAction acknowledges an action for a given transaction
func (txm *transactionManager) AcknowledgeAction(transactionID, actionID string) {
	txm.transactionChannel <- AckAction{
		TransactionID: transactionID,
		ActionID:      actionID,
	}
}

// RejectAction rejects an action for a given transaction. This will result in a transaction failure
func (txm *transactionManager) RejectAction(transactionID, actionID, reason string) {
	txm.transactionChannel <- RejectAction{
		TransactionID: transactionID,
		ActionID:      actionID,
		Reason:        reason,
	}
}

// Start sets up the transaction checkmanager to consume messages on the txm.transactionChannel. It consumes one message at
// a time using the `select` statement and populates / evicts transactions in the transaction checkmanager.
func (txm *transactionManager) Start() {
	go func() {
	transactionHandler:
		for {
			select {
			case input := <-txm.transactionChannel:
				switch msg := input.(type) {
				// transaction operations
				case StartTransaction:
					log.Debugf("Creating new transaction %s for check %s", msg.TransactionID, msg.CheckID)
					if _, err := txm.startTransaction(msg.TransactionID, msg.NotifyChannel); err != nil {
						txm.transactionChannel <- err
					}
				case CommitAction:
					log.Debugf("Committing action %s for transaction %s", msg.ActionID, msg.TransactionID)
					if err := txm.commitAction(msg.TransactionID, msg.ActionID); err != nil {
						txm.transactionChannel <- err
					}
				case AckAction:
					log.Debugf("Acknowledging action %s for transaction %s", msg.ActionID, msg.TransactionID)
					if err := txm.ackAction(msg.TransactionID, msg.ActionID); err != nil {
						txm.transactionChannel <- err
					}
				case RejectAction:
					_ = log.Errorf("Rejecting action %s for transaction %s: %s", msg.ActionID, msg.TransactionID, msg.Reason)
					if err := txm.rejectAction(msg.TransactionID, msg.ActionID); err != nil {
						txm.transactionChannel <- err
					} else {
						// rollback the transaction
						reason := fmt.Sprintf("rejected action %s for transaction %s: %s", msg.ActionID, msg.TransactionID, msg.Reason)
						txm.transactionChannel <- RollbackTransaction{TransactionID: msg.TransactionID, Reason: reason}
					}
				case CompleteTransaction:
					log.Debugf("Completing transaction %s", msg.TransactionID)
					if err := txm.completeTransaction(msg.TransactionID); err != nil {
						txm.transactionChannel <- err
					}
				// error cases
				case RollbackTransaction:
					_ = log.Errorf(msg.Error())
					if err := txm.rollbackTransaction(msg.TransactionID); err != nil {
						txm.transactionChannel <- err
					}
				case TransactionNotFound:
					_ = log.Errorf(msg.Error())
				case ActionNotFound:
					_ = log.Errorf(msg.Error())
				// shutdown transaction checkmanager
				case StopTransactionManager:
					// clean the transaction map
					txm.transactions = make(map[string]*IntakeTransaction, 0)
					break transactionHandler
				default:
					_ = log.Errorf("Got unexpected msg %v", msg)
				}
			case <-txm.transactionTicker.C:
				// expire stale transactions, clean up expired transactions that exceed the eviction duration
				for _, transaction := range txm.transactions {
					if transaction.State == Failed || transaction.State == Succeeded {
						log.Debugf("Cleaning up %s transaction: %s", transaction.State.String(), transaction.TransactionID)
						// delete the transaction, already notified on success or failure status so no need to notify again
						delete(txm.transactions, transaction.TransactionID)
					} else if transaction.State != Stale && transaction.LastUpdatedTimestamp.Before(time.Now().Add(-txm.transactionTimeoutDuration)) {
						// last updated timestamp is before current time - checkmanager timeout duration => Tx is stale
						transaction.State = Stale
					} else if transaction.State == Stale && transaction.LastUpdatedTimestamp.Before(time.Now().Add(-txm.transactionEvictionDuration)) {
						// last updated timestamp is before current time - checkmanager eviction duration => Tx can be evicted
						txm.evictTransaction(transaction.TransactionID)
					}
				}

				// TODO: produce some transaction checkmanager metrics
			default:
			}
		}
	}()

	txm.running = true
}

// Stop shuts down the transaction checkmanager and stops the transactionHandler receiver loop
func (txm *transactionManager) Stop() {
	txm.running = false
	txm.transactionChannel <- StopTransactionManager{}
	txm.transactionTicker.Stop()
}

// startTransaction creates a transaction and puts it into the transactions map
func (txm *transactionManager) startTransaction(transactionID string, notify chan interface{}) (*IntakeTransaction, error) {
	if !txm.running {
		return nil, TransactionManagerNotRunning{}
	}

	transaction := &IntakeTransaction{
		TransactionID:        transactionID,
		State:                InProgress,
		Actions:              map[string]*Action{},
		NotifyChannel:        notify,
		LastUpdatedTimestamp: time.Now(),
	}

	txm.transactions[transaction.TransactionID] = transaction

	return transaction, nil
}

// commitAction commits / promises an action for a certain transaction. A commit is only a promise that something needs
// to be fulfilled. An unacknowledged action results in a transaction failure.
func (txm *transactionManager) commitAction(transactionID, actionID string) error {
	transaction, exists := txm.transactions[transactionID]
	if !exists {
		return TransactionNotFound{TransactionID: transactionID}
	}

	action := &Action{
		ActionID:           actionID,
		CommittedTimestamp: time.Now(),
	}
	txm.updateTransaction(transaction, action, InProgress)

	return nil
}

// updateTransaction is a helper function to set the state of a transaction as well as update it's LastUpdatedTimestamp.
func (txm *transactionManager) updateTransaction(transaction *IntakeTransaction, action *Action, state TransactionState) {
	transaction.Actions[action.ActionID] = action
	transaction.State = state
	transaction.LastUpdatedTimestamp = time.Now()
}

// ackAction acknowledges an action for a given transaction. This marks the action as acknowledged.
func (txm *transactionManager) ackAction(transactionID, actionID string) error {
	return txm.findAndUpdateAction(transactionID, actionID, true)
}

// rejectAction acknowledges an action for a given transaction. This marks the action as acknowledged and results in a
// failed transaction and rollback.
func (txm *transactionManager) rejectAction(transactionID, actionID string) error {
	return txm.findAndUpdateAction(transactionID, actionID, false)
}

// findAndUpdateAction is a helper function to find a transaction and action for the given ID's, marks the action as
// acknowledged and updates the transaction is updateTransaction is set to true.
func (txm *transactionManager) findAndUpdateAction(transactionID, actionID string, updateTransaction bool) error {
	transaction, exists := txm.transactions[transactionID]
	if !exists {
		return TransactionNotFound{TransactionID: transactionID}
	}

	action, exists := transaction.Actions[actionID]
	if !exists {
		return ActionNotFound{ActionID: actionID, TransactionID: transactionID}
	}
	action.Acknowledged = true
	action.AcknowledgedTimestamp = time.Now()

	if updateTransaction {
		txm.updateTransaction(transaction, action, InProgress)
	}

	return nil
}

// completeTransaction marks a transaction for a given transactionID as Succeeded, if all the committed actions
// of a transaction has been acknowledged
func (txm *transactionManager) completeTransaction(transactionID string) error {
	transaction, exists := txm.transactions[transactionID]
	if !exists {
		return TransactionNotFound{TransactionID: transactionID}
	}

	// ensure all actions have been acknowledged
	for _, action := range transaction.Actions {
		if !action.Acknowledged {
			reason := fmt.Sprintf("Not all actions have been acknowledged, rolling back checkmanager: %s", transaction.TransactionID)
			return RollbackTransaction{TransactionID: transactionID, Reason: reason}
		}
	}
	transaction.State = Succeeded
	transaction.LastUpdatedTimestamp = time.Now()

	transaction.NotifyChannel <- CompleteTransaction{}

	return nil
}

// evictTransaction delete a given transactionID from the transactions map
func (txm *transactionManager) evictTransaction(transactionID string) {
	transaction, exists := txm.transactions[transactionID]
	if !exists {
		return
	}

	delete(txm.transactions, transactionID)

	transaction.NotifyChannel <- EvictedTransaction{}
}

// rollbackTransaction rolls back the transaction in the event of a failure
func (txm *transactionManager) rollbackTransaction(transactionID string) error {
	transaction, exists := txm.transactions[transactionID]
	if !exists {
		return TransactionNotFound{TransactionID: transactionID}
	}

	transaction.State = Failed
	transaction.LastUpdatedTimestamp = time.Now()

	transaction.NotifyChannel <- RollbackTransaction{}

	return nil
}
