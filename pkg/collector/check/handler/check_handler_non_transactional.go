package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/persistentcache"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// NonTransactionalCheckHandler is a wrapper for check that have no register handler.
type NonTransactionalCheckHandler struct {
	CheckIdentifier
	CheckReloader CheckReloader
}

// MakeNonTransactionalCheckHandler returns an instance of CheckHandler which functions as a fallback.
func MakeNonTransactionalCheckHandler(check CheckIdentifier, cr CheckReloader) CheckHandler {
	return &NonTransactionalCheckHandler{
		CheckIdentifier: check,
		CheckReloader:   cr,
	}
}

// SubmitComponent submits a component to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitComponent(instance topology.Instance, component topology.Component) {
	batcher.GetBatcher().SubmitComponent(ch.ID(), instance, component)
}

// SubmitRelation submits a relation to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitRelation(instance topology.Instance, relation topology.Relation) {
	batcher.GetBatcher().SubmitRelation(ch.ID(), instance, relation)
}

// SubmitStartSnapshot submits a start snapshot to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitStartSnapshot(instance topology.Instance) {
	batcher.GetBatcher().SubmitStartSnapshot(ch.ID(), instance)
}

// SubmitStopSnapshot submits a stop snapshot to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitStopSnapshot(instance topology.Instance) {
	batcher.GetBatcher().SubmitStopSnapshot(ch.ID(), instance)
}

// SubmitDelete submits a topology element delete to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitDelete(instance topology.Instance, topologyElementID string) {
	batcher.GetBatcher().SubmitDelete(ch.ID(), instance, topologyElementID)
}

// SubmitHealthCheckData submits health check data to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitHealthCheckData(stream health.Stream, data health.CheckData) {
	batcher.GetBatcher().SubmitHealthCheckData(ch.ID(), stream, data)
}

// SubmitHealthStartSnapshot submits a health start snapshot to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitHealthStartSnapshot(stream health.Stream, intervalSeconds int, expirySeconds int) {
	batcher.GetBatcher().SubmitHealthStartSnapshot(ch.ID(), stream, intervalSeconds, expirySeconds)
}

// SubmitHealthStopSnapshot submits a health stop snapshot to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitHealthStopSnapshot(stream health.Stream) {
	batcher.GetBatcher().SubmitHealthStopSnapshot(ch.ID(), stream)
}

// SubmitRawMetricsData submits a raw metric value to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitRawMetricsData(data telemetry.RawMetrics) {
	batcher.GetBatcher().SubmitRawMetricsData(ch.ID(), data)
}

// SubmitComplete submits a complete to the Global Batcher.
func (ch *NonTransactionalCheckHandler) SubmitComplete() {
	batcher.GetBatcher().SubmitComplete(ch.ID())
}

// Reload is a wrapper around the CheckReloader reload function
func (ch *NonTransactionalCheckHandler) Reload() {
	config, initConfig := ch.GetConfig()
	_ = ch.CheckReloader.ReloadCheck(ch.ID(), config, initConfig, ch.ConfigSource())
}

// StartTransaction logs a warning for the no check handler. This should never be called.
func (ch *NonTransactionalCheckHandler) StartTransaction() string {
	_ = log.Warnf("StartTransaction called on NonTransactionalCheckHandler. This should never happen.")
	return ""
}

// StopTransaction logs a warning for the no check handler. This should never be called.
func (ch *NonTransactionalCheckHandler) StopTransaction() {
	_ = log.Warnf("StopTransaction called on NonTransactionalCheckHandler. This should never happen.")
}

// SetTransactionState logs a warning for the no check handler. This should never be called.
func (ch *NonTransactionalCheckHandler) SetTransactionState(key string, state string) {
	_ = log.Warnf("SetTransactionState called on NonTransactionalCheckHandler. This should never happen.")
}

// GetState Get the last stored state for a certain key
func (ch *NonTransactionalCheckHandler) GetState(key string) string {
	state, err := persistentcache.Read(key)
	if err != nil {
		_ = log.Errorf("Unable to retrieve the agent check state for the key '%s', Received the following error: %s", key, err)
		return "{}"
	}

	return state
}

// SetState Set a new state for a certain key
func (ch *NonTransactionalCheckHandler) SetState(key string, state string) {
	err := persistentcache.Write(key, state)
	if err != nil {
		_ = log.Errorf("Unable to set the agent check state for the key '%s', Received the following error: %s", key, err)
	}
}

// GetConfig is the NonTransactionalCheckHandler implementation which just returns nil. This should never be called.
func (ch *NonTransactionalCheckHandler) GetConfig() (integration.Data, integration.Data) {
	_ = log.Warnf("GetConfig called on NonTransactionalCheckHandler. This should never happen.")
	return nil, nil
}

// GetBatcher returns the global batcher instance (non-transactional)
func (ch *NonTransactionalCheckHandler) GetBatcher() batcher.Batcher {
	return batcher.GetBatcher()
}

// GetCheckReloader returns the configured CheckReloader
func (ch *NonTransactionalCheckHandler) GetCheckReloader() CheckReloader {
	return ch.CheckReloader
}

// NewCheckIdentifier returns a IDOnlyCheckIdentifier that implements the CheckIdentifier interface for a given check ID.
func NewCheckIdentifier(checkID check.ID) CheckIdentifier {
	return &IDOnlyCheckIdentifier{checkID: checkID}
}

// IDOnlyCheckIdentifier is used to create a CheckIdentifier when only the checkID is present
type IDOnlyCheckIdentifier struct {
	checkID check.ID
}

// String returns the IDOnlyCheckIdentifier checkID as a string
func (idCI *IDOnlyCheckIdentifier) String() string {
	return string(idCI.checkID)
}

// ID returns the IDOnlyCheckIdentifier checkID
func (idCI *IDOnlyCheckIdentifier) ID() check.ID {
	return idCI.checkID
}

// ConfigSource returns an empty string for the IDOnlyCheckIdentifier
func (*IDOnlyCheckIdentifier) ConfigSource() string {
	return ""
}

// NoCheckReloader is a implementation of the CheckLoader interface that does a noop on ReloadCheck
type NoCheckReloader struct{}

// ReloadCheck returns nil
func (n NoCheckReloader) ReloadCheck(check.ID, integration.Data, integration.Data, string) error {
	return nil
}
