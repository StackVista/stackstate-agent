package aws

import (
	"github.com/StackVista/stackstate-agent/pkg/util"
)

// GetHostname returns the AWS cloud specific hostname from the k8s providerId. This must be compatible with
// github.com/StackVista/stackstate-agent/pkg/util/cloudproviders/ec2
// Example providerId: aws:///eu-west-1b/i-09e7ef36c4efd7cbe
func GetHostname(providerID string) string {
	return util.ExtractLastFragment(providerID)
}
