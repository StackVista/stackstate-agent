package handler

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"sync"
)

var (
	cmInstance *CheckManager
	cmInit     *sync.Once
)

func init() {
	cmInit = new(sync.Once)
}

// InitCheckManager ...
func InitCheckManager(reloader CheckReloader) {
	cmInit.Do(func() {
		cmInstance = newCheckManager(reloader)
	})
}

// GetCheckManager returns a handle on the global check checkmanager Instance
func GetCheckManager() *CheckManager {
	return cmInstance
}

// CheckManager acts as the grouping of check handlers and deals with the "lifecycle" of check handlers.
type CheckManager struct {
	reloader      CheckReloader
	checkHandlers map[string]CheckHandler
	config        Config
}

// newCheckManager returns a instance of the Check Manager
func newCheckManager(reloader CheckReloader) *CheckManager {
	return &CheckManager{
		reloader:      reloader,
		checkHandlers: make(map[string]CheckHandler),
		config:        GetCheckManagerConfig(),
	}
}

// GetCheckHandler returns a check handler (if found) for a given check ID
func (cm *CheckManager) GetCheckHandler(checkID check.ID) CheckHandler {
	ch, found := cm.checkHandlers[string(checkID)]
	if !found {
		_ = log.Errorf(fmt.Sprintf("No check handler found for %s. Registering a non-transactional check handler.", checkID))
		return cm.RegisterNonTransactionalCheckHandler(NewCheckIdentifier(checkID), nil, nil)
	}
	return ch
}

// RegisterNonTransactionalCheckHandler registers a non-transactional check handler for a given check
func (cm *CheckManager) RegisterNonTransactionalCheckHandler(check CheckIdentifier, config, initConfig integration.Data) CheckHandler {
	ch := MakeNonTransactionalCheckHandler(check, CheckNoReloader{}, config, initConfig)
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
	transactionalCheckHandler := NewTransactionalCheckHandler(ch, cm.reloader, config, initConfig)
	cm.checkHandlers[string(checkID)] = transactionalCheckHandler

	return transactionalCheckHandler
}

// RegisterCheckHandler registers a check handler for the given check using a transactionbatcher for this instance
func (cm *CheckManager) RegisterCheckHandler(check CheckIdentifier, config, initConfig integration.Data) CheckHandler {
	ch := cm.RegisterNonTransactionalCheckHandler(check, config, initConfig)
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
}
