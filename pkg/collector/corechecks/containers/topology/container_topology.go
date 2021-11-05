// StackState
package topology

import (
	"github.com/StackVista/stackstate-agent/pkg/util/containers"
)

const (
	containerType = "container"
)

type ContainerTopology interface {
	BuildContainerTopology(du *containers.ContainerUtil) error
}
