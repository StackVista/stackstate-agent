package transactionalclientimpl

import (
	"context"
	"github.com/DataDog/datadog-agent/comp/core/hostname"
	"github.com/DataDog/datadog-agent/comp/core/log"
	"github.com/DataDog/datadog-agent/comp/stackstate/transactionalclient"
	"github.com/DataDog/datadog-agent/pkg/httpclient"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionforwarder"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionmanager"
	"go.uber.org/fx"
)

// Module defines the fx options for this component.
func Module() fxutil.Module {
	return fxutil.Component(
		fx.Provide(newTransactionalClient))
}

type dependencies struct {
	fx.In
	Log   log.Component
	HName hostname.Component

	Params Params
}

type provides struct {
	fx.Out
	TransactionalClient transactionalclient.Component
}

func newTransactionalClient(deps dependencies) (provides, error) {
	hname, err := deps.HName.Get(context.TODO())
	if err != nil {
		return provides{}, err
	}

	return provides{
		TransactionalClient: NewComponent(deps.Params, hname),
	}, nil
}

type transactionalClient struct {
	batcher   transactionbatcher.TransactionalBatcher
	manager   transactionmanager.TransactionManager
	forwarder transactionforwarder.TransactionalForwarder
}

func NewComponent(params Params, hostname string) transactionalclient.Component {
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
