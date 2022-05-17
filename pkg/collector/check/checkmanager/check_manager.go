package checkmanager

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/cmd/agent/common"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/handler"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"sync"
)

var (
	cmInstance *CheckManager
	cmInit     sync.Once
)

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
	checkHandlers map[string]handler.CheckHandler
}

// newCheckManager returns a instance of the Check Manager
func newCheckManager() *CheckManager {
	return &CheckManager{
		checkHandlers: make(map[string]handler.CheckHandler),
	}
}

// GetCheckHandler returns a check handler (if found) for a given check ID
func (cm *CheckManager) GetCheckHandler(checkID check.ID) handler.CheckHandler {
	ch, found := cm.checkHandlers[string(checkID)]
	if !found {
		_ = log.Errorf(fmt.Sprintf("No check handler found for %s", checkID))
		return handler.MakeCheckNoHandler(checkID, handler.NoCheckReloader{})
	}
	return ch
}

// RegisterCheckHandler registers a check handler for the given check using a transactionbatcher for this instance
func (cm *CheckManager) RegisterCheckHandler(check check.Check, config, initConfig integration.Data) handler.CheckHandler {
	ch := handler.NewCheckHandler(check, common.Coll, config, initConfig)
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
