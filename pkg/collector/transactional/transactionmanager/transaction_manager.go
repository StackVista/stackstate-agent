package transactionmanager

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"sync"
	"time"
)

var (
	tmInstance TransactionManager
	tmInit     sync.Once
)

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
	tmInit.Do(func() {
		tmInstance = newTestTransactionManager()
	})
	return tmInstance.(*MockTransactionManager)
}

// newTransactionManager returns an instance of a TransactionManager
func newTransactionManager(transactionChannelBufferSize int, tickerInterval, transactionTimeoutDuration,
	transactionEvictionDuration time.Duration) TransactionManager {
	tm := &transactionManager{
		transactionChannel:          make(chan interface{}, transactionChannelBufferSize),
		transactionTicker:           time.NewTicker(tickerInterval),
		transactions:                make(map[string]*IntakeTransaction),
		transactionTimeoutDuration:  transactionTimeoutDuration,
		transactionEvictionDuration: transactionEvictionDuration,
	}

	go tm.Start()

	return tm
}

// TransactionManager keeps track of all transactions for agent checks
type transactionManager struct {
	transactionChannel          chan interface{}
	transactionTicker           *time.Ticker
	transactions                map[string]*IntakeTransaction // pointer for in-place mutation
	transactionTimeoutDuration  time.Duration
	transactionEvictionDuration time.Duration
	mux                         sync.RWMutex
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
					if err := txm.rollbackTransaction(msg.TransactionID, msg.Reason); err != nil {
						txm.transactionChannel <- err
					}
				case TransactionNotFound:
					_ = log.Errorf(msg.Error())
				case ActionNotFound:
					_ = log.Errorf(msg.Error())
				// shutdown transaction checkmanager
				case StopTransactionManager:
					// clean the transaction map
					txm.mux.Lock()
					txm.transactions = make(map[string]*IntakeTransaction, 0)
					txm.mux.Unlock()
					break transactionHandler
				default:
					_ = log.Errorf("Got unexpected msg %v", msg)
				}
			case <-txm.transactionTicker.C:
				// expire stale transactions, clean up expired transactions that exceed the eviction duration
				txm.mux.Lock()
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
						delete(txm.transactions, transaction.TransactionID)
						transaction.NotifyChannel <- EvictedTransaction{TransactionID: transaction.TransactionID}
					}
				}
				txm.mux.Unlock()

				// TODO: produce some transaction checkmanager metrics
			default:
			}
		}
	}()
}

// startTransaction creates a transaction and puts it into the transactions map
func (txm *transactionManager) startTransaction(transactionID string, notify chan interface{}) (*IntakeTransaction, error) {
	transaction := &IntakeTransaction{
		TransactionID:        transactionID,
		State:                InProgress,
		Actions:              map[string]*Action{},
		NotifyChannel:        notify,
		LastUpdatedTimestamp: time.Now(),
	}
	txm.mux.Lock()
	txm.transactions[transaction.TransactionID] = transaction
	txm.mux.Unlock()

	return transaction, nil
}

// commitAction commits / promises an action for a certain transaction. A commit is only a promise that something needs
// to be fulfilled. An unacknowledged action results in a transaction failure.
func (txm *transactionManager) commitAction(transactionID, actionID string) error {
	transaction, err := txm.GetTransaction(transactionID)
	if err != nil {
		return err
	}
	txm.mux.Lock()
	action := &Action{
		ActionID:           actionID,
		CommittedTimestamp: time.Now(),
	}
	txm.updateTransaction(transaction, action, InProgress)
	txm.mux.Unlock()

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
	transaction, err := txm.GetTransaction(transactionID)
	if err != nil {
		return err
	}
	txm.mux.Lock()
	action, exists := transaction.Actions[actionID]
	if !exists {
		txm.mux.Unlock()
		return ActionNotFound{ActionID: actionID, TransactionID: transactionID}
	}
	action.Acknowledged = true
	action.AcknowledgedTimestamp = time.Now()

	if updateTransaction {
		txm.updateTransaction(transaction, action, InProgress)
	}
	txm.mux.Unlock()

	return nil
}

// completeTransaction marks a transaction for a given transactionID as Succeeded, if all the committed actions
// of a transaction has been acknowledged
func (txm *transactionManager) completeTransaction(transactionID string) error {
	transaction, err := txm.GetTransaction(transactionID)
	if err != nil {
		return err
	}
	txm.mux.Lock()
	// ensure all actions have been acknowledged
	for _, action := range transaction.Actions {
		if !action.Acknowledged {
			reason := fmt.Sprintf("Not all actions have been acknowledged, rolling back transaction: %s", transaction.TransactionID)
			txm.mux.Unlock()
			return RollbackTransaction{TransactionID: transactionID, Reason: reason}
		}
	}
	transaction.State = Succeeded
	transaction.LastUpdatedTimestamp = time.Now()
	txm.mux.Unlock()
	transaction.NotifyChannel <- CompleteTransaction{TransactionID: transactionID}

	return nil
}

// rollbackTransaction rolls back the transaction in the event of a failure
func (txm *transactionManager) rollbackTransaction(transactionID, reason string) error {
	transaction, err := txm.GetTransaction(transactionID)
	if err != nil {
		return err
	}

	txm.mux.Lock()
	transaction.State = Failed
	transaction.LastUpdatedTimestamp = time.Now()
	txm.mux.Unlock()
	transaction.NotifyChannel <- RollbackTransaction{TransactionID: transactionID, Reason: reason}

	return nil
}
