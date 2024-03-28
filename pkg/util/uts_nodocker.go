//go:build !docker
// +build !docker

package util

import (
	"context"

	"github.com/StackVista/stackstate-agent/pkg/util/containers"
)

// GetAgentUTSMode retrieves from Docker the UTS mode of the Agent container
func GetAgentUTSMode(context.Context) (containers.UTSMode, error) {
	return containers.UnknownUTSMode, nil
}
