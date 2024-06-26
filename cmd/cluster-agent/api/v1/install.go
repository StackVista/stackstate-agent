// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package v1

import (
	"strconv"

	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/gorilla/mux"

	"github.com/StackVista/stackstate-agent/pkg/clusteragent"
)

var (
	apiRequests = telemetry.NewCounterWithOpts("", "api_requests",
		[]string{"handler", "status"}, "Counter of requests made to the cluster agent API.",
		telemetry.Options{NoDoubleUnderscoreSep: true})
)

func incrementRequestMetric(handler string, status int) {
	apiRequests.Inc(handler, strconv.Itoa(status))
}

// InstallMetadataEndpoints registers endpoints for metadata
func InstallMetadataEndpoints(r *mux.Router) {
	log.Debug("Registering metadata endpoints")
	if config.Datadog.GetBool("cloud_foundry") {
		installCloudFoundryMetadataEndpoints(r)
	} else {
		installKubernetesMetadataEndpoints(r)
	}
}

// InstallChecksEndpoints registers endpoints for cluster checks
func InstallChecksEndpoints(r *mux.Router, sc clusteragent.ServerContext) {
	log.Debug("Registering checks endpoints")
	installClusterCheckEndpoints(r, sc)
	installEndpointsCheckEndpoints(r, sc)
}
