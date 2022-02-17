package topologycollectors

import "github.com/StackVista/stackstate-agent/pkg/topology"

func testCaseName(baseName string, sourcePropertiesEnabled bool) string {
	if sourcePropertiesEnabled {
		return baseName + " with sourceProperties"
	} else {
		return baseName + " without sourceProperties"
	}
}

func chooseBySourcePropertiesFeature(
	sourcePropertiesEnabled bool,
	componentNoSP *topology.Component,
	componentSP *topology.Component) *topology.Component {
	if sourcePropertiesEnabled {
		return componentSP
	} else {
		return componentNoSP
	}
}
