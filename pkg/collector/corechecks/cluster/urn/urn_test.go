package urn

import (
	"testing"
)

func TestUrnBuilder_BuildNodeInstanceIdentifier(t *testing.T) {

	//builder := NewURNBuilder(Kubernetes, "uurrll")

	//azureProviderID := "azure:///subscriptions/02ee9821-1b92-4cc8-84d2-bc87f369c88f/resourceGroups/aro-qjomvtki/providers/Microsoft.Compute/virtualMachines/stac14538openshift-5bljc-worker-westeurope3-6lq9w"
	//azureNode := coreV1.Node{Spec: coreV1.NodeSpec{ProviderID: azureProviderID}}

	//builder.BuildNodeURNs(
	//
	//
	//azureInstanceID := builder.BuildNodeInstanceIdentifier(azureNode)
	//assert.Equal(t, "aro-qjomvtki:stac14538openshift-5bljc-worker-westeurope3-6lq9w", azureInstanceID))
	//
	//awsProviderID := "aws:///us-east-1b/i-024b28584ed2e6321"
	//awsNode := coreV1.Node{Spec: coreV1.NodeSpec{ProviderID: awsProviderID}}
	//awsInstanceID := builder.BuildNodeInstanceIdentifier(awsNode)
	//assert.Equal(t, "i-024b28584ed2e6321", awsInstanceID)
	//
	//azureProviderID := "azure:///subscriptions/02ee9821-1b92-4cc8-84d2-bc87f369c88f/resourceGroups/aro-qjomvtki/providers/Microsoft.Compute/virtualMachines/stac14538openshift-5bljc-worker-westeurope3-6lq9w"
	//azureNode := coreV1.Node{Spec: coreV1.NodeSpec{ProviderID: azureProviderID}}
	//azureInstanceID := builder.BuildNodeInstanceIdentifier(azureNode)
	//assert.Equal(t, "aro-qjomvtki:stac14538openshift-5bljc-worker-westeurope3-6lq9w", azureInstanceID)
}
