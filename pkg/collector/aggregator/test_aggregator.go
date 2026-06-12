// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build test

package aggregator

import (
	"github.com/DataDog/datadog-agent/comp/core/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/batcher"
	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"
	"github.com/DataDog/datadog-agent/pkg/collector/check/test"
	check2 "github.com/StackVista/stackstate-receiver-go-client/pkg/model/check"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/health"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/telemetry"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionmanager"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	tagger "github.com/DataDog/datadog-agent/comp/core/tagger/def"
	nooptagger "github.com/DataDog/datadog-agent/comp/core/tagger/impl-noop"
	workloadfilter "github.com/DataDog/datadog-agent/comp/core/workloadfilter/def"
	workloadfilterfxmock "github.com/DataDog/datadog-agent/comp/core/workloadfilter/fx-mock"
	integrations "github.com/DataDog/datadog-agent/comp/logs/integrations/def"
	pkgaggregator "github.com/DataDog/datadog-agent/pkg/aggregator" // [sts] for NewNoOpSenderManager in inlined scopeInitCheckManager
	"github.com/DataDog/datadog-agent/pkg/aggregator/mocksender"
	"github.com/DataDog/datadog-agent/pkg/aggregator/sender"
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/DataDog/datadog-agent/pkg/metrics/event"
	"github.com/DataDog/datadog-agent/pkg/metrics/servicecheck"
	"github.com/DataDog/datadog-agent/pkg/util/option"
)

/*
#include "rtloader_types.h"
*/
import "C"

func testSubmitMetric(t *testing.T) {
	sender := mocksender.NewMockSender(checkid.ID("testID"))
	logReceiver := option.None[integrations.Component]()
	tagger := nooptagger.NewComponent()
	filterStore := workloadfilterfxmock.SetupMockFilter(t)
	release := ScopeInitCheckContext(sender.GetSenderManager(), logReceiver, tagger, filterStore)
	defer release()

	sender.SetupAcceptAll()

	cTags := []*C.char{C.CString("tag1"), C.CString("tag2"), nil}
	SubmitMetric(C.CString("testID"),
		C.DATADOG_AGENT_RTLOADER_GAUGE,
		C.CString("test_gauge"),
		C.double(21),
		&cTags[0],
		C.CString("my_hostname"),
		C.bool(false))
	SubmitMetric(C.CString("testID"),
		C.DATADOG_AGENT_RTLOADER_RATE,
		C.CString("test_rate"),
		C.double(21),
		&cTags[0],
		C.CString("my_hostname"),
		C.bool(false))
	SubmitMetric(C.CString("testID"),
		C.DATADOG_AGENT_RTLOADER_COUNT,
		C.CString("test_count"),
		C.double(21),
		&cTags[0],
		C.CString("my_hostname"),
		C.bool(false))
	SubmitMetric(C.CString("testID"),
		C.DATADOG_AGENT_RTLOADER_MONOTONIC_COUNT,
		C.CString("test_monotonic_count"),
		C.double(21),
		&cTags[0],
		C.CString("my_hostname"),
		C.bool(false))
	SubmitMetric(C.CString("testID"),
		C.DATADOG_AGENT_RTLOADER_MONOTONIC_COUNT,
		C.CString("test_monotonic_count_flush_first_value"),
		C.double(21),
		&cTags[0],
		C.CString("my_hostname"),
		C.bool(true))
	SubmitMetric(C.CString("testID"),
		C.DATADOG_AGENT_RTLOADER_COUNTER,
		C.CString("test_counter"),
		C.double(21),
		&cTags[0],
		C.CString("my_hostname"),
		C.bool(false))
	SubmitMetric(C.CString("testID"),
		C.DATADOG_AGENT_RTLOADER_HISTOGRAM,
		C.CString("test_histogram"),
		C.double(21),
		&cTags[0],
		C.CString("my_hostname"),
		C.bool(false))
	SubmitMetric(C.CString("testID"),
		C.DATADOG_AGENT_RTLOADER_HISTORATE,
		C.CString("test_historate"),
		C.double(21),
		&cTags[0],
		C.CString("my_hostname"),
		C.bool(false))

	sender.AssertMetric(t, "Gauge", "test_gauge", 21, "my_hostname", []string{"tag1", "tag2"})
	sender.AssertMetric(t, "Rate", "test_rate", 21, "my_hostname", []string{"tag1", "tag2"})
	sender.AssertMetric(t, "Count", "test_count", 21, "my_hostname", []string{"tag1", "tag2"})
	sender.AssertMonotonicCount(t, "MonotonicCountWithFlushFirstValue", "test_monotonic_count", 21, "my_hostname", []string{"tag1", "tag2"}, false)
	sender.AssertMonotonicCount(t, "MonotonicCountWithFlushFirstValue", "test_monotonic_count_flush_first_value", 21, "my_hostname", []string{"tag1", "tag2"}, true)
	sender.AssertMetric(t, "Counter", "test_counter", 21, "my_hostname", []string{"tag1", "tag2"})
	sender.AssertMetric(t, "Histogram", "test_histogram", 21, "my_hostname", []string{"tag1", "tag2"})
	sender.AssertMetric(t, "Historate", "test_historate", 21, "my_hostname", []string{"tag1", "tag2"})
}

