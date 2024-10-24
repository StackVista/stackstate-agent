// +build docker

package util

import (
	"context"
	"fmt"

	"github.com/StackVista/stackstate-agent/pkg/util/cache"
	"github.com/StackVista/stackstate-agent/pkg/util/docker"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// GetAgentNetworkMode retrieves from Docker the network mode of the Agent container
func GetAgentNetworkMode(ctx context.Context) (string, error) {
	cacheNetworkModeKey := cache.BuildAgentKey("networkMode")
	if cacheNetworkMode, found := cache.Cache.Get(cacheNetworkModeKey); found {
		return cacheNetworkMode.(string), nil
	}

	log.Debugf("GetAgentNetworkMode trying Docker")
	networkMode, err := docker.GetAgentContainerNetworkMode(ctx)
	cache.Cache.Set(cacheNetworkModeKey, networkMode, cache.NoExpiration)
	if err != nil {
		return networkMode, fmt.Errorf("could not detect agent network mode: %v", err)
	}
	log.Debugf("GetAgentNetworkMode: using network mode from Docker: %s", networkMode)
	return networkMode, nil
}
