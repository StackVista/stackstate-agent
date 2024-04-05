package urn

import (
	"context"
	"testing"

	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/clustername"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAwsUrnBuilder_BuildNodeInstanceIdentifier(t *testing.T) {
	mockConfig := config.Mock()
	var testClusterName = "mycluster"
	mockConfig.Set("cluster_name", testClusterName)

	clustername.GetClusterName(context.TODO(), "")

	builder := NewURNBuilder(Kubernetes, "uurrll")

	awsProviderID := "aws:///us-east-1b/i-024b28584ed2e6321"
	awsNode := coreV1.Node{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "ip-10-0-01-01.eu-west-1.compute.internal",
		}, Spec: coreV1.NodeSpec{ProviderID: awsProviderID},
	}
	awsNode.Status.Addresses = []coreV1.NodeAddress{
		{Type: coreV1.NodeInternalIP, Address: "10.20.01.01"},
		{Type: coreV1.NodeExternalIP, Address: "10.20.01.02"},
		{Type: coreV1.NodeInternalDNS, Address: "cluster.internal.amazon.net"},
		{Type: coreV1.NodeExternalDNS, Address: "amazon.com"},
	}
	awsIdentifiers := builder.BuildNodeURNs(awsNode)
	assert.Equal(t, []string{
		"urn:ip:/uurrll:ip-10-0-01-01.eu-west-1.compute.internal:10.20.01.01",
		"urn:ip:/uurrll:10.20.01.02",
		"urn:host:/uurrll:cluster.internal.amazon.net",
		"urn:host:/amazon.com",
		"urn:host:/ip-10-0-01-01.eu-west-1.compute.internal-mycluster",
	}, awsIdentifiers)
}

func TestAzureUrnBuilder_BuildNodeInstanceIdentifier(t *testing.T) {
	mockConfig := config.Mock()
	var testClusterName = "mycluster"
	mockConfig.Set("cluster_name", testClusterName)

	clustername.GetClusterName(context.TODO(), "")

	builder := NewURNBuilder(Kubernetes, "uurrll")

	azureProviderID := "azure:///subscriptions/d7e2ab8d-5edd-4db4-bc04-b1a193778fa3/resourceGroups/mc_test-stackstate_dev-cluster_westeurope/providers/Microsoft.Compute/virtualMachineScaleSets/aks-nodepool1-11692903-vmss/virtualMachines/0"
	azureNode := coreV1.Node{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "aks-nodepool1-11692903-vmss000000",
		},
		Spec: coreV1.NodeSpec{ProviderID: azureProviderID},
	}
	azureNode.Status.Addresses = []coreV1.NodeAddress{
		{Type: coreV1.NodeInternalIP, Address: "10.20.01.01"},
		{Type: coreV1.NodeExternalIP, Address: "10.20.01.02"},
		{Type: coreV1.NodeInternalDNS, Address: "cluster.internal.dns.azure.net"},
		{Type: coreV1.NodeExternalDNS, Address: "host.azure.com"},
	}
	azureIdentifiers := builder.BuildNodeURNs(azureNode)
	assert.Equal(t, []string{
		"urn:ip:/uurrll:aks-nodepool1-11692903-vmss000000:10.20.01.01",
		"urn:ip:/uurrll:10.20.01.02",
		"urn:host:/uurrll:cluster.internal.dns.azure.net",
		"urn:host:/host.azure.com",
		"urn:azure:/subscriptions/d7e2ab8d-5edd-4db4-bc04-b1a193778fa3/resourceGroups/mc_test-stackstate_dev-cluster_westeurope/providers/Microsoft.Compute/virtualMachineScaleSets/aks-nodepool1-11692903-vmss/virtualMachines/0",
		"urn:azure:/SUBSCRIPTIONS/D7E2AB8D-5EDD-4DB4-BC04-B1A193778FA3/RESOURCEGROUPS/MC_TEST-STACKSTATE_DEV-CLUSTER_WESTEUROPE/PROVIDERS/MICROSOFT.COMPUTE/VIRTUALMACHINESCALESETS/AKS-NODEPOOL1-11692903-VMSS/VIRTUALMACHINES/0",
		"urn:host:/aks-nodepool1-11692903-vmss000000-mycluster",
	}, azureIdentifiers)
}

func TestGceUrnBuilder_BuildNodeInstanceIdentifier(t *testing.T) {
	mockConfig := config.Mock()
	var testClusterName = "mycluster"
	mockConfig.Set("cluster_name", testClusterName)

	clustername.GetClusterName(context.TODO(), "")

	builder := NewURNBuilder(Kubernetes, "uurrll")

	gceProviderID := "gce://test-stackstate/europe-west4-a/gke-test-default-pool-9f8f65a4-2kld"
	gceNode := coreV1.Node{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "gke-test-default-pool-bbd2dc11-9wxt",
		}, Spec: coreV1.NodeSpec{ProviderID: gceProviderID}}
	gceNode.Status.Addresses = []coreV1.NodeAddress{
		{Type: coreV1.NodeInternalIP, Address: "10.20.01.01"},
		{Type: coreV1.NodeExternalIP, Address: "10.20.01.02"},
		{Type: coreV1.NodeInternalDNS, Address: "cluster.internal.dns.gce.net"},
		{Type: coreV1.NodeExternalDNS, Address: "host.gce.com"},
	}
	gceIdentifiers := builder.BuildNodeURNs(gceNode)
	assert.Equal(t, []string{
		"urn:ip:/uurrll:gke-test-default-pool-bbd2dc11-9wxt:10.20.01.01",
		"urn:ip:/uurrll:10.20.01.02",
		"urn:host:/uurrll:cluster.internal.dns.gce.net",
		"urn:host:/host.gce.com",
		"urn:host:/gke-test-default-pool-bbd2dc11-9wxt-mycluster",
	}, gceIdentifiers)
}
