package transactional

import "github.com/StackVista/stackstate-agent/pkg/util/log"

type TransactionManager struct {
	Transactions map[string]IntakeTransaction
}

func (txm TransactionManager) StartTransaction(tx IntakeTransaction) {
	tx, exists := txm.Transactions[tx.TransactionID]
	if exists {
		_ = log.Warnf("Transaction %s already exists, updating existing transaction", tx.TransactionID)
	}

	txm.Transactions[tx.TransactionID] = tx
}
