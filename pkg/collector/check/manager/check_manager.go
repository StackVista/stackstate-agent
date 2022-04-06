package manager

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/handler"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// CheckManager acts as the grouping of check handlers and deals with the "lifecycle" of check handlers.
type CheckManager struct {
	checkHandlers        map[string]handler.CheckHandler
	fallbackCheckHandler handler.CheckHandler
}

// MakeCheckManager returns a instance of the Check Manager
func MakeCheckManager() *CheckManager {
	return &CheckManager{
		checkHandlers:        make(map[string]handler.CheckHandler),
		fallbackCheckHandler: handler.MakeCheckNoHandler(),
	}
}

// GetCheckHandler returns a check handler (if found) for a given check ID
func (cm *CheckManager) GetCheckHandler(checkID check.ID) handler.CheckHandler {
	ch, found := cm.checkHandlers[string(checkID)]
	if !found {
		_ = log.Errorf(fmt.Sprintf("No check handler found for %s", checkID))
		return cm.fallbackCheckHandler
	}
	return ch
}

// SubscribeCheckHandler registers a check handler for the given check using a batcher for this instance
func (cm *CheckManager) SubscribeCheckHandler(check check.Check, checkReloader handler.CheckReloader,
	checkBatcher batcher.Batcher, config, initConfig integration.Data) handler.CheckHandler {
	ch := handler.MakeCheckHandler(check, checkReloader, checkBatcher, config, initConfig)
	log.Debugf("Registering Check Handler for: %s", ch.ID())
	cm.checkHandlers[string(check.ID())] = ch
	return ch
}

// UnsubscribeCheckHandler removes a check handler for the given check
func (cm *CheckManager) UnsubscribeCheckHandler(checkID check.ID) {
	log.Debugf("Removing Check Handler for: %s", checkID)
	delete(cm.checkHandlers, string(checkID))
}

// Clear removes any existing check handlers
func (cm *CheckManager) Clear() {
	log.Debug("Removing all Check Handlers")
	cm.checkHandlers = make(map[string]handler.CheckHandler)
}
