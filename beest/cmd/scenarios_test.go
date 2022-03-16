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

	assert.Equal(t, fmt.Sprintf("%s:%s", "randomId", scenarioName), create.RunId())
	assert.Equal(t, fmt.Sprintf("beest-%s-%s", runId, scenarioName), create.Variables()["yard_id"])
}

func TestRunIdWithNonAlphanumerics(t *testing.T) {
	scenarioName := "dockerd-eks"
	runId := "r.and:omI-d"
	create := findScenario(scenarioName).generateCreateStep(runId)

	assert.Equal(t, fmt.Sprintf("%s:%s", "randomId", scenarioName), create.RunId())
	assert.Equal(t, fmt.Sprintf("beest-%s-%s", "randomId", scenarioName), create.Variables()["yard_id"])
}
