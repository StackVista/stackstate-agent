package py

import (
	"github.com/StackVista/stackstate-agent/pkg/aggregator/mocksender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSubmitTopologyEvent(t *testing.T) {
	check, _ := getCheckInstance("testtelemetry", "TestTopologyEvents")

	mockSender := mocksender.NewMockSender(check.ID())

	mockSender.On("Event", mock.AnythingOfType("metrics.Event")).Return().Times(10)
	mockSender.On("Commit").Return().Times(1)

	err := check.Run()
	assert.Nil(t, err)
}
