package cmd

import (
	"beest/sut"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
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
	Name      string                 `yaml:"name"`
	Yard      Yard                   `yaml:"yard"`
	Test      Test                   `yaml:"test"`
	Variables map[string]interface{} `yaml:"variables"`
}

func (s *Scenario) mergeVars(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range append(maps, s.Variables) {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

type Scenarios struct {
	Scenarios []Scenario `yaml:"scenarios"`
}

///

func loadScenarios() *Scenarios {
	scenariosYaml, err := ioutil.ReadFile(ScenariosPath)
	if err != nil {
		log.Fatalf("Error while reading scenarios file: %s", err)
	}

	availableScenarios := &Scenarios{}
	err = yaml.UnmarshalStrict(scenariosYaml, availableScenarios)
	if err != nil {
		log.Fatalf("Error while unmarshalling scenarios: %s", err)
	}
	return availableScenarios
}

func choseScenario(name string) *Scenario {
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
