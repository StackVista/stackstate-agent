package check

import (
	"errors"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// CheckIdentifier encapsulates all the functionality needed to describe and configure an agent check
type CheckIdentifier interface {
	String() string                                                     // provide a printable version of the check name
	Configure(config, initConfig integration.Data, source string) error // configure the check from the outside
	ID() ID                                                             // provide a unique identifier for every check instance
	Version() string                                                    // return the version of the check if available
	ConfigSource() string                                               // return the configuration source of the check
}

// CheckHandler provides an interface between the Agent Check and the transactional components
type CheckHandler struct {
	CheckIdentifier
	batcher.Batcher
}

func makeCheckHandler(check Check, checkBatcher batcher.Batcher) *CheckHandler {
	return &CheckHandler{
		CheckIdentifier: check,
		Batcher:         checkBatcher,
	}
}

// CheckManager acts as the grouping of check handlers and deals with the "lifecycle" of check handlers.
type CheckManager struct {
	checkHandlers map[string]*CheckHandler
}

// MakeCheckManager returns a instance of the Check Manager
func MakeCheckManager() *CheckManager {
	return &CheckManager{}
}

// GetCheckHandler returns a check handler (if found) for a given check ID
func (cm *CheckManager) GetCheckHandler(checkID ID) (*CheckHandler, error) {
	ch, found := cm.checkHandlers[string(checkID)]
	if !found {
		return nil, errors.New(fmt.Sprintf("No check handler found for %s", checkID))
	}
	return ch, nil
}

// SubscribeCheckHandler registers a check handler for the given check using a batcher for this instance
func (cm *CheckManager) SubscribeCheckHandler(check Check, checkBatcher batcher.Batcher) *CheckHandler {
	ch := makeCheckHandler(check, checkBatcher)
	log.Debugf("Registering Check Handler for: %s", ch.ID())
	cm.checkHandlers[string(check.ID())] = makeCheckHandler(check, checkBatcher)
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
	cm.checkHandlers = make(map[string]*CheckHandler)
}
