package handler

import (
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/batcher"
	"github.com/DataDog/datadog-agent/pkg/collector/check/test"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionmanager"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckManagerSubscription(t *testing.T) {
	checkManager := NewCheckManager(batcher.NewMockBatcher(), transactionbatcher.NewMockTransactionalBatcher(), transactionmanager.NewMockTransactionManager())
	testCheck := &test.STSTestCheck{Name: "test-check-1"}

	// assert that we start at an empty state
	assert.EqualValues(t, checkManager.checkHandlers, map[string]CheckHandler{})

	// subscribe my test check, assert that we can get it with the check handler and that it's present in the inner map
	checkManager.RegisterCheckHandler(testCheck, integration.Data{1, 2, 3}, integration.Data{0, 0, 0})
	_, found := checkManager.checkHandlers[testCheck.String()]
	assert.True(t, found, "TestCheck handler not found in the checkManager.checkHandlers map")
	ch := checkManager.GetCheckHandler(testCheck.ID())
	assert.Equal(t, ch.ID(), testCheck.ID())
	actualInstanceCfg, actualInitCfg := ch.GetConfig()
	assert.EqualValues(t, integration.Data{1, 2, 3}, actualInstanceCfg)
	assert.EqualValues(t, integration.Data{0, 0, 0}, actualInitCfg)
	assert.Equal(t, "NonTransactionalCheckHandler", ch.Name())

	// subscribe another check handler and assert it
	testCheck2 := &test.STSTestCheck{Name: "test-check-2"}
	checkManager.RegisterCheckHandler(testCheck2, integration.Data{4, 5, 6}, integration.Data{10, 10, 10})
	_, found = checkManager.checkHandlers[testCheck.String()]
	assert.True(t, found, "TestCheck handler not found in the checkManager.checkHandlers map")
	ch2 := checkManager.GetCheckHandler(testCheck2.ID())
	assert.Equal(t, ch2.ID(), testCheck2.ID())
	actualInstanceCfg2, actualInitCfg2 := ch2.GetConfig()
	assert.EqualValues(t, integration.Data{4, 5, 6}, actualInstanceCfg2)
	assert.EqualValues(t, integration.Data{10, 10, 10}, actualInitCfg2)
	assert.Equal(t, "NonTransactionalCheckHandler", ch2.Name())

	// assert that we have 2 check handlers in the map
	assert.Equal(t, 2, len(checkManager.checkHandlers))

	// unsubscribe testCheck2, assert that checkManager.checkHandlers only contains 1 check handler and that testCheck2
	// is no longer present
	checkManager.UnsubscribeCheckHandler(testCheck2.ID())
	_, found = checkManager.checkHandlers[string(testCheck2.ID())]
	assert.False(t, found, "TestCheck handler not found in the checkManager.checkHandlers map")

	// subscribe testCheck2 again
	checkManager.RegisterCheckHandler(testCheck2, integration.Data{4, 5, 6}, integration.Data{10, 10, 10})
	_, found = checkManager.checkHandlers[string(testCheck2.ID())]
	assert.True(t, found, "TestCheck handler not found in the checkManager.checkHandlers map")

	// remove all check handlers with clear
	checkManager.Stop()
	assert.Equal(t, 0, len(checkManager.checkHandlers))

}
