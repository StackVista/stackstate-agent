package gce

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
)

// GetHostname returns the GCE cloud specific hostname from the k8s providerId. This must be compatible with
// github.com/StackVista/stackstate-agent/pkg/util/cloudproviders/gce
// Example providerId: gce://test-stackstate/europe-west4-a/gke-test-default-pool-9f8f65a4-2kld
// FQDN's use this pattern VM_NAME.ZONE.c.PROJECT_ID.internal (from https://cloud.google.com/compute/docs/internal-dns)
func GetHostname(node v1.Node) string {
	parts := strings.Split(strings.TrimPrefix(node.Spec.ProviderID, "gce://"), "/")
	if len(parts) != 3 {
		return ""
	}
	projectID := parts[0]
	zone := parts[1]
	vmName := parts[2]
	return fmt.Sprintf("%s.%s.c.%s.internal", vmName, zone, projectID)
}