func testSubmitMetricEmptyTags(t *testing.T) {
	sender := mocksender.NewMockSender(checkid.ID("testID"))
	logReceiver := option.None[integrations.Component]()
	tagger := nooptagger.NewComponent()
	filterStore := workloadfilterfxmock.SetupMockFilter(t)
	release := ScopeInitCheckContext(sender.GetSenderManager(), logReceiver, tagger, filterStore)
	defer release()

	sender.SetupAcceptAll()

	cTags := []*C.char{nil}
	SubmitMetric(C.CString("testID"),
		C.DATADOG_AGENT_RTLOADER_GAUGE,
		C.CString("test_gauge"),
		C.double(21),
		&cTags[0],
		C.CString("my_hostname"),
		C.bool(false))

	sender.AssertMetric(t, "Gauge", "test_gauge", 21, "my_hostname", nil)
}

func testSubmitMetricEmptyHostname(t *testing.T) {
	sender := mocksender.NewMockSender(checkid.ID("testID"))
	logReceiver := option.None[integrations.Component]()
	tagger := nooptagger.NewComponent()
	filterStore := workloadfilterfxmock.SetupMockFilter(t)
	release := ScopeInitCheckContext(sender.GetSenderManager(), logReceiver, tagger, filterStore)
	defer release()

	sender.SetupAcceptAll()

	cTags := []*C.char{nil}
	SubmitMetric(C.CString("testID"),
		C.DATADOG_AGENT_RTLOADER_GAUGE,
		C.CString("test_gauge"),
		C.double(21),
		&cTags[0],
		nil,
		C.bool(false))

	sender.AssertMetric(t, "Gauge", "test_gauge", 21, "", nil)
}

func testSubmitServiceCheck(t *testing.T) {
	sender := mocksender.NewMockSender(checkid.ID("testID"))
	logReceiver := option.None[integrations.Component]()
	tagger := nooptagger.NewComponent()
	filterStore := workloadfilterfxmock.SetupMockFilter(t)
	release := ScopeInitCheckContext(sender.GetSenderManager(), logReceiver, tagger, filterStore)
	defer release()

	sender.SetupAcceptAll()

	cTags := []*C.char{C.CString("tag1"), C.CString("tag2"), nil}
	SubmitServiceCheck(C.CString("testID"),
		C.CString("service_name"),
		C.int(1),
		&cTags[0],
		C.CString("my_hostname"),
		C.CString("my_message"))

	sender.AssertServiceCheck(t, "service_name", servicecheck.ServiceCheckWarning, "my_hostname", []string{"tag1", "tag2"}, "my_message")
}

