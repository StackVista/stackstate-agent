// StackState
package topology

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/container_runtime"
)

const (
	containerType = "container"
)

type ContainerTopology interface {
	BuildContainerTopology(du *container_runtime.ContainerUtil) error
}
