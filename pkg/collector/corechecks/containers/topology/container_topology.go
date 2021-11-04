// StackState
package topology

import (
	"github.com/StackVista/stackstate-agent/pkg/util"
)

type ContainerTopology interface {
	BuildContainerTopology(du *util.ContainerUtil) error
}
