package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
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
	ReloadCheck()
	GetConfig() (config, initConfig integration.Data)
	GetBatcher() batcher.Batcher
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
}

func MakeCheckHandler(check check.Check, checkReloader CheckReloader, checkBatcher batcher.Batcher, config, initConfig integration.Data) CheckHandler {
	return &checkHandler{CheckHandlerBase{
		CheckIdentifier: check,
		CheckReloader:   checkReloader,
		Batcher:         checkBatcher,
		config:          config,
		initConfig:      initConfig,
	}}
}

// ReloadCheck ...
func (ch *checkHandler) ReloadCheck() {
	err := ch.CheckReloader.ReloadCheck(ch.ID(), ch.config, ch.initConfig, ch.ConfigSource())
	if err != nil {
		_ = log.Errorf("Error reloading check %s, %s", ch.String(), err)
		return
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
