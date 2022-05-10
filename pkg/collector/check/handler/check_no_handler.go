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

// CheckNoHandler ...
type CheckNoHandler struct {
	CheckID       check.ID
	CheckReloader CheckReloader
}

// MakeCheckNoHandler returns an instance of CheckHandler which functions as a fallback
func MakeCheckNoHandler(checkID check.ID, cr CheckReloader) CheckHandler {
	return &CheckNoHandler{
		CheckID:       checkID,
		CheckReloader: cr,
	}
}

func (ch *CheckNoHandler) String() string {
	return string(ch.CheckID) + "-name"
}

func (ch *CheckNoHandler) ID() check.ID {
	return ch.CheckID
}

func (ch *CheckNoHandler) ConfigSource() string {
	return "no-source"
}

func (ch *CheckNoHandler) SubmitComponent(instance topology.Instance, component topology.Component) {
	batcher.GetBatcher().SubmitComponent(ch.ID(), instance, component)
}

func (ch *CheckNoHandler) SubmitRelation(instance topology.Instance, relation topology.Relation) {
	batcher.GetBatcher().SubmitRelation(ch.ID(), instance, relation)
}

func (ch *CheckNoHandler) SubmitStartSnapshot(instance topology.Instance) {
	batcher.GetBatcher().SubmitStartSnapshot(ch.ID(), instance)
}

func (ch *CheckNoHandler) SubmitStopSnapshot(instance topology.Instance) {
	batcher.GetBatcher().SubmitStopSnapshot(ch.ID(), instance)
}

func (ch *CheckNoHandler) SubmitDelete(instance topology.Instance, topologyElementID string) {
	batcher.GetBatcher().SubmitDelete(ch.ID(), instance, topologyElementID)
}

func (ch *CheckNoHandler) SubmitHealthCheckData(stream health.Stream, data health.CheckData) {
	batcher.GetBatcher().SubmitHealthCheckData(ch.ID(), stream, data)
}

func (ch *CheckNoHandler) SubmitHealthStartSnapshot(stream health.Stream, intervalSeconds int, expirySeconds int) {
	batcher.GetBatcher().SubmitHealthStartSnapshot(ch.ID(), stream, intervalSeconds, expirySeconds)
}

func (ch *CheckNoHandler) SubmitHealthStopSnapshot(stream health.Stream) {
	batcher.GetBatcher().SubmitHealthStopSnapshot(ch.ID(), stream)
}

func (ch *CheckNoHandler) SubmitRawMetricsData(data telemetry.RawMetrics) {
	batcher.GetBatcher().SubmitRawMetricsData(ch.ID(), data)
}

func (ch *CheckNoHandler) SubmitComplete() {
	batcher.GetBatcher().SubmitComplete(ch.ID())
}

func (ch *CheckNoHandler) Shutdown() {
	batcher.GetBatcher().Shutdown()
}

func (ch *CheckNoHandler) Reload() {
	config, initConfig := ch.GetConfig()
	_ = ch.CheckReloader.ReloadCheck(ch.ID(), config, initConfig, ch.ConfigSource())
}

// StartTransaction ...
func (ch *CheckNoHandler) StartTransaction(check.ID, string) {
	_ = log.Warnf("StartTransaction called on CheckNoHandler. This should never happen.")
}

// GetConfig is the CheckNoHandler implementation which just returns nil. This should never be called.
func (ch *CheckNoHandler) GetConfig() (integration.Data, integration.Data) {
	_ = log.Warnf("GetConfig called on CheckNoHandler. This should never happen.")
	return nil, nil
}

func (ch *CheckNoHandler) GetBatcher() batcher.Batcher {
	return batcher.GetBatcher()
}

func (ch *CheckNoHandler) GetCheckReloader() CheckReloader {
	return ch.CheckReloader
}

type NoCheckReloader struct{}

func (n NoCheckReloader) ReloadCheck(id check.ID, config, initConfig integration.Data, newSource string) error {
	return nil
}
