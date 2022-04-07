package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckHandler(t *testing.T) {
	ch := MakeCheckHandler(&check.TestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
		batcher.MockBatcher{}, integration.Data{1, 2, 3}, integration.Data{0, 0, 0})

	assert.Equal(t, check.ID("my-check-handler-test-check"), ch.ID())
	assert.EqualValues(t, ch.GetBatcher(), batcher.MockBatcher{})
	actualInstanceCfg, actualInitCfg := ch.GetConfig()
	assert.EqualValues(t, integration.Data{1, 2, 3}, actualInstanceCfg)
	assert.EqualValues(t, integration.Data{0, 0, 0}, actualInitCfg)
	assert.Equal(t, "test-config-source", ch.ConfigSource())

	cr := ch.GetCheckReloader().(*check.TestCheckReloader)
	assert.Equal(t, 0, cr.Reloaded)
	ch.ReloadCheck()
	assert.Equal(t, 1, cr.Reloaded)
	ch.ReloadCheck()
	assert.Equal(t, 2, cr.Reloaded)
}
