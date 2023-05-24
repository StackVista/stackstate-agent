package azure

import (
	"fmt"
	"regexp"

	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
)

const hostnameStyleSetting = "azure_hostname_style"
const azureProviderIDPattern = `^azure:///subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft\.Compute/virtualMachineScaleSets/([^/]+)/virtualMachines/(\d+)$`

// GetHostname returns the Azure cloud specific hostname from the k8s providerID. This must be compatible with
// github.com/StackVista/stackstate-agent/pkg/util/cloudproviders/azure
// Example providerID:
// Configuration option `azure_hostname_style`:
//
//	"os" - use the hostname reported by the operating system (default)
//	"name" - use the instance name
//	"name_and_resource_group" - use a combination of the instance name and resource group name
//	"full" - use a combination of the instance name, resource group name and subscription id
//	"vmid" - use the instance id
func GetHostname(node v1.Node) string {
	return GetHostnameWithConfig(node, config.Datadog)
}

// GetHostnameWithConfig returns the Azure cloud specific hostname from the k8s providerID. This must be compatible with
// github.com/StackVista/stackstate-agent/pkg/util/cloudproviders/azure
// Example providerID:
// Configuration option `azure_hostname_style`:
//
//	"os" - use the hostname reported by the operating system (default)
//	"name" - use the instance name
//	"name_and_resource_group" - use a combination of the instance name and resource group name
//	"full" - use a combination of the instance name, resource group name and subscription id
//	"vmid" - use the instance id
func GetHostnameWithConfig(node v1.Node, config config.Config) string {
	style := config.GetString(hostnameStyleSetting)
	var providerIDMatch = regexp.MustCompile(azureProviderIDPattern)
	switch style {
	case "os":
		return node.ObjectMeta.Name
	case "vmid":
		return node.Status.NodeInfo.SystemUUID
	default:
		if rs := providerIDMatch.FindStringSubmatch(node.Spec.ProviderID); rs != nil {
			subscription := rs[1]
			resourceGroup := rs[2]
			scaleSet := rs[3]
			instance := rs[4]
			switch style {
			case "name":
				return fmt.Sprintf("%s_%s", scaleSet, instance)
			case "name_and_resource_group":
				log.Warnf("Using the `name_and_resource_group` hostname style can result in mismatches in reported hostnames for the same VM because AKS lower cases the resource group name in the Kubernetes providerID. It is recommended to use the default `os` style instead.")
				return fmt.Sprintf("%s_%s.%s", scaleSet, instance, resourceGroup)
			case "full":
				log.Warnf("Using the `full` hostname style can result in mismatches in reported hostnames for the same VM because AKS lower cases the resource group name in the Kubernetes providerID. It is recommended to use the default `os` style instead.")
				return fmt.Sprintf("%s_%s.%s.%s", scaleSet, instance, resourceGroup, subscription)
			default:
				return ""
			}
		}
	}
	return ""
}
