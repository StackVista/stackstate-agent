package transactional

import (
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/google/uuid"
	"time"
)

type CommitAction struct {
	txID, actionId string
}

type AckAction struct {
	txID, actionId string
}

type RejectAction struct {
	txID, actionId string
	reason         string
}

type StopTxManager struct{}

// TransactionManager ...
type TransactionManager struct {
	TxChan             chan interface{}
	TxTicker           *time.Ticker
	Transactions       map[string]*IntakeTransaction
	txTimeoutDuration  time.Duration
	txEvictionDuration time.Duration
}

func MakeTransactionManager(txChannelBufferSize int, tickerInterval, txTimeoutDuration, txEvictionDuration time.Duration) *TransactionManager {
	return &TransactionManager{
		TxChan:             make(chan interface{}, txChannelBufferSize),
		TxTicker:           time.NewTicker(tickerInterval),
		Transactions:       make(map[string]*IntakeTransaction),
		txTimeoutDuration:  txTimeoutDuration,
		txEvictionDuration: txEvictionDuration,
	}
}

func (txm TransactionManager) Start() {
transactionHandler:
	for {
		select {
		case input := <-txm.TxChan:
			switch action := input.(type) {
			case CommitAction:
				txm.CommitAction(action.txID, action.actionId)
			case AckAction:
				txm.AckAction(action.txID, action.actionId)
			case RejectAction:
				txm.RollbackTransaction(action.txID)
			case StopTxManager:
				break transactionHandler
			}
		case <-txm.TxTicker.C:
			// expire stale transactions, clean up expired transactions that exceed the eviction duration
			for _, tx := range txm.Transactions {
				if tx.LastUpdatedTimestamp.After(time.Now().Add(-txm.txTimeoutDuration)) {
					// last updated timestamp is before current time - transaction timeout duration => Tx is stale
					tx.State = Stale
				} else if tx.State == Stale && tx.LastUpdatedTimestamp.After(time.Now().Add(-txm.txEvictionDuration)) {
					// last updated timestamp is before current time - transaction eviction duration => Tx can be evicted
					txm.EvictTransaction(tx.TransactionID)
				}
			}
		default:
		}
	}
}

func (txm TransactionManager) Stop() {
	txm.TxChan <- StopTxManager{}
	txm.TxTicker.Stop()
}

// StartTransaction ...
func (txm TransactionManager) StartTransaction() *IntakeTransaction {
	tx := &IntakeTransaction{
		TransactionID:        uuid.New().String(),
		State:                InProgress,
		Actions:              map[string]*Action{},
		LastUpdatedTimestamp: time.Now(),
	}
	txm.Transactions[tx.TransactionID] = tx

	return tx
}

// CommitAction ...
func (txm TransactionManager) CommitAction(txID, actionId string) {
	tx, exists := txm.Transactions[txID]
	if exists {
		action := &Action{
			ActionID:           actionId,
			CommittedTimestamp: time.Now(),
		}
		tx.Actions[actionId] = action
		tx.State = InProgress
		tx.LastUpdatedTimestamp = time.Now()
		log.Debugf("Transaction %s, committing action %s", txID, actionId)
	}
}

// AckAction ...
func (txm TransactionManager) AckAction(txID, actionId string) {
	tx, exists := txm.Transactions[txID]
	if exists {
		act, exists := tx.Actions[actionId]
		if exists {
			act.Acknowledged = true
			act.AcknowledgedTimestamp = time.Now()
			tx.Actions[actionId] = act
			tx.State = InProgress
			tx.LastUpdatedTimestamp = time.Now()
			log.Debugf("Transaction %s, acknowledged action %s", txID, actionId)
		}
	}
}

// SucceedTransaction marks the transaction successful
func (txm TransactionManager) SucceedTransaction(txID string) {
	tx, exists := txm.Transactions[txID]
	if exists {
		// ensure all actions have been acknowledged
		for _, a := range tx.Actions {
			if !a.Acknowledged {
				_ = log.Errorf("Not all actions have been acknowledged, rolling back transaction: %s", tx.TransactionID)
				txm.RollbackTransaction(txID)
			}
		}
		tx.State = Succeeded
		tx.LastUpdatedTimestamp = time.Now()
		log.Debugf("Transaction succeeded %s", tx.TransactionID)
	} else {
		_ = log.Warnf("Transaction not found %s, no operation", tx.TransactionID)
	}

	tx.OnComplete(tx)

	txm.EvictTransaction(txID)
}

// EvictTransaction evicts the intake transaction due to failure or timeout
func (txm TransactionManager) EvictTransaction(txID string) {
	delete(txm.Transactions, txID)
}

// RollbackTransaction rolls back the intake transaction
func (txm TransactionManager) RollbackTransaction(txID string) {
	tx, exists := txm.Transactions[txID]
	if exists {
		// transaction failed, rollback
		tx.State = Failed
		tx.LastUpdatedTimestamp = time.Now()
		log.Debugf("Transaction failed %s", tx.TransactionID)
	}

	tx.OnComplete(tx)

	txm.EvictTransaction(txID)
}
