package topologycollectors

import "github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"

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
	if sourcePropertiesEnabled {
		if kubernetesStatusEnabled {
			return componentSPPlustStatus
		}
		return componentSP
	}
	return componentNoSP
}
