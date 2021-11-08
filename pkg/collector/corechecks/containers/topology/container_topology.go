// StackState
package topology

import (
	"github.com/StackVista/stackstate-agent/pkg/util"
)

const (
	containerType = "container"
)

type ContainerTopology interface {
	BuildContainerTopology(du *util.ContainerUtil) error
}
