package azure

import (
	"github.com/StackVista/stackstate-agent/pkg/util"
)

// GetHostname returns the Azure cloud specific hostname from the k8s providerID. This must be compatible with
// github.com/StackVista/stackstate-agent/pkg/util/cloudproviders/azure
// Example providerID:
func GetHostname(providerID string) string {
	return util.ExtractLastFragment(providerID)
}
