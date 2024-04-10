package transactionmanager

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"sync"
	"time"
)

var (
	tmInstance TransactionManager
	tmInit     *sync.Once
)

func init() {
	tmInit = new(sync.Once)
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
transactionHandler:
	for {
		select {
		case input := <-txm.transactionChannel:
			switch msg := input.(type) {
			// transaction operations
			case StartTransaction:
				log.Debugf("Creating new transaction %s for check %s", msg.TransactionID, msg.CheckID)
				if _, err := txm.startTransaction(msg.TransactionID, msg.CheckID, msg.NotifyChannel); err != nil {
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
			case SetTransactionState:
				log.Debugf("Setting state %s for transaction %s: %s", msg.Key, msg.TransactionID, msg.State)
				if err := txm.setTransactionState(msg.TransactionID, msg.Key, msg.State); err != nil {
					txm.transactionChannel <- err
				}
			case RejectAction:
				_ = log.Errorf("Rejecting action %s for transaction %s: %s", msg.ActionID, msg.TransactionID, msg.Reason)
				if err := txm.rejectAction(msg.TransactionID, msg.ActionID); err != nil {
					txm.transactionChannel <- err
				} else {
					// discard the transaction
					reason := fmt.Sprintf("rejected action %s for transaction %s: %s", msg.ActionID, msg.TransactionID, msg.Reason)
					txm.transactionChannel <- DiscardTransaction{TransactionID: msg.TransactionID, Reason: reason}
				}
			case CompleteTransaction:
				log.Debugf("Completing transaction %s", msg.TransactionID)
				if err := txm.completeTransaction(msg.TransactionID); err != nil {
					txm.transactionChannel <- err
				}
			// error cases
			case DiscardTransaction:
				_ = log.Errorf(msg.Error())
				if err := txm.discardTransaction(msg.TransactionID, msg.Reason); err != nil {
					txm.transactionChannel <- err
				}
			case TransactionNotFound:
				_ = log.Errorf(msg.Error())
			case ActionNotFound:
				_ = log.Errorf(msg.Error())
			case TransactionCompleted:
				log.Debugf(msg.Error())
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
				if transaction.Status == Failed || transaction.Status == Succeeded {
					log.Debugf("Cleaning up %s transaction: %s for check: %s", transaction.Status.String(),
						transaction.TransactionID, transaction.CheckID)
					// delete the transaction, already notified on success or failure status so no need to notify again
					delete(txm.transactions, transaction.TransactionID)
				} else if transaction.Status != Stale && transaction.LastUpdatedTimestamp.Before(time.Now().Add(-txm.transactionTimeoutDuration)) {
					// last updated timestamp is before current time - checkmanager timeout duration => Tx is stale
					_ = log.Warnf("Transaction: %s for check %s has become stale, last updated %s",
						transaction.TransactionID, transaction.CheckID, transaction.LastUpdatedTimestamp.String())
					transaction.Status = Stale
				} else if transaction.Status == Stale && transaction.LastUpdatedTimestamp.Before(time.Now().Add(-txm.transactionEvictionDuration)) {
					// last updated timestamp is before current time - checkmanager eviction duration => Tx can be evicted
					_ = log.Warnf("Transaction: %s for check %s is stale and will be evicted, last updated %s",
						transaction.TransactionID, transaction.CheckID, transaction.LastUpdatedTimestamp.String())
					delete(txm.transactions, transaction.TransactionID)
					transaction.NotifyChannel <- EvictedTransaction{TransactionID: transaction.TransactionID}
				}
			}
			txm.mux.Unlock()

			// TODO: produce some transaction checkmanager metrics
		}
	}
}

// startTransaction creates a transaction and puts it into the transactions map
func (txm *transactionManager) startTransaction(transactionID string, checkID check.ID, notify chan interface{}) (*IntakeTransaction, error) {
	transaction := &IntakeTransaction{
		TransactionID:        transactionID,
		CheckID:              checkID,
		Status:               InProgress,
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
	transaction, err := txm.GetActiveTransaction(transactionID)
	if err != nil {
		return err
	}
	txm.mux.Lock()
	action := &Action{
		ActionID:               actionID,
		CommittedTimestamp:     time.Now(),
		Status:                 Committed,
		StatusUpdatedTimestamp: time.Now(),
	}
	txm.updateTransaction(transaction, action, InProgress)
	txm.mux.Unlock()

	return nil
}

// updateTransaction is a helper function to set the state of a transaction as well as update it's LastUpdatedTimestamp.
func (txm *transactionManager) updateTransaction(transaction *IntakeTransaction, action *Action, status TransactionStatus) {
	transaction.Actions[action.ActionID] = action
	transaction.Status = status
	transaction.LastUpdatedTimestamp = time.Now()
}

// ackAction acknowledges an action for a given transaction. This marks the action as acknowledged.
func (txm *transactionManager) ackAction(transactionID, actionID string) error {
	transaction, err := txm.GetActiveTransaction(transactionID)
	if err != nil {
		return err
	}

	txm.mux.Lock()

	action, exists := transaction.Actions[actionID]
	if !exists {
		txm.mux.Unlock()
		return ActionNotFound{ActionID: actionID, TransactionID: transactionID}
	}
	action.Status = Acknowledged
	action.StatusUpdatedTimestamp = time.Now()

	txm.updateTransaction(transaction, action, InProgress)

	txm.mux.Unlock()

	return nil
}

// setTransactionState sets the state for a given key and CheckState. The state for a given transaction will be
// committed on a successful completion of the transaction
func (txm *transactionManager) setTransactionState(transactionID, key string, state string) error {
	transaction, err := txm.GetActiveTransaction(transactionID)
	if err != nil {
		return err
	}
	txm.mux.Lock()
	transaction.State = &TransactionState{
		Key:   key,
		State: state,
	}
	transaction.Status = InProgress
	transaction.LastUpdatedTimestamp = time.Now()
	txm.mux.Unlock()

	return nil
}

// rejectAction rejects an action for a given transaction. This marks the action as rejected and results in a
// failed transaction and discarding.
func (txm *transactionManager) rejectAction(transactionID, actionID string) error {
	transaction, err := txm.GetActiveTransaction(transactionID)
	if err != nil {
		return err
	}

	txm.mux.Lock()

	action, exists := transaction.Actions[actionID]
	if !exists {
		txm.mux.Unlock()
		return ActionNotFound{ActionID: actionID, TransactionID: transactionID}
	}
	action.Status = Rejected
	action.StatusUpdatedTimestamp = time.Now()

	txm.mux.Unlock()

	return nil
}

// completeTransaction marks a transaction for a given transactionID as Succeeded, if all the committed actions
// of a transaction has been acknowledged
func (txm *transactionManager) completeTransaction(transactionID string) error {
	transaction, err := txm.GetActiveTransaction(transactionID)
	if err != nil {
		return err
	}
	txm.mux.Lock()
	// ensure all actions have been acknowledged
	for _, action := range transaction.Actions {
		if action.Status != Acknowledged {
			_ = log.Errorf("Action %s for transaction %s has not been acknowledged. Discarding transaction",
				action.ActionID, transaction.TransactionID)
			reason := fmt.Sprintf("Not all actions have been acknowledged, discarding transaction: %s", transaction.TransactionID)
			txm.mux.Unlock()
			return DiscardTransaction{TransactionID: transactionID, Reason: reason}
		}
	}
	transaction.Status = Succeeded
	transaction.LastUpdatedTimestamp = time.Now()
	state := transaction.State
	txm.mux.Unlock()
	transaction.NotifyChannel <- CompleteTransaction{TransactionID: transactionID, State: state}

	return nil
}

// discardTransaction rolls back the transaction in the event of a failure
func (txm *transactionManager) discardTransaction(transactionID, reason string) error {
	transaction, err := txm.GetActiveTransaction(transactionID)
	if err != nil {
		return err
	}

	txm.mux.Lock()
	transaction.Status = Failed
	transaction.LastUpdatedTimestamp = time.Now()
	txm.mux.Unlock()
	transaction.NotifyChannel <- DiscardTransaction{TransactionID: transactionID, Reason: reason}

	return nil
}
