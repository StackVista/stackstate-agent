package transactional

import (
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/google/uuid"
	"time"
)

// TransactionManager ...
type TransactionManager struct {
	Transactions map[string]*IntakeTransaction
}

// StartTransaction ...
func (txm TransactionManager) StartTransaction() {
	tx := &IntakeTransaction{
		TransactionID: uuid.New().String(),
		State:         InProgress,
		Actions:       map[string]*Action{},
	}
	txm.Transactions[tx.TransactionID] = tx
}

// CommitTransaction ...
func (txm TransactionManager) CommitTransaction(txID, actionId string) {
	tx, exists := txm.Transactions[txID]
	if exists {
		action := &Action{
			ActionID:  actionId,
			Timestamp: time.Now(),
		}
		tx.Actions[actionId] = action
		log.Debugf("Transaction %s, committing action %s", txID, actionId)
	}
}

// AckTransaction ...
func (txm TransactionManager) AckTransaction(txID, actionId string) {
	tx, exists := txm.Transactions[txID]
	if exists {
		act, exists := tx.Actions[actionId]
		if exists {
			act.Acknowledged = true
			act.AcknowledgedTimestamp = time.Now()
			tx.Actions[actionId] = act
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
		log.Debugf("Transaction succeeded %s", tx.TransactionID)
	} else {
		_ = log.Warnf("Transaction not found %s, no operation", tx.TransactionID)
	}

	tx.OnComplete(tx)
}

// RollbackTransaction rolls back the intake transaction
func (txm TransactionManager) RollbackTransaction(txID string) {
	tx, exists := txm.Transactions[txID]
	if exists {
		// transaction failed, rollback
		tx.State = Failed
		log.Debugf("Transaction failed %s", tx.TransactionID)
	}

	tx.OnComplete(tx)
}
