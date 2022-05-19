package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// CheckNoHandler is a wrapper for check that have no register handler.
type CheckNoHandler struct {
	CheckID       check.ID
	CheckReloader CheckReloader
}

// MakeCheckNoHandler returns an instance of CheckHandler which functions as a fallback.
func MakeCheckNoHandler(checkID check.ID, cr CheckReloader) CheckHandler {
	return &CheckNoHandler{
		CheckID:       checkID,
		CheckReloader: cr,
	}
}

// String() returns the checkID + -name as a string.
func (ch *CheckNoHandler) String() string {
	return string(ch.CheckID) + "-name"
}

// ID returns the configured checkID.
func (ch *CheckNoHandler) ID() check.ID {
	return ch.CheckID
}

// ConfigSource returns no-source for the config source. This should never be called.
func (ch *CheckNoHandler) ConfigSource() string {
	return "no-source"
}

// SubmitComponent submits a component to the Global Batcher to be batched.
func (ch *CheckNoHandler) SubmitComponent(instance topology.Instance, component topology.Component) {
	batcher.GetBatcher().SubmitComponent(ch.ID(), instance, component)
}

// SubmitRelation submits a relation to the Global Batcher to be batched.
func (ch *CheckNoHandler) SubmitRelation(instance topology.Instance, relation topology.Relation) {
	batcher.GetBatcher().SubmitRelation(ch.ID(), instance, relation)
}

// SubmitStartSnapshot submits a start snapshot to the Global Batcher to be batched.
func (ch *CheckNoHandler) SubmitStartSnapshot(instance topology.Instance) {
	batcher.GetBatcher().SubmitStartSnapshot(ch.ID(), instance)
}

// SubmitStopSnapshot submits a stop snapshot to the Global Batcher to be batched.
func (ch *CheckNoHandler) SubmitStopSnapshot(instance topology.Instance) {
	batcher.GetBatcher().SubmitStopSnapshot(ch.ID(), instance)
}

// SubmitDelete submits a topology element delete to the Global Batcher to be batched.
func (ch *CheckNoHandler) SubmitDelete(instance topology.Instance, topologyElementID string) {
	batcher.GetBatcher().SubmitDelete(ch.ID(), instance, topologyElementID)
}

// SubmitHealthCheckData submits health check data to the Global Batcher to be batched.
func (ch *CheckNoHandler) SubmitHealthCheckData(stream health.Stream, data health.CheckData) {
	batcher.GetBatcher().SubmitHealthCheckData(ch.ID(), stream, data)
}

// SubmitHealthStartSnapshot submits a health start snapshot to the Global Batcher to be batched.
func (ch *CheckNoHandler) SubmitHealthStartSnapshot(stream health.Stream, intervalSeconds int, expirySeconds int) {
	batcher.GetBatcher().SubmitHealthStartSnapshot(ch.ID(), stream, intervalSeconds, expirySeconds)
}

// SubmitHealthStopSnapshot submits a health stop snapshot to the Global Batcher to be batched.
func (ch *CheckNoHandler) SubmitHealthStopSnapshot(stream health.Stream) {
	batcher.GetBatcher().SubmitHealthStopSnapshot(ch.ID(), stream)
}

// SubmitRawMetricsData submits a raw metric value to the Global Batcher to be batched.
func (ch *CheckNoHandler) SubmitRawMetricsData(data telemetry.RawMetrics) {
	batcher.GetBatcher().SubmitRawMetricsData(ch.ID(), data)
}

// SubmitComplete submits a complete to the Global Batcher.
func (ch *CheckNoHandler) SubmitComplete() {
	batcher.GetBatcher().SubmitComplete(ch.ID())
}

// Reload is a wrapper around the CheckReloader reload function
func (ch *CheckNoHandler) Reload() {
	config, initConfig := ch.GetConfig()
	_ = ch.CheckReloader.ReloadCheck(ch.ID(), config, initConfig, ch.ConfigSource())
}

// GetCurrentTransaction returns an empty string for the CheckNoHandler
func (ch *CheckNoHandler) GetCurrentTransaction() string {
	_ = log.Warnf("GetCurrentTransaction called on CheckNoHandler. This should never happen.")
	return ""
}

// SubmitStartTransaction logs a warning for the no check handler. This should never be called.
func (ch *CheckNoHandler) SubmitStartTransaction() {
	_ = log.Warnf("StartTransaction called on CheckNoHandler. This should never happen.")
}

// SubmitStopTransaction logs a warning for the no check handler. This should never be called.
func (ch *CheckNoHandler) SubmitStopTransaction() {
	_ = log.Warnf("SubmitStopTransaction called on CheckNoHandler. This should never happen.")
}

// GetConfig is the CheckNoHandler implementation which just returns nil. This should never be called.
func (ch *CheckNoHandler) GetConfig() (integration.Data, integration.Data) {
	_ = log.Warnf("GetConfig called on CheckNoHandler. This should never happen.")
	return nil, nil
}

// GetBatcher returns the global batcher instance (non-transactional)
func (ch *CheckNoHandler) GetBatcher() batcher.Batcher {
	return batcher.GetBatcher()
}

// GetCheckReloader returns the configured CheckReloader
func (ch *CheckNoHandler) GetCheckReloader() CheckReloader {
	return ch.CheckReloader
}

// NoCheckReloader is a implementation of the CheckLoader interface that does a noop on ReloadCheck
type NoCheckReloader struct{}

// ReloadCheck returns nil
func (n NoCheckReloader) ReloadCheck(check.ID, integration.Data, integration.Data, string) error {
	return nil
}
