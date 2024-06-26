// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package clusteragent

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/StackVista/stackstate-agent/pkg/clusteragent/clusterchecks/types"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

const (
	dcaClusterChecksPath        = "api/v1/clusterchecks"
	dcaClusterChecksStatusPath  = dcaClusterChecksPath + "/status"
	dcaClusterChecksConfigsPath = dcaClusterChecksPath + "/configs"
)

// PostClusterCheckStatus is called by the clustercheck config provider
func (c *DCAClient) PostClusterCheckStatus(ctx context.Context, identifier string, status types.NodeStatus) (types.StatusResponse, error) {
	// Retry on the main URL if the leader fails
	willRetry := c.leaderClient.hasLeader()

	result, err := c.doPostClusterCheckStatus(ctx, identifier, status)
	if err != nil && willRetry {
		log.Debugf("Got error on leader, retrying via the service: %s", err)
		c.leaderClient.resetURL()
		return c.doPostClusterCheckStatus(ctx, identifier, status)
	}
	return result, err
}

func (c *DCAClient) doPostClusterCheckStatus(ctx context.Context, identifier string, status types.NodeStatus) (types.StatusResponse, error) {
	var response types.StatusResponse

	queryBody, err := json.Marshal(status)
	if err != nil {
		return response, err
	}

	// https://host:port/api/v1/clusterchecks/status/{identifier}
	rawURL := c.leaderClient.buildURL(dcaClusterChecksStatusPath, identifier)
	req, err := http.NewRequestWithContext(ctx, "POST", rawURL, bytes.NewBuffer(queryBody))
	if err != nil {
		return response, err
	}
	req.Header = c.clusterAgentAPIRequestHeaders

	resp, err := c.leaderClient.Do(req)
	if err != nil {
		return response, err
	}

	// sts
	err = parseJSONResponse(resp, &response)
	return response, err
}

// GetClusterCheckConfigs is called by the clustercheck config provider
func (c *DCAClient) GetClusterCheckConfigs(ctx context.Context, identifier string) (types.ConfigResponse, error) {
	// Retry on the main URL if the leader fails
	willRetry := c.leaderClient.hasLeader()

	result, err := c.doGetClusterCheckConfigs(ctx, identifier)
	if err != nil && willRetry {
		log.Debugf("Got error on leader, retrying via the service: %s", err)
		c.leaderClient.resetURL()
		return c.doGetClusterCheckConfigs(ctx, identifier)
	}
	return result, err
}

func (c *DCAClient) doGetClusterCheckConfigs(ctx context.Context, identifier string) (types.ConfigResponse, error) {
	var configs types.ConfigResponse
	var err error

	// https://host:port/api/v1/clusterchecks/configs/{identifier}
	rawURL := c.leaderClient.buildURL(dcaClusterChecksConfigsPath, identifier)
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return configs, err
	}
	req.Header = c.clusterAgentAPIRequestHeaders

	resp, err := c.leaderClient.Do(req)
	if err != nil {
		return configs, err
	}

	// sts
	err = parseJSONResponse(resp, &configs)
	return configs, err
}
