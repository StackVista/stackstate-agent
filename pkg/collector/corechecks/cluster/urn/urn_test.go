package urn

import (
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	"testing"
)

func TestUrnBuilder_BuildNodeInstanceIdentifier(t *testing.T) {

	builder := NewURNBuilder(Kubernetes, "uurrll")

	azureProviderID := "azure:///subscriptions/02ee9821-1b92-4cc8-84d2-bc87f369c88f/resourceGroups/aro-qjomvtki/providers/Microsoft.Compute/virtualMachines/stac14538openshift-5bljc-worker-westeurope3-6lq9w"
	azureNode := coreV1.Node{Spec: coreV1.NodeSpec{ProviderID: azureProviderID}}
	azureNode.Status.Addresses = []coreV1.NodeAddress{
		{Type: coreV1.NodeInternalIP, Address: "10.20.01.01"},
		{Type: coreV1.NodeExternalIP, Address: "10.20.01.02"},
		{Type: coreV1.NodeInternalDNS, Address: "cluster.internal.dns.azure.net"},
		{Type: coreV1.NodeExternalDNS, Address: "host.azure.com"},
	}
	azureIdentifiers := builder.BuildNodeURNs(azureNode)
	assert.Equal(t, []string{
		"urn:ip:/uurrll::10.20.01.01",
		"urn:ip:/uurrll:10.20.01.02",
		"urn:host:/uurrll:cluster.internal.dns.azure.net",
		"urn:host:/host.azure.com",
		"urn:azure:/subscriptions/02ee9821-1b92-4cc8-84d2-bc87f369c88f/resourceGroups/aro-qjomvtki/providers/Microsoft.Compute/virtualMachines/stac14538openshift-5bljc-worker-westeurope3-6lq9w",
		"urn:azure:/SUBSCRIPTIONS/02EE9821-1B92-4CC8-84D2-BC87F369C88F/RESOURCEGROUPS/ARO-QJOMVTKI/PROVIDERS/MICROSOFT.COMPUTE/VIRTUALMACHINES/STAC14538OPENSHIFT-5BLJC-WORKER-WESTEUROPE3-6LQ9W",
	}, azureIdentifiers)

	awsProviderID := "aws:///us-east-1b/i-024b28584ed2e6321"
	awsNode := coreV1.Node{Spec: coreV1.NodeSpec{ProviderID: awsProviderID}}
	awsNode.Status.Addresses = []coreV1.NodeAddress{
		{Type: coreV1.NodeInternalIP, Address: "10.20.01.01"},
		{Type: coreV1.NodeExternalIP, Address: "10.20.01.02"},
		{Type: coreV1.NodeInternalDNS, Address: "cluster.internal.amazon.net"},
		{Type: coreV1.NodeExternalDNS, Address: "amazon.com"},
	}
	awsIdentifiers := builder.BuildNodeURNs(awsNode)
	assert.Equal(t, []string{
		"urn:ip:/uurrll::10.20.01.01",
		"urn:ip:/uurrll:10.20.01.02",
		"urn:host:/uurrll:cluster.internal.amazon.net",
		"urn:host:/amazon.com",
	}, awsIdentifiers)
}
