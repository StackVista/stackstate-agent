package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadScenarios(t *testing.T) {
	scenarioName := "dockerd-eks"
	scenario := findScenario(scenarioName)

	assert.True(t, scenario.Name == scenarioName)
}

func TestCreateWorkspace(t *testing.T) {
	scenarioName := "dockerd-eks"
	runId := "randomId"
	create := findScenario(scenarioName).generateCreateStep(runId)

	assert.True(t, create.RunId() == fmt.Sprintf("%s-%s", runId, scenarioName))
}
