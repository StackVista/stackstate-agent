package handler

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/batcher"
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	checkState "github.com/DataDog/datadog-agent/pkg/collector/check/state"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionmanager"
)

// CheckManager acts as the grouping of check handlers and deals with the "lifecycle" of check handlers. This is a stackstate-specific addition to assist in producing stackstate-specific
// data from the agent.
type CheckManager interface {
	GetCheckHandler(checkID checkid.ID) CheckHandler
	MakeCheckHandlerTransactional(checkID checkid.ID) CheckHandler
	RegisterCheckHandler(check CheckIdentifier, config, initConfig integration.Data) CheckHandler
	UnsubscribeCheckHandler(checkID checkid.ID)
	Stop()
}

// CheckManager acts as the grouping of check handlers and deals with the "lifecycle" of check handlers.
type checkManagerImpl struct {
	checkHandlers        map[string]CheckHandler
	config               Config
	stateManager         checkState.CheckStateAPI
	transactionalBatcher transactionbatcher.TransactionalBatcher
	transactionalManager transactionmanager.TransactionManager
	batcher              batcher.Component
}

func SetupMockTransactionalComponents() (*batcher.MockBatcher, *transactionbatcher.MockTransactionalBatcher, *transactionmanager.MockTransactionManager, CheckManager) {
	// Set storage root for tests
	config.Datadog.SetWithoutSource("check_state_root_path", "/tmp/fake-datadog-run")

	batcher := batcher.NewMockBatcher()
	transactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	transactionalManager := transactionmanager.NewMockTransactionManager()

	return batcher, transactionalBatcher, transactionalManager, NewCheckManager(batcher, transactionalBatcher, transactionalManager)
}

func NewMockCheckManager() *checkManagerImpl {
	return NewCheckManager(batcher.NewMockBatcher(), transactionbatcher.NewMockTransactionalBatcher(), transactionmanager.NewMockTransactionManager())
}

// newCheckManager returns a instance of the Check Manager
func NewCheckManager(batcher batcher.Component, transactionalBatcher transactionbatcher.TransactionalBatcher, manager transactionmanager.TransactionManager) *checkManagerImpl {
	return &checkManagerImpl{
		checkHandlers:        make(map[string]CheckHandler),
		config:               GetCheckManagerConfig(),
		stateManager:         checkState.NewCheckStateManager(),
		transactionalBatcher: transactionalBatcher,
		transactionalManager: manager,
		batcher:              batcher,
	}
}

// GetCheckHandler returns a check handler (if found) for a given check ID
func (cm *checkManagerImpl) GetCheckHandler(checkID checkid.ID) CheckHandler {
	ch, found := cm.checkHandlers[string(checkID)]
	if !found {
		log.Debugf(fmt.Sprintf("No check handler found for %s. Registering a non-transactional check handler.", checkID))
		return cm.registerNonTransactionalCheckHandler(NewCheckIdentifier(checkID), nil, nil)
	}
	return ch
}

// registerNonTransactionalCheckHandler registers a non-transactional check handler for a given check
func (cm *checkManagerImpl) registerNonTransactionalCheckHandler(check CheckIdentifier, config, initConfig integration.Data) CheckHandler {
	ch := MakeNonTransactionalCheckHandler(cm, cm.batcher, cm.stateManager, check, config, initConfig)
	cm.checkHandlers[string(check.ID())] = ch
	return ch
}

// MakeCheckHandlerTransactional converts a non-transactional check handler to a transactional check handler
func (cm *checkManagerImpl) MakeCheckHandlerTransactional(checkID checkid.ID) CheckHandler {
	ch, found := cm.checkHandlers[string(checkID)]
	if !found {
		_ = log.Errorf("No check handler found for %s.", checkID)
		return nil
	}

	config, initConfig := ch.GetConfig()
	transactionalCheckHandler := NewTransactionalCheckHandler(cm.stateManager, cm.transactionalBatcher, cm.transactionalManager, NewCheckIdentifier(checkID), config, initConfig)
	cm.checkHandlers[string(checkID)] = transactionalCheckHandler

	return transactionalCheckHandler
}

// RegisterCheckHandler registers a check handler for the given check using a transactionbatcher for this instance
func (cm *checkManagerImpl) RegisterCheckHandler(check CheckIdentifier, config, initConfig integration.Data) CheckHandler {
	ch := cm.registerNonTransactionalCheckHandler(check, config, initConfig)
	log.Debugf("Registering Check Handler for: %s", ch.ID())
	cm.checkHandlers[string(check.ID())] = ch
	return ch
}

// UnsubscribeCheckHandler removes a check handler for the given check
func (cm *checkManagerImpl) UnsubscribeCheckHandler(checkID checkid.ID) {
	log.Debugf("Removing Check Handler for: %s", checkID)
	delete(cm.checkHandlers, string(checkID))
}

// Stop clears the check handlers and re-initializes the singleton init
func (cm *checkManagerImpl) Stop() {
	log.Debug("Removing all Check Handlers")
	// Question: should we stop the check handlers here?
	cm.checkHandlers = make(map[string]CheckHandler)
}
