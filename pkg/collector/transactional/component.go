package transactional

import (
	"github.com/DataDog/datadog-agent/pkg/httpclient"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionforwarder"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionmanager"
)

type Component interface {
	GetBatcher() transactionbatcher.TransactionalBatcher
	GetManager() transactionmanager.TransactionManager
	Stop()
}

type transactionalClient struct {
	batcher   transactionbatcher.TransactionalBatcher
	manager   transactionmanager.TransactionManager
	forwarder transactionforwarder.TransactionalForwarder
}

func NewComponent(params Params, hostname string) Component {
	manager := transactionmanager.NewTransactionManager(params.transactionChannelBufferSize, params.tickerInterval, params.transactionTimeoutDuration, params.transactionTimeoutDuration)
	client := httpclient.NewStackStateClient()
	forwarder := transactionforwarder.NewTransactionalForwarder(client, manager)
	batcher := transactionbatcher.NewTransactionalBatcher(hostname, params.maxCapacity, forwarder, manager, false)

	return transactionalClient{
		batcher:   batcher,
		manager:   manager,
		forwarder: forwarder,
	}
}

func (c transactionalClient) GetBatcher() transactionbatcher.TransactionalBatcher {
	return c.batcher
}

func (c transactionalClient) GetManager() transactionmanager.TransactionManager {
	return c.manager
}

func (c transactionalClient) Stop() {
	c.batcher.Stop()
	c.forwarder.Stop()
	c.manager.Stop()
}
