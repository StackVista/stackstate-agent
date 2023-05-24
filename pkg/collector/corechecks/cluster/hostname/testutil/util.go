package testutil

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeForProviderID(providerID string) v1.Node {
	return TestNode(providerID, "test-node", "test-system-uuid")
}

func TestNode(providerID string, nodeName string, systemUUID string) v1.Node {
	return v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
		},
		Spec: v1.NodeSpec{
			ProviderID: providerID,
		},
		Status: v1.NodeStatus{
			NodeInfo: v1.NodeSystemInfo{
				SystemUUID: systemUUID,
			},
		},
	}
}
