package transactionalclient

import (
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionmanager"
)

type Component interface {
	GetBatcher() transactionbatcher.TransactionalBatcher
	GetManager() transactionmanager.TransactionManager
	Stop()
}
