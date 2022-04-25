package handler

import (
	"github.com/StackVista/stackstate-agent/cmd/agent/common"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/manager"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
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

type CheckHandler interface {
	CheckIdentifier
	CheckReloader
	GetConfig() (config, initConfig integration.Data)
	GetBatcher() batcher.Batcher
	GetCheckReloader() CheckReloader
}

type CheckHandlerBase struct {
	CheckIdentifier
	CheckReloader
	batcher.Batcher
	config, initConfig integration.Data
}

// checkHandler provides an interface between the Agent Check and the transactional components
type checkHandler struct {
	CheckHandlerBase
	CheckReloader
	batcher.Batcher
	shutdownChannel    chan interface{}
	transactionChannel chan SubmitStartTransaction
}

func MakeCheckHandler(check check.Check, checkReloader CheckReloader, checkBatcher batcher.Batcher, config, initConfig integration.Data) CheckHandler {
	return &checkHandler{
		CheckHandlerBase: CheckHandlerBase{
			CheckIdentifier: check,
			CheckReloader:   checkReloader,
			config:          config,
			initConfig:      initConfig,
		},
		Batcher: checkBatcher,
	}
}

type TxChan struct {
	TransactionID string
	TxChan        chan interface{}
}

// Start ...
func (ch *checkHandler) Start() {
txReceiverHandler:
	select {
	case startTx := <-ch.transactionChannel:
		thisTransactionChannel := make(chan interface{}, 20)
		common.TxManager.StartTransaction(startTx.CheckID, startTx.TransactionID, thisTransactionChannel)
		defer close(thisTransactionChannel)
		ch.handleCurrentTransaction(thisTransactionChannel)
	case _ = <-ch.shutdownChannel:
		break txReceiverHandler
	}
}

func (ch *checkHandler) handleCurrentTransaction(txChan chan interface{}) {
currentTxHandler:
	select {
	case tx := <-txChan:
		switch _ := tx.(type) {
		case manager.RollbackTransaction:
			if err := ch.ReloadCheck(ch.ID(), ch.config, ch.initConfig, ch.ConfigSource()); err != nil {
				_ = log.Errorf("failed to reload check %s: %s", ch.ID(), err)
			}
			break currentTxHandler
		case manager.CompleteTransaction:
			break currentTxHandler
		}
	default:
		// received nothing
	}
}

// StartTransaction ...
func (ch *checkHandler) StartTransaction(CheckID check.ID, TransactionID string) {
	ch.transactionChannel <- SubmitStartTransaction{
		CheckID:       CheckID,
		TransactionID: TransactionID,
	}
}

// GetCheckIdentifier ...
func (ch *checkHandler) GetCheckIdentifier() CheckIdentifier {
	return ch.CheckIdentifier
}

// GetConfig ...
func (ch *checkHandler) GetConfig() (integration.Data, integration.Data) {
	return ch.config, ch.initConfig
}

// GetBatcher ...
func (ch *checkHandler) GetBatcher() batcher.Batcher {
	return ch.Batcher
}

// GetCheckReloader returns the configured CheckReloader
func (ch *checkHandler) GetCheckReloader() CheckReloader {
	return ch.CheckReloader
}

// SubmitStartTransaction is used to start a transaction to the input channel
type SubmitStartTransaction struct {
	CheckID       check.ID
	TransactionID string
}
