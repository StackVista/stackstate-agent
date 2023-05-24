package gce

import (
	"context"
	"fmt"
	"strings"

	"github.com/StackVista/stackstate-agent/pkg/util/hostname"
)

// GetHostname returns the GCE cloud specific hostname from the k8s providerId. This must be compatible with
// github.com/StackVista/stackstate-agent/pkg/util/cloudproviders/gce
// Example providerId: gce://test-stackstate/europe-west4-a/gke-test-default-pool-9f8f65a4-2kld
// FQDN's use this pattern VM_NAME.ZONE.c.PROJECT_ID.internal (from https://cloud.google.com/compute/docs/internal-dns)
func GetHostname(providerID string) string {
	parts := strings.Split(strings.TrimPrefix(providerID, "gce://"), "/")
	if len(parts) != 3 {
		return ""
	}

	projectID := parts[0]
	zone := parts[1]
	vmName := parts[2]

	isZoned, err := IsZonedHostname()
	if err != nil {
		return ""
	}

	if isZoned {
		return fmt.Sprintf("%s.%s.c.%s.internal", vmName, zone, projectID)
	} else {
		return fmt.Sprintf("%s.c.%s.internal", vmName, projectID)
	}
}

// IsZonedHostname returns true if the hostname of the agent is a zoned hostname
func IsZonedHostname() (bool, error) {
	agentHostname, err := hostname.GetHostname(context.Background(), "gce", map[string]interface{}{})
	if err != nil {
		return false, err
	}

	hostnameParts := strings.Split(agentHostname, ".")

	return len(hostnameParts) == 5, nil
}
