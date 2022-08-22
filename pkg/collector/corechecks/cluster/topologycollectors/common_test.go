package topologycollectors

import "github.com/StackVista/stackstate-agent/pkg/topology"

func testCaseName(baseName string, sourcePropertiesEnabled bool) string {
	if sourcePropertiesEnabled {
		baseName = baseName + " w/ sourceProps"
	} else {
		baseName = baseName + " w/o sourceProps"
	}
	return baseName
}

func chooseBySourcePropertiesFeature(
	sourcePropertiesEnabled bool,
	componentNoSP *topology.Component,
	componentSP *topology.Component,
) *topology.Component {
	if sourcePropertiesEnabled {
		return componentSP
	}
	return componentNoSP
}
