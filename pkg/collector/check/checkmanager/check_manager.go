package checkmanager

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/handler"
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
func InitCheckManager(reloader handler.CheckReloader) {
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
	reloader      handler.CheckReloader
	checkHandlers map[string]handler.CheckHandler
	config        Config
}

// newCheckManager returns a instance of the Check Manager
func newCheckManager(reloader handler.CheckReloader) *CheckManager {
	return &CheckManager{
		reloader:      reloader,
		checkHandlers: make(map[string]handler.CheckHandler),
		config:        GetCheckManagerConfig(),
	}
}

// GetCheckHandler returns a check handler (if found) for a given check ID
func (cm *CheckManager) GetCheckHandler(checkID check.ID) handler.CheckHandler {
	ch, found := cm.checkHandlers[string(checkID)]
	if !found {
		_ = log.Errorf(fmt.Sprintf("No check handler found for %s. Registering a non-transactional check handler.", checkID))
		return cm.RegisterNonTransactionalCheckHandler(handler.NewCheckIdentifier(checkID))
	}
	return ch
}

// RegisterNonTransactionalCheckHandler registers a non-transactional check handler for a given check
func (cm *CheckManager) RegisterNonTransactionalCheckHandler(check handler.CheckIdentifier) handler.CheckHandler {
	ch := handler.MakeNonTransactionalCheckHandler(check, handler.NoCheckReloader{})
	cm.checkHandlers[string(check.ID())] = ch
	return ch
}

// RegisterCheckHandler registers a check handler for the given check using a transactionbatcher for this instance
func (cm *CheckManager) RegisterCheckHandler(check check.Check, config, initConfig integration.Data) handler.CheckHandler {
	if !cm.config.CheckTransactionalityEnabled {
		log.Debugf("Check transaction is disabled, defaulting %s to non-transactional check", check.ID())
		return cm.RegisterNonTransactionalCheckHandler(check)
	}

	ch := handler.NewCheckHandler(check, cm.reloader, config, initConfig)
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
	cm.checkHandlers = make(map[string]handler.CheckHandler)
	cmInit = new(sync.Once)
}
