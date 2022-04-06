package manager

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/handler"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// FIXTURE
type TestCheck struct {
	Name string
}

func (c *TestCheck) String() string                                             { return c.Name }
func (c *TestCheck) Version() string                                            { return "" }
func (c *TestCheck) ConfigSource() string                                       { return "" }
func (c *TestCheck) Stop()                                                      {}
func (c *TestCheck) Configure(integration.Data, integration.Data, string) error { return nil }
func (c *TestCheck) Interval() time.Duration                                    { return 1 }
func (c *TestCheck) Run() error                                                 { return nil }
func (c *TestCheck) ID() check.ID                                               { return check.ID(c.String()) }
func (c *TestCheck) GetWarnings() []error                                       { return []error{} }
func (c *TestCheck) GetMetricStats() (map[string]int64, error)                  { return make(map[string]int64), nil }
func (c *TestCheck) IsTelemetryEnabled() bool                                   { return false }

type TestCheckReloader struct{}

func (_ TestCheckReloader) ReloadCheck(check.ID, integration.Data, integration.Data, string) error {
	return nil
}

func TestMakeCheckManager(t *testing.T) {
	checkManager := MakeCheckManager()
	expected := &CheckManager{
		checkHandlers:        make(map[string]handler.CheckHandler),
		fallbackCheckHandler: handler.MakeCheckNoHandler(),
	}

	assert.EqualValues(t, expected, checkManager)
}

func TestCheckManagerSubscription(t *testing.T) {
	checkManager := MakeCheckManager()
	testCheck := &TestCheck{Name: "test-check-1"}

	// assert that we start at an empty state
	assert.EqualValues(t, checkManager.checkHandlers, map[string]handler.CheckHandler{})
	assert.EqualValues(t, checkManager.fallbackCheckHandler, handler.MakeCheckNoHandler())

	// attempt to get a non-existing check handler, should default to the fallback
	assert.EqualValues(t, checkManager.GetCheckHandler(testCheck.ID()), checkManager.fallbackCheckHandler)

	// subscribe my test check, assert that we can get it with the check handler and that it's present in the inner map
	checkManager.SubscribeCheckHandler(testCheck, TestCheckReloader{}, batcher.MockBatcher{}, integration.Data{1, 2, 3}, integration.Data{0, 0, 0})
	_, found := checkManager.checkHandlers[string(testCheck.ID())]
	assert.True(t, found, "TestCheck handler not found in the checkManager.checkHandlers map")
	ch := checkManager.GetCheckHandler(testCheck.ID())
	assert.Equal(t, ch.ID(), testCheck.ID())
	assert.EqualValues(t, ch.GetBatcher(), batcher.MockBatcher{})
	actualInstanceCfg, actualInitCfg := ch.GetConfig()
	assert.EqualValues(t, integration.Data{1, 2, 3}, actualInstanceCfg)
	assert.EqualValues(t, integration.Data{0, 0, 0}, actualInitCfg)

	// subscribe another check handler and assert it
	testCheck2 := &TestCheck{Name: "test-check-2"}
	checkManager.SubscribeCheckHandler(testCheck2, TestCheckReloader{}, batcher.MockBatcher{}, integration.Data{4, 5, 6}, integration.Data{10, 10, 10})
	_, found = checkManager.checkHandlers[string(testCheck.ID())]
	assert.True(t, found, "TestCheck handler not found in the checkManager.checkHandlers map")
	ch2 := checkManager.GetCheckHandler(testCheck2.ID())
	assert.Equal(t, ch2.ID(), testCheck2.ID())
	assert.EqualValues(t, ch2.GetBatcher(), batcher.MockBatcher{})
	actualInstanceCfg2, actualInitCfg2 := ch2.GetConfig()
	assert.EqualValues(t, integration.Data{4, 5, 6}, actualInstanceCfg2)
	assert.EqualValues(t, integration.Data{10, 10, 10}, actualInitCfg2)

	// assert that we have 2 check handlers in the map
	assert.Equal(t, 2, len(checkManager.checkHandlers))

	// unsubscribe testCheck2, assert that checkManager.checkHandlers only contains 1 check handler and that testCheck2
	// is no longer present
	checkManager.UnsubscribeCheckHandler(testCheck2.ID())
	_, found = checkManager.checkHandlers[string(testCheck2.ID())]
	assert.False(t, found, "TestCheck handler not found in the checkManager.checkHandlers map")

	// subscribe testCheck2 again
	checkManager.SubscribeCheckHandler(testCheck2, TestCheckReloader{}, batcher.MockBatcher{}, integration.Data{4, 5, 6}, integration.Data{10, 10, 10})
	_, found = checkManager.checkHandlers[string(testCheck2.ID())]
	assert.True(t, found, "TestCheck handler not found in the checkManager.checkHandlers map")

	// remove all check handlers with clear
	checkManager.Clear()
	assert.Equal(t, 0, len(checkManager.checkHandlers))

}
