package transactionmanager

import (
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
)

// TransactionAPI contains the functions required for transactional behaviour
type TransactionAPI interface {
	GetActiveTransaction(transactionID string) (*IntakeTransaction, error)
	TransactionCount() int
	StartTransaction(CheckID checkid.ID, TransactionID string, NotifyChannel chan interface{})
	CompleteTransaction(transactionID string)
	DiscardTransaction(transactionID, reason string)
	CommitAction(transactionID, actionID string)
	AcknowledgeAction(transactionID, actionID string)
	SetState(transactionID, key string, state string)
	RejectAction(transactionID, actionID, reason string)
}
