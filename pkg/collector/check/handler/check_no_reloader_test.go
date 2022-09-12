package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckNoReloader_ReloadCheck(t *testing.T) {
	err := CheckNoReloader{}.ReloadCheck("check-no-reloader", integration.Data{1, 2, 3}, integration.Data{4, 5, 6}, "")
	assert.NoError(t, err)
}
