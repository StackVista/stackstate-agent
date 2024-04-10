package handler

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"sync"
)

var (
	cmInstance    *CheckManager
	cmInit        *sync.Once
	cmInitialized bool
)

func init() {
	cmInit = new(sync.Once)
}

// InitCheckManager ...
func InitCheckManager() {
	cmInit.Do(func() {
		cmInstance = newCheckManager()
	})
}

// GetCheckManager returns a handle on the global check checkmanager Instance
func GetCheckManager() *CheckManager {
	return cmInstance
}

// CheckManager acts as the grouping of check handlers and deals with the "lifecycle" of check handlers.
type CheckManager struct {
	checkHandlers map[string]CheckHandler
	config        Config
}

// newCheckManager returns a instance of the Check Manager
func newCheckManager() *CheckManager {
	cmInitialized = true

	return &CheckManager{
		checkHandlers: make(map[string]CheckHandler),
		config:        GetCheckManagerConfig(),
	}
}

// GetCheckHandler returns a check handler (if found) for a given check ID
func (cm *CheckManager) GetCheckHandler(checkID check.ID) CheckHandler {
	if !cmInitialized {
		_ = log.Errorf("CheckManager not initialized, initialize it using handler.InitCheckManager()")
		return nil
	}

	ch, found := cm.checkHandlers[string(checkID)]
	if !found {
		log.Debugf(fmt.Sprintf("No check handler found for %s. Registering a non-transactional check handler.", checkID))
		return cm.registerNonTransactionalCheckHandler(NewCheckIdentifier(checkID), nil, nil)
	}
	return ch
}

// registerNonTransactionalCheckHandler registers a non-transactional check handler for a given check
func (cm *CheckManager) registerNonTransactionalCheckHandler(check CheckIdentifier, config, initConfig integration.Data) CheckHandler {
	ch := MakeNonTransactionalCheckHandler(check, config, initConfig)
	cm.checkHandlers[string(check.ID())] = ch
	return ch
}

// MakeCheckHandlerTransactional converts a non-transactional check handler to a transactional check handler
func (cm *CheckManager) MakeCheckHandlerTransactional(checkID check.ID) CheckHandler {
	if !cm.config.CheckTransactionalityEnabled {
		_ = log.Warnf("Check transaction is disabled, defaulting %s to non-transactional check", checkID)
		return nil
	}

	ch, found := cm.checkHandlers[string(checkID)]
	if !found {
		_ = log.Errorf("No check handler found for %s.", checkID)
		return nil
	}

	config, initConfig := ch.GetConfig()
	transactionalCheckHandler := NewTransactionalCheckHandler(ch, config, initConfig)
	cm.checkHandlers[string(checkID)] = transactionalCheckHandler

	return transactionalCheckHandler
}

// RegisterCheckHandler registers a check handler for the given check using a transactionbatcher for this instance
func (cm *CheckManager) RegisterCheckHandler(check CheckIdentifier, config, initConfig integration.Data) CheckHandler {
	if !cmInitialized {
		_ = log.Errorf("CheckManager not initialized, initialize it using handler.InitCheckManager()")
		return nil
	}

	ch := cm.registerNonTransactionalCheckHandler(check, config, initConfig)
	log.Debugf("Registering Check Handler for: %s", ch.ID())
	cm.checkHandlers[string(check.ID())] = ch
	return ch
}

// UnsubscribeCheckHandler removes a check handler for the given check
func (cm *CheckManager) UnsubscribeCheckHandler(checkID check.ID) {
	log.Debugf("Removing Check Handler for: %s", checkID)
	delete(cm.checkHandlers, string(checkID))
}

// Stop clears the check handlers and re-initializes the singleton init
func (cm *CheckManager) Stop() {
	log.Debug("Removing all Check Handlers")
	cm.checkHandlers = make(map[string]CheckHandler)
	cmInit = new(sync.Once)
	cmInitialized = false
}
