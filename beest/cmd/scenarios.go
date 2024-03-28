package cmd

import (
	"beest/cmd/step"
	"beest/sut"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

type Yard struct {
	Name string `yaml:"name"`
}

func (y *Yard) path() string {
	return sut.YardPath(y.Name)
}

type Test struct {
	Group string `yaml:"group"`
}

func (t *Test) path() string {
	return sut.TestPath(t.Group)
}

type Scenario struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Yard        Yard                   `yaml:"yard"`
	Test        Test                   `yaml:"test"`
	Variables   map[string]interface{} `yaml:"variables"`
}

func (s *Scenario) generateCreateStep(runId string) *step.CreationStep {
	runVariables := make(map[string]interface{})
	for k, v := range s.Variables {
		runVariables[k] = v
	}
	normalizedRunId := strings.ToLower(removeNonAlphanumeric(runId))
	workspace := fmt.Sprintf("%s:%s", normalizedRunId, s.Name)
	runVariables["yard_id"] = fmt.Sprintf("beest-%s-%s", normalizedRunId, s.Name)
	runVariables["runners_ip"] = fmt.Sprintf(runnersIp)

	return step.Create(workspace, s.Yard.path(), s.Test.path(), runVariables)
}

type Scenarios struct {
	Scenarios []Scenario `yaml:"scenarios"`
}

///

func loadScenarios() *Scenarios {
	scenariosYaml, err := ioutil.ReadFile(ScenariosPath)
	if err != nil {
		log.Fatalf("Error reading scenarios file: %s", err)
	}

	availableScenarios := &Scenarios{}
	err = yaml.Unmarshal(scenariosYaml, availableScenarios)
	if err != nil {
		log.Fatalf("Error unmarshalling scenarios: %s", err)
	}
	return availableScenarios
}

func findScenario(name string) *Scenario {
	scenarios := loadScenarios()

	keys := make([]string, 0, len(scenarios.Scenarios))
	for _, s := range scenarios.Scenarios {
		keys = append(keys, s.Name)
		if s.Name == name {
			return &s
		}
	}

	log.Println(fmt.Sprintf("Available scenarios: %v", strings.Join(keys, ", ")))
	log.Fatalf("Scenario not found: %s", name)
	return nil
}

func removeNonAlphanumeric(target string) string {
	// Make a Regex to say we only want letters and numbers
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(target, "-")
}