func testSubmitServiceCheckEmptyTag(t *testing.T) {
	sender := mocksender.NewMockSender(checkid.ID("testID"))
	logReceiver := option.None[integrations.Component]()
	tagger := nooptagger.NewComponent()
	filterStore := workloadfilterfxmock.SetupMockFilter(t)
	release := ScopeInitCheckContext(sender.GetSenderManager(), logReceiver, tagger, filterStore)
	defer release()

	sender.SetupAcceptAll()

	cTags := []*C.char{nil}
	SubmitServiceCheck(C.CString("testID"),
		C.CString("service_name"),
		C.int(1),
		&cTags[0],
		C.CString("my_hostname"),
		C.CString("my_message"))

	sender.AssertServiceCheck(t, "service_name", servicecheck.ServiceCheckWarning, "my_hostname", nil, "my_message")
}

func testSubmitServiceCheckEmptyHostame(t *testing.T) {
	sender := mocksender.NewMockSender(checkid.ID("testID"))
	logReceiver := option.None[integrations.Component]()
	tagger := nooptagger.NewComponent()
	filterStore := workloadfilterfxmock.SetupMockFilter(t)
	release := ScopeInitCheckContext(sender.GetSenderManager(), logReceiver, tagger, filterStore)
	defer release()

	sender.SetupAcceptAll()

	cTags := []*C.char{nil}
	SubmitServiceCheck(C.CString("testID"),
		C.CString("service_name"),
		C.int(1),
		&cTags[0],
		nil,
		C.CString("my_message"))

	sender.AssertServiceCheck(t, "service_name", servicecheck.ServiceCheckWarning, "", nil, "my_message")
}

func testSubmitEvent(t *testing.T) {
	_, mockTransactionalBatcher, _, manager := handler.SetupMockTransactionalComponents()

	// [sts] Inlined scopeInitCheckManager (originally pkg/collector/python/test_util.go:99).
	// Cleans any prior context, then locks a fresh CheckContext bound to `manager`. Returns
	// releaseCheckContext as the cleanup func. Adapted to aggregator's 5-arg withLockedCheckContext
	// (filterStore arg added in 7.78.2) and aggregator-package-private vars.
	release := func() func() {
		checkContextMutex.Lock()
		if checkCtx != nil {
			checkCtx.checkManager.Stop()
			checkCtx = nil
		}
		checkContextMutex.Unlock()
		withLockedCheckContext(pkgaggregator.NewNoOpSenderManager(), manager, option.None[integrations.Component](), nooptagger.NewComponent(), workloadfilterfxmock.SetupMockFilter(t))
		return releaseCheckContext
	}()
	defer release()

	testCheck := &test.STSTestCheck{Name: "check-id-event-test"}
	manager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	ev := C.event_t{}
	ev.title = C.CString("ev_title")
	ev.text = C.CString("ev_text")
	ev.ts = 21
	ev.priority = C.CString("ev_priority")
	ev.host = C.CString("ev_host")
	ev.alert_type = C.CString("alert_type")
	ev.aggregation_key = C.CString("aggregation_key")
	ev.source_type_name = C.CString("source_type")
	ev.event_type = C.CString("event_type")
	tags := []*C.char{C.CString("tag1"), C.CString("tag2"), nil}
	ev.tags = &tags[0]

	checkId := C.CString(testCheck.String())

	// [sts] Inlined StartTransaction (originally pkg/collector/python/transactional_api.go:31).
	// pkg/collector/python imports pkg/collector/aggregator, so we can't import it back without
	// a cycle. Direct CheckContext access works since we're now in the aggregator package.
	if cc, err := GetCheckContext(); err == nil {
		cc.checkManager.GetCheckHandler(checkid.ID(C.GoString(checkId))).StartTransaction()
	}
	SubmitEvent(checkId, &ev)

	expectedEvent := event.Event{
		Title:          "ev_title",
		Text:           "ev_text",
		Ts:             21,
		Priority:       "ev_priority",
		Host:           "ev_host",
		Tags:           []string{"tag1", "tag2"},
		AlertType:      "alert_type",
		AggregationKey: "aggregation_key",
		SourceTypeName: "source_type",
	}

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	currentCheckState, found := mockTransactionalBatcher.GetCheckState(check2.CheckID(testCheck.ID()))
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())

	expectedCheckState := transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: currentCheckState.Transaction, // not asserting this specifically, it just needs to be present
		Health:      map[string]health.Health{},
		Events:      &telemetry.IntakeEvents{Events: []telemetry.Event{handler.ConvertToStsEvent(expectedEvent)}},
	}
	assert.Equal(t, expectedCheckState, currentCheckState)

	manager.UnsubscribeCheckHandler(testCheck.ID())
}

