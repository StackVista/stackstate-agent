package transactionmanager

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
)

type TransactionAPI interface {
	GetTransaction(transactionID string) (*IntakeTransaction, error)
	TransactionCount() int
	StartTransaction(CheckID check.ID, TransactionID string, NotifyChannel chan interface{})
	CompleteTransaction(transactionID string)
	RollbackTransaction(transactionID, reason string) // TODO: rename to DiscardTransaction
	CommitAction(transactionID, actionID string)
	AcknowledgeAction(transactionID, actionID string)
	SetState(transactionID, key string, state string)
	RejectAction(transactionID, actionID, reason string)
}
