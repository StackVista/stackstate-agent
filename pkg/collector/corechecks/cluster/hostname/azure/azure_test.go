package azure

import (
	"testing"

	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/hostname/testutil"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/stretchr/testify/assert"
)

const testProviderID = "azure:///subscriptions/d7e2ab8d-5edd-4db4-bc04-b1a193778fa3/resourceGroups/mc_test-stackstate_dev-cluster_westeurope/providers/Microsoft.Compute/virtualMachineScaleSets/aks-nodepool1-11692903-vmss/virtualMachines/0"
const osHostname = "aks-nodepool1-11692903-vmss000000"
const vmID = "877e90e3-041e-4784-8a55-096f71d0ff6c"

var testNode = testutil.TestNode(testProviderID, osHostname, vmID)

func TestAzureHostnameUsesOsHost(t *testing.T) {
	config := config.Mock()
	config.Set(hostnameStyleSetting, "os")
	assert.Equal(t, osHostname, GetHostnameWithConfig(testNode, config))
}

func TestAzureHostnameUsesVmID(t *testing.T) {
	config := config.Mock()
	config.Set(hostnameStyleSetting, "vmid")
	assert.Equal(t, vmID, GetHostnameWithConfig(testNode, config))
}

func TestAzureHostnameUsesName(t *testing.T) {
	config := config.Mock()
	config.Set(hostnameStyleSetting, "name")
	assert.Equal(t, "aks-nodepool1-11692903-vmss_0", GetHostnameWithConfig(testNode, config))
}
func TestAzureHostnameUsesNameAndResourceGroup(t *testing.T) {
	config := config.Mock()
	config.Set(hostnameStyleSetting, "name_and_resource_group")
	// Note that the Azure metadata based hostname reported by the node and process agent would be aks-nodepool1-11692903-vmss_0.MC_test-stackstate_dev-cluster_westeurope instead
	// AKS lower cases resource group names by default (actually any cluster running on Azure does that).
	assert.Equal(t, "aks-nodepool1-11692903-vmss_0.mc_test-stackstate_dev-cluster_westeurope", GetHostnameWithConfig(testNode, config))
}
func TestAzureHostnameUsesVmFull(t *testing.T) {
	config := config.Mock()
	config.Set(hostnameStyleSetting, "full")
	// Note that the Azure metadata based hostname reported by the node and process agent would be aks-nodepool1-11692903-vmss_0.MC_test-stackstate_dev-cluster_westeurope.d7e2ab8d-5edd-4db4-bc04-b1a193778fa3 instead
	// AKS lower cases resource group names by default (actually any cluster running on Azure does that).
	assert.Equal(t, "aks-nodepool1-11692903-vmss_0.mc_test-stackstate_dev-cluster_westeurope.d7e2ab8d-5edd-4db4-bc04-b1a193778fa3", GetHostnameWithConfig(testNode, config))
}
func TestAzureHostnameEmpty(t *testing.T) {
	assert.Equal(t, "", GetHostname(testutil.TestNodeForProviderID("")))
}
