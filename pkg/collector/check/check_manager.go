package check

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// CheckManager acts as the grouping of check handlers and deals with the "lifecycle" of check handlers.
type CheckManager struct {
	checkHandlers        map[string]CheckHandler
	fallbackCheckHandler CheckHandler
}

// MakeCheckManager returns a instance of the Check Manager
func MakeCheckManager() *CheckManager {
	return &CheckManager{
		checkHandlers:        make(map[string]CheckHandler),
		fallbackCheckHandler: MakeCheckNoHandler(),
	}
}

// GetCheckHandler returns a check handler (if found) for a given check ID
func (cm *CheckManager) GetCheckHandler(checkID ID) CheckHandler {
	ch, found := cm.checkHandlers[string(checkID)]
	if !found {
		_ = log.Errorf(fmt.Sprintf("No check handler found for %s", checkID))
		return cm.fallbackCheckHandler
	}
	return ch
}

// SubscribeCheckHandler registers a check handler for the given check using a batcher for this instance
func (cm *CheckManager) SubscribeCheckHandler(check Check, checkBatcher batcher.Batcher, config, initConfig integration.Data) CheckHandler {
	ch := MakeCheckHandler(check, checkBatcher, config, initConfig)
	log.Debugf("Registering Check Handler for: %s", ch.ID())
	cm.checkHandlers[string(check.ID())] = ch
	return ch
}

// UnsubscribeCheckHandler removes a check handler for the given check
func (cm *CheckManager) UnsubscribeCheckHandler(check Check) {
	log.Debugf("Removing Check Handler for: %s", check.ID())
	delete(cm.checkHandlers, string(check.ID()))
}

// Clear removes any existing check handlers
func (cm CheckManager) Clear() {
	log.Debug("Removing all Check Handlers")
	cm.checkHandlers = make(map[string]CheckHandler)
}
