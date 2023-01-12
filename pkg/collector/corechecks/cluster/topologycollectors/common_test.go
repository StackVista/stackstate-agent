package topologycollectors

import "github.com/StackVista/stackstate-agent/pkg/topology"

func testCaseName(baseName string, sourcePropertiesEnabled bool, kubernetesStatusEnabled bool) string {
	if sourcePropertiesEnabled {
		if kubernetesStatusEnabled {
			baseName = baseName + " w/ sourceProps plus status"
		} else {
			baseName = baseName + " w/ sourceProps"
		}
	} else {
		baseName = baseName + " w/o sourceProps"
	}
	return baseName
}

func chooseBySourcePropertiesFeature(
	sourcePropertiesEnabled bool,
	kubernetesStatusEnabled bool,
	componentNoSP *topology.Component,
	componentSP *topology.Component,
	componentSPPlustStatus *topology.Component,
) *topology.Component {
	var result *topology.Component
	if sourcePropertiesEnabled {
		if kubernetesStatusEnabled {
			result = componentSPPlustStatus
		} else {
			result = componentSP
		}
	} else {
		result = componentNoSP
	}
	return result
}
