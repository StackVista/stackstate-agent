package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadScenarios(t *testing.T) {
	scenarioName := "dockerd-eks-1-21"
	scenario := findScenario(scenarioName)

	assert.True(t, scenario.Name == scenarioName)
}

func TestCreateWorkspace(t *testing.T) {
	scenarioName := "dockerd-eks-1-21"
	runId := "randomId"
	create := findScenario(scenarioName).generateCreateStep(runId)

	assert.Equal(t, fmt.Sprintf("%s:%s", "randomid", scenarioName), create.RunId())
	assert.Equal(t, fmt.Sprintf("beest-%s-%s", "randomid", scenarioName), create.Variables()["yard_id"])
}

func TestRunIdWithNonAlphanumerics(t *testing.T) {
	scenarioName := "dockerd-eks-1-21"
	runId := "r.and:omI-d"
	create := findScenario(scenarioName).generateCreateStep(runId)

	assert.Equal(t, fmt.Sprintf("%s:%s", "r-and-omi-d", scenarioName), create.RunId())
	assert.Equal(t, fmt.Sprintf("beest-%s-%s", "r-and-omi-d", scenarioName), create.Variables()["yard_id"])
}
