//go:build kubeapiserver

package topologycollectors

import (
	"fmt"
	"testing"

	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPersistentVolumeCollectorSkipsInlineVolumeAttachments(t *testing.T) {
	componentChannel := make(chan *topology.Component)
	relationChannel := make(chan *topology.Relation)

	ebs := coreV1.AWSElasticBlockStoreVolumeSource{VolumeID: "id-of-the-aws-block-store"}
	inline := storageV1.VolumeAttachment{
		ObjectMeta: v1.ObjectMeta{Name: "csi-inline"},
		Spec: storageV1.VolumeAttachmentSpec{
			Attacher: "ebs.csi.aws.com",
			NodeName: "test-node-2",
			Source: storageV1.VolumeAttachmentSource{
				InlineVolumeSpec: &coreV1.PersistentVolumeSpec{
					PersistentVolumeSource: coreV1.PersistentVolumeSource{
						CSI: &coreV1.CSIPersistentVolumeSource{Driver: "ebs.csi.aws.com", VolumeHandle: "vol-inline"},
					},
				},
			},
		},
	}
	client := &MockPersistentVolumeAPICollectorClient{
		getPersistentVolumes: func() ([]coreV1.PersistentVolume, error) {
			pv := NewTestPV("aws-elastic-block-store-volume")
			pv.Spec.PersistentVolumeSource = coreV1.PersistentVolumeSource{AWSElasticBlockStore: &ebs}
			return []coreV1.PersistentVolume{pv}, nil
		},
		getPersistentVolumeClaims: func() ([]coreV1.PersistentVolumeClaim, error) {
			return []coreV1.PersistentVolumeClaim{}, nil
		},
		getVolumeAttachments: func() ([]storageV1.VolumeAttachment, error) {
			return []storageV1.VolumeAttachment{inline, NewTestVolumeAttachment("aws-elastic-block-store-volume")}, nil
		},
	}
	common := NewTestCommonClusterCollector(client, componentChannel, relationChannel)
	common.SetUseRelationCache(false)
	collector := NewPersistentVolumeCollector(common, true)

	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("collector panicked: %v", r)
			}
		}()
		done <- collector.CollectorFunction()
	}()

	var persistentVolume *topology.Component
	var nodeRelations []*topology.Relation
	for {
		select {
		case component := <-componentChannel:
			if component.Type.Name == "persistent-volume" {
				persistentVolume = component
			}
		case relation := <-relationChannel:
			if relation.SourceID == "urn:kubernetes:/test-cluster-name:node/test-node-1" ||
				relation.SourceID == "urn:kubernetes:/test-cluster-name:node/test-node-2" {
				nodeRelations = append(nodeRelations, relation)
			}
		case err := <-done:
			require.NoError(t, err)
			require.NotNil(t, persistentVolume)
			tags, ok := persistentVolume.Data["tags"].(map[string]string)
			require.True(t, ok)
			assert.Equal(t, "test-node-1", tags["persistent-volume-node"])
			require.Len(t, nodeRelations, 1, "only the attachment naming a PersistentVolume relates a node to it")
			assert.Equal(t, "urn:kubernetes:/test-cluster-name:persistent-volume/aws-elastic-block-store-volume", nodeRelations[0].TargetID)
			return
		}
	}
}
