package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/google/uuid"
)

type CheckReloader interface {
	ReloadCheck(id check.ID, config, initConfig integration.Data, newSource string) error
}

// CheckIdentifier encapsulates all the functionality needed to describe and configure an agent check
type CheckIdentifier interface {
	String() string       // provide a printable version of the check name
	ID() check.ID         // provide a unique identifier for every check instance
	ConfigSource() string // return the configuration source of the check
}

// CheckAPI ...
type CheckAPI interface {
	// Transactionality
	SubmitStartTransaction()
	SubmitStopTransaction()

	// Topology
	SubmitComponent(instance topology.Instance, component topology.Component)
	SubmitRelation(instance topology.Instance, relation topology.Relation)
	SubmitStartSnapshot(instance topology.Instance)
	SubmitStopSnapshot(instance topology.Instance)
	SubmitDelete(instance topology.Instance, topologyElementID string)

	// Health
	SubmitHealthCheckData(stream health.Stream, data health.CheckData)
	SubmitHealthStartSnapshot(stream health.Stream, intervalSeconds int, expirySeconds int)
	SubmitHealthStopSnapshot(stream health.Stream)

	// Raw Metrics
	SubmitRawMetricsData(data telemetry.RawMetrics)

	// lifecycle
	SubmitComplete()
	Shutdown()
}

type CheckHandler interface {
	CheckIdentifier
	CheckAPI
	GetConfig() (config, initConfig integration.Data)
	GetCheckReloader() CheckReloader
	Reload()
}

// checkHandler provides an interface between the Agent Check and the transactional components
type checkHandler struct {
	CheckIdentifier
	CheckReloader
	config, initConfig        integration.Data
	shutdownChannel           chan bool
	transactionChannel        chan SubmitStartTransaction
	currentTransactionChannel chan interface{}
	currentTransaction        string
}

func (ch *checkHandler) GetCheckReloader() CheckReloader {
	return ch.CheckReloader
}

// SubmitComponent ...
func (ch *checkHandler) SubmitComponent(instance topology.Instance, component topology.Component) {
	transactionbatcher.GetTransactionalBatcher().SubmitComponent(ch.ID(), ch.currentTransaction, instance, component)
}

// SubmitRelation ...
func (ch *checkHandler) SubmitRelation(instance topology.Instance, relation topology.Relation) {
	transactionbatcher.GetTransactionalBatcher().SubmitRelation(ch.ID(), ch.currentTransaction, instance, relation)
}

// SubmitStartSnapshot ...
func (ch *checkHandler) SubmitStartSnapshot(instance topology.Instance) {
	transactionbatcher.GetTransactionalBatcher().SubmitStartSnapshot(ch.ID(), ch.currentTransaction, instance)
}

// SubmitStopSnapshot ...
func (ch *checkHandler) SubmitStopSnapshot(instance topology.Instance) {
	transactionbatcher.GetTransactionalBatcher().SubmitStopSnapshot(ch.ID(), ch.currentTransaction, instance)
}

// SubmitDelete ...
func (ch *checkHandler) SubmitDelete(instance topology.Instance, topologyElementID string) {
	transactionbatcher.GetTransactionalBatcher().SubmitDelete(ch.ID(), ch.currentTransaction, instance, topologyElementID)
}

// SubmitHealthCheckData ...
func (ch *checkHandler) SubmitHealthCheckData(stream health.Stream, data health.CheckData) {
	transactionbatcher.GetTransactionalBatcher().SubmitHealthCheckData(ch.ID(), ch.currentTransaction, stream, data)
}

// SubmitHealthStartSnapshot ...
func (ch *checkHandler) SubmitHealthStartSnapshot(stream health.Stream, intervalSeconds int, expirySeconds int) {
	transactionbatcher.GetTransactionalBatcher().SubmitHealthStartSnapshot(ch.ID(), ch.currentTransaction, stream, intervalSeconds, expirySeconds)
}

// SubmitHealthStopSnapshot ...
func (ch *checkHandler) SubmitHealthStopSnapshot(stream health.Stream) {
	transactionbatcher.GetTransactionalBatcher().SubmitHealthStopSnapshot(ch.ID(), ch.currentTransaction, stream)
}