func testSubmitHistogramBucket(t *testing.T) {
	sender := mocksender.NewMockSender(checkid.ID("testID"))
	logReceiver := option.None[integrations.Component]()
	tagger := nooptagger.NewComponent()
	filterStore := workloadfilterfxmock.SetupMockFilter(t)
	release := ScopeInitCheckContext(sender.GetSenderManager(), logReceiver, tagger, filterStore)
	defer release()

	sender.SetupAcceptAll()

	cTags := []*C.char{C.CString("tag1"), C.CString("tag2"), nil}
	SubmitHistogramBucket(
		C.CString("testID"),
		C.CString("test_histogram"),
		C.longlong(42),
		C.float(1.0),
		C.float(2.0),
		C.int(1),
		C.CString("my_hostname"),
		&cTags[0],
		true,
	)

	sender.AssertHistogramBucket(t, "HistogramBucket", "test_histogram", 42, 1.0, 2.0, true, "my_hostname", []string{"tag1", "tag2"}, true)
}

func testSubmitEventPlatformEvent(t *testing.T) {
	sender := mocksender.NewMockSender("testID")
	logReceiver := option.None[integrations.Component]()
	tagger := nooptagger.NewComponent()
	filterStore := workloadfilterfxmock.SetupMockFilter(t)
	release := ScopeInitCheckContext(sender.GetSenderManager(), logReceiver, tagger, filterStore)
	defer release()

	sender.SetupAcceptAll()
	SubmitEventPlatformEvent(
		C.CString("testID"),
		C.CString("raw-event"),
		C.int(len("raw-event")),
		C.CString("dbm-sample"),
	)

	sender.AssertEventPlatformEvent(t, []byte("raw-event"), "dbm-sample")
}

// ScopeInitCheckContext initializes a check context and returns a release
// function the caller must defer.
// [sts] Function added during the 7.71.2 -> 7.78.2 merge to inline the test
// helper that DD removed from python/test_loader.go. Defensive cleanup of any
// previous test's leaked context lives inside withLockedCheckContext now (so
// the testMutex is acquired first and only released by the returned function).
func ScopeInitCheckContext(senderManager sender.SenderManager, logReceiver option.Option[integrations.Component], taggerComp tagger.Component, filterStore workloadfilter.Component) func() {
	checkManager := handler.NewCheckManager(batcher.NewMockBatcher(), transactionbatcher.NewMockTransactionalBatcher(), transactionmanager.NewMockTransactionManager())
	withLockedCheckContext(senderManager, checkManager, logReceiver, taggerComp, filterStore)
	return releaseCheckContext
}

// ScopeInitCheckContextWithCheckManager is the same as ScopeInitCheckContext but
// lets the caller supply a check manager directly (rather than using a default mock
// over the standard mock batchers). Returns a release function the caller must defer.
// [sts] Added for STS Python API tests (pkg/collector/python/test_*_api.go) that
// need to inject custom CheckManager mocks per test. Lets them migrate off the
// legacy pkg/collector/python/check_context.go scaffolding and route through the
// canonical aggregator.CheckContext (STAC-24699).
func ScopeInitCheckContextWithCheckManager(senderManager sender.SenderManager, checkManager handler.CheckManager, logReceiver option.Option[integrations.Component], taggerComp tagger.Component, filterStore workloadfilter.Component) func() {
	withLockedCheckContext(senderManager, checkManager, logReceiver, taggerComp, filterStore)
	return releaseCheckContext
}