// SubmitRawMetricsData ...
func (ch *checkHandler) SubmitRawMetricsData(data telemetry.RawMetrics) {
	transactionbatcher.GetTransactionalBatcher().SubmitRawMetricsData(ch.ID(), ch.currentTransaction, data)
}

// SubmitStartTransaction ...
func (ch *checkHandler) SubmitStartTransaction() {
	ch.transactionChannel <- SubmitStartTransaction{}
}

// SubmitStopTransaction ...
func (ch *checkHandler) SubmitStopTransaction() {
	transactionbatcher.GetTransactionalBatcher().SubmitComplete(ch.ID())
}

// SubmitComplete ...
func (ch *checkHandler) SubmitComplete() {
	transactionbatcher.GetTransactionalBatcher().SubmitComplete(ch.ID())
}

// Shutdown ...
func (ch *checkHandler) Shutdown() {
	transactionbatcher.GetTransactionalBatcher().Shutdown()
}

// Reload ...
func (ch *checkHandler) Reload() {
	config, initConfig := ch.GetConfig()
	if err := ch.ReloadCheck(ch.ID(), config, initConfig, ch.ConfigSource()); err != nil {
		_ = log.Errorf("could not reload check %s", string(ch.ID()))
	}
}

// NewCheckHandler ...
func NewCheckHandler(check check.Check, checkReloader CheckReloader, config, initConfig integration.Data) CheckHandler {
	return &checkHandler{
		CheckIdentifier:    check,
		CheckReloader:      checkReloader,
		config:             config,
		initConfig:         initConfig,
		shutdownChannel:    make(chan bool, 1),
		transactionChannel: make(chan SubmitStartTransaction, 1),
	}
}

// Start ...
func (ch *checkHandler) Start() {
	go func() {
	txReceiverHandler:
		for {
			select {
			case <-ch.transactionChannel:
				// set the current transaction
				ch.currentTransaction = uuid.New().String()
				log.Debugf("Starting transaction: %s", ch.currentTransaction)
				ch.currentTransactionChannel = make(chan interface{})
				transactionmanager.GetTransactionManager().StartTransaction(ch.ID(), ch.currentTransaction, ch.currentTransactionChannel)
				ch.handleCurrentTransaction(ch.currentTransactionChannel)
				close(ch.currentTransactionChannel)
			case <-ch.shutdownChannel:
				log.Debug("Shutting down check handler. Closing transaction channels.")
				// try closing currentTransactionChannel if the transaction is still in progress
				safeCloseTransactionChannel(ch.currentTransactionChannel)
				close(ch.transactionChannel)
				break txReceiverHandler
			default:
				log.Debug("got here")
				// nothing
			}
		}
	}()
}

// Stop ...
func (ch *checkHandler) Stop() {
	ch.shutdownChannel <- true
}

func (ch *checkHandler) handleCurrentTransaction(txChan chan interface{}) {
currentTxHandler:
	for {
		select {
		case tx := <-txChan:
			switch txMsg := tx.(type) {
			case transactionmanager.RollbackTransaction, transactionmanager.EvictedTransaction:
				if err := ch.ReloadCheck(ch.ID(), ch.config, ch.initConfig, ch.ConfigSource()); err != nil {
					_ = log.Errorf("failed to reload check %s: %s", ch.ID(), err)
				}
				break currentTxHandler
			case transactionmanager.CompleteTransaction:
				log.Debugf("Received: %s", txMsg.TransactionID)
				break currentTxHandler
			}
		}
	}
}

// GetConfig returns the config and the init config of the check
func (ch *checkHandler) GetConfig() (integration.Data, integration.Data) {
	return ch.config, ch.initConfig
}

// SubmitStartTransaction is used to start a transaction to the input channel
type SubmitStartTransaction struct{}

// SubmitStopTransaction is used to stop a transaction to the input channel
type SubmitStopTransaction struct{}

// safeCloseTransactionChannel closes the tx channel that can potentially already be closed. It handles the panic and does a no-op.
func safeCloseTransactionChannel(ch chan interface{}) {
	defer func() {
		if recover() != nil {
		}
	}()

	close(ch) // panic if ch is closed
}
