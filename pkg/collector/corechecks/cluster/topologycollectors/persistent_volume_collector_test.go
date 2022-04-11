// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"testing"
	"time"

	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestPersistentVolumeCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)

	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)
	pathType = coreV1.HostPathFileOrCreate
	gcePersistentDisk = coreV1.GCEPersistentDiskVolumeSource{
		PDName: "name-of-the-gce-persistent-disk",
	}
	awsElasticBlockStore = coreV1.AWSElasticBlockStoreVolumeSource{
		VolumeID: "id-of-the-aws-block-store",
	}
	hostPath = coreV1.HostPathVolumeSource{
		Path: "some/path/to/the/volume",
		Type: &pathType,
	}
	csiPersistentVolume := coreV1.CSIPersistentVolumeSource{
		Driver:       "csi.trident.netapp.io",
		VolumeHandle: "pvc-03dr24ca-1sf4-acaw-1252-b8b232211244",
		ReadOnly:     false,
		VolumeAttributes: map[string]string{
			"backendUUID":  "127ebcb8-15gs-4fq1-acbn-021245ghgd05",
			"internalName": "NPO_TEST_pvc_0c8f1r14_a12a_1234_x1v2_b8b12341c1ab",
			"name":         "pvc-03dr24ca-1sf4-acaw-1252-b8b232211244",
			"protocol":     "file",
			"storage.kubernetes.io/csiProvisionerIdentity": "1245742285214-1234-csi.trident.netapp.io",
		},
	}

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, tc := range []struct {
			testCase                  string
			apiCollectorClientFactory func() apiserver.APICollectorClient
			assertions                []func(t *testing.T)
		}{
			{
				testCase: "Test Persistent Volume 1 - AWS Elastic Block Store",
				apiCollectorClientFactory: func() apiserver.APICollectorClient {
					return &MockPersistentVolumeAPICollectorClient{getPersistentVolumes: func() ([]coreV1.PersistentVolume, error) {
						persistentVolume := NewTestPV("aws-elastic-block-store-volume")
						persistentVolume.Spec.PersistentVolumeSource = coreV1.PersistentVolumeSource{
							AWSElasticBlockStore: &awsElasticBlockStore,
						}
						return []coreV1.PersistentVolume{persistentVolume}, nil
					}}
				},
				assertions: []func(*testing.T){
					func(t *testing.T) {
						component := <-componentChannel
						expected :=
							chooseBySourcePropertiesFeature(
								sourcePropertiesEnabled,
								&topology.Component{
									ExternalID: "urn:kubernetes:/test-cluster-name:persistent-volume/aws-elastic-block-store-volume",
									Type:       topology.Type{Name: "persistent-volume"},
									Data: topology.Data{
										"name":              "aws-elastic-block-store-volume",
										"creationTimestamp": creationTime,
										"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
										"uid":               types.UID("aws-elastic-block-store-volume"),
										"identifiers":       []string{},
										"status":            coreV1.VolumeAvailable,
										"statusMessage":     "Volume is available for use",
										"storageClassName":  "Storage-Class-Name",
									},
								},
								&topology.Component{
									ExternalID: "urn:kubernetes:/test-cluster-name:persistent-volume/aws-elastic-block-store-volume",
									Type:       topology.Type{Name: "persistent-volume"},
									Data: topology.Data{
										"name":        "aws-elastic-block-store-volume",
										"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
										"identifiers": []string{},
									},
									SourceProperties: map[string]interface{}{
										"metadata": map[string]interface{}{
											"creationTimestamp": creationTimeFormatted,
											"labels":            map[string]interface{}{"test": "label"},
											"name":              "aws-elastic-block-store-volume",
											"namespace":         "test-namespace",
											"uid":               "aws-elastic-block-store-volume",
										},
										"spec": map[string]interface{}{
											"persistentVolumeSource": map[string]interface{}{
												"awsElasticBlockStore": map[string]interface{}{
													"volumeID": "id-of-the-aws-block-store"}},
											"storageClassName": "Storage-Class-Name"},
										"status": map[string]interface{}{
											"phase":   "Available",
											"message": "Volume is available for use",
										},
									},
								},
							)
						assert.EqualValues(t, expected, component)
					},
					func(t *testing.T) {
						component := <-componentChannel
						expected := &topology.Component{
							ExternalID: "urn:kubernetes:external-volume:aws-ebs/id-of-the-aws-block-store/0",
							Type:       topology.Type{Name: "volume-source"},
							Data: topology.Data{
								"name": "id-of-the-aws-block-store",
								"tags": map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace", "partition": "0", "volume-id": "id-of-the-aws-block-store", "kind": "aws-ebs"},
								"source": coreV1.PersistentVolumeSource{
									AWSElasticBlockStore: &awsElasticBlockStore,
								},
							}}
						assert.EqualValues(t, expected, component)
					},
					func(t *testing.T) {
						relation := <-relationChannel
						expectedRelation := &topology.Relation{
							ExternalID: "urn:kubernetes:/test-cluster-name:persistent-volume/aws-elastic-block-store-volume->" +
								"urn:kubernetes:external-volume:aws-ebs/id-of-the-aws-block-store/0",
							Type:     topology.Type{Name: "exposes"},
							SourceID: "urn:kubernetes:/test-cluster-name:persistent-volume/aws-elastic-block-store-volume",
							TargetID: "urn:kubernetes:external-volume:aws-ebs/id-of-the-aws-block-store/0",
							Data:     map[string]interface{}{},
						}
						assert.EqualValues(t, expectedRelation, relation)
					},
				},
			},
			{
				testCase: "Test Persistent Volume 2 - GCE Persistent Disk",
				apiCollectorClientFactory: func() apiserver.APICollectorClient {
					return &MockPersistentVolumeAPICollectorClient{getPersistentVolumes: func() ([]coreV1.PersistentVolume, error) {
						persistentVolume := NewTestPV("gce-persistent-disk-volume")
						persistentVolume.Spec.PersistentVolumeSource = coreV1.PersistentVolumeSource{
							GCEPersistentDisk: &gcePersistentDisk,
						}
						return []coreV1.PersistentVolume{persistentVolume}, nil
					}}
				},
				assertions: []func(*testing.T){
					func(t *testing.T) {
						component := <-componentChannel
						expected := chooseBySourcePropertiesFeature(
							sourcePropertiesEnabled,
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:persistent-volume/gce-persistent-disk-volume",
								Type:       topology.Type{Name: "persistent-volume"},
								Data: topology.Data{
									"name":              "gce-persistent-disk-volume",
									"creationTimestamp": creationTime,
									"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"uid":               types.UID("gce-persistent-disk-volume"),
									"identifiers":       []string{},
									"status":            coreV1.VolumeAvailable,
									"statusMessage":     "Volume is available for use",
									"storageClassName":  "Storage-Class-Name",
								}},
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:persistent-volume/gce-persistent-disk-volume",
								Type:       topology.Type{Name: "persistent-volume"},
								Data: topology.Data{
									"name":        "gce-persistent-disk-volume",
									"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"identifiers": []string{},
								},
								SourceProperties: map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": creationTimeFormatted,
										"labels":            map[string]interface{}{"test": "label"},
										"name":              "gce-persistent-disk-volume",
										"namespace":         "test-namespace",
										"uid":               "gce-persistent-disk-volume",
									},
									"spec": map[string]interface{}{
										"persistentVolumeSource": map[string]interface{}{
											"gcePersistentDisk": map[string]interface{}{
												"pdName": "name-of-the-gce-persistent-disk"}},
										"storageClassName": "Storage-Class-Name"},
									"status": map[string]interface{}{
										"phase":   "Available",
										"message": "Volume is available for use",
									},
								}},
						)
						assert.EqualValues(t, expected, component)
					},
					func(t *testing.T) {
						component := <-componentChannel
						expected := &topology.Component{
							ExternalID: "urn:kubernetes:external-volume:gce-pd/name-of-the-gce-persistent-disk",
							Type:       topology.Type{Name: "volume-source"},
							Data: topology.Data{
								"name": "name-of-the-gce-persistent-disk",
								"tags": map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace", "kind": "gce-pd", "pd-name": "name-of-the-gce-persistent-disk"},
								"source": coreV1.PersistentVolumeSource{
									GCEPersistentDisk: &gcePersistentDisk,
								},
							}}
						assert.EqualValues(t, expected, component)
					},
					func(t *testing.T) {
						relation := <-relationChannel
						expectedRelation := &topology.Relation{
							ExternalID: "urn:kubernetes:/test-cluster-name:persistent-volume/gce-persistent-disk-volume->" +
								"urn:kubernetes:external-volume:gce-pd/name-of-the-gce-persistent-disk",
							Type:     topology.Type{Name: "exposes"},
							SourceID: "urn:kubernetes:/test-cluster-name:persistent-volume/gce-persistent-disk-volume",
							TargetID: "urn:kubernetes:external-volume:gce-pd/name-of-the-gce-persistent-disk",
							Data:     map[string]interface{}{},
						}
						assert.EqualValues(t, expectedRelation, relation)
					},
				},
			},
			{
				testCase: "Test Persistent Volume 3 - Host Path + Kind + Generate Name",
				apiCollectorClientFactory: func() apiserver.APICollectorClient {
					return &MockPersistentVolumeAPICollectorClient{getPersistentVolumes: func() ([]coreV1.PersistentVolume, error) {
						persistentVolume := NewTestPV("host-path-volume")
						persistentVolume.Spec.PersistentVolumeSource = coreV1.PersistentVolumeSource{
							HostPath: &hostPath,
						}
						persistentVolume.TypeMeta.Kind = "some-specified-kind"
						persistentVolume.ObjectMeta.GenerateName = "some-specified-generation"
						return []coreV1.PersistentVolume{persistentVolume}, nil
					}}
				},
				assertions: []func(*testing.T){
					func(t *testing.T) {
						component := <-componentChannel
						expected := chooseBySourcePropertiesFeature(
							sourcePropertiesEnabled,
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:persistent-volume/host-path-volume",
								Type:       topology.Type{Name: "persistent-volume"},
								Data: topology.Data{
									"name":              "host-path-volume",
									"creationTimestamp": creationTime,
									"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"uid":               types.UID("host-path-volume"),
									"identifiers":       []string{},
									"kind":              "some-specified-kind",
									"generateName":      "some-specified-generation",
									"status":            coreV1.VolumeAvailable,
									"statusMessage":     "Volume is available for use",
									"storageClassName":  "Storage-Class-Name",
								},
							},
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:persistent-volume/host-path-volume",
								Type:       topology.Type{Name: "persistent-volume"},
								Data: topology.Data{
									"name":        "host-path-volume",
									"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"identifiers": []string{},
								},
								SourceProperties: map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": creationTimeFormatted,
										"labels":            map[string]interface{}{"test": "label"},
										"name":              "host-path-volume",
										"namespace":         "test-namespace",
										"uid":               "host-path-volume",
										"generateName":      "some-specified-generation",
									},
									"spec": map[string]interface{}{
										"persistentVolumeSource": map[string]interface{}{
											"hostPath": map[string]interface{}{
												"path": "some/path/to/the/volume",
												"type": "FileOrCreate"}},
										"storageClassName": "Storage-Class-Name"},
									"status": map[string]interface{}{
										"phase":   "Available",
										"message": "Volume is available for use",
									},
								},
							},
						)
						assert.EqualValues(t, expected, component)
					},
				},
			},
			{
				testCase: "Test Persistent Volume 4 - Trident CSI Storage",
				apiCollectorClientFactory: func() apiserver.APICollectorClient {
					return &MockPersistentVolumeAPICollectorClient{getPersistentVolumes: func() ([]coreV1.PersistentVolume, error) {
						persistentVolume := NewTestPV("trident-csi-storage-volume")
						persistentVolume.Spec.PersistentVolumeSource = coreV1.PersistentVolumeSource{
							CSI: &csiPersistentVolume,
						}
						return []coreV1.PersistentVolume{persistentVolume}, nil
					}}
				},
				assertions: []func(*testing.T){
					func(t *testing.T) {
						component := <-componentChannel
						expected :=
							chooseBySourcePropertiesFeature(
								sourcePropertiesEnabled,
								&topology.Component{
									ExternalID: "urn:kubernetes:/test-cluster-name:persistent-volume/trident-csi-storage-volume",
									Type:       topology.Type{Name: "persistent-volume"},
									Data: topology.Data{
										"name":              "trident-csi-storage-volume",
										"creationTimestamp": creationTime,
										"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
										"uid":               types.UID("trident-csi-storage-volume"),
										"identifiers":       []string{},
										"status":            coreV1.VolumeAvailable,
										"statusMessage":     "Volume is available for use",
										"storageClassName":  "Storage-Class-Name",
									},
								},
								&topology.Component{
									ExternalID: "urn:kubernetes:/test-cluster-name:persistent-volume/trident-csi-storage-volume",
									Type:       topology.Type{Name: "persistent-volume"},
									Data: topology.Data{
										"name":        "trident-csi-storage-volume",
										"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
										"identifiers": []string{},
									},
									SourceProperties: map[string]interface{}{
										"metadata": map[string]interface{}{
											"creationTimestamp": creationTimeFormatted,
											"labels":            map[string]interface{}{"test": "label"},
											"name":              "trident-csi-storage-volume",
											"namespace":         "test-namespace",
											"uid":               "trident-csi-storage-volume",
										},
										"spec": map[string]interface{}{
											"persistentVolumeSource": map[string]interface{}{
												"csi": map[string]interface{}{
													"driver":       "csi.trident.netapp.io",
													"volumeHandle": "pvc-03dr24ca-1sf4-acaw-1252-b8b232211244",
													"volumeAttributes": map[string]interface{}{
														"driver":       "csi.trident.netapp.io",
														"kind":         "csi",
														"backendUUID":  "127ebcb8-15gs-4fq1-acbn-021245ghgd05",
														"internalName": "NPO_TEST_pvc_0c8f1r14_a12a_1234_x1v2_b8b12341c1ab",
														"name":         "pvc-03dr24ca-1sf4-acaw-1252-b8b232211244",
														"protocol":     "file",
														"storage.kubernetes.io/csiProvisionerIdentity": "1245742285214-1234-csi.trident.netapp.io",
													},
												},
											},
											"storageClassName": "Storage-Class-Name"},
										"status": map[string]interface{}{
											"phase":   "Available",
											"message": "Volume is available for use",
										},
									},
								},
							)
						assert.EqualValues(t, expected, component)
					},
					func(t *testing.T) {
						component := <-componentChannel
						expected := &topology.Component{
							ExternalID: "urn:kubernetes:external-volume:csi/csi.trident.netapp.io/pvc-03dr24ca-1sf4-acaw-1252-b8b232211244",
							Type:       topology.Type{Name: "volume-source"},
							Data: topology.Data{
								"name": "pvc-03dr24ca-1sf4-acaw-1252-b8b232211244",
								"tags": map[string]string{
									"test":         "label",
									"cluster-name": "test-cluster-name",
									"namespace":    "test-namespace",
									"kind":         "csi",
									"driver":       "csi.trident.netapp.io",
									"backendUUID":  "127ebcb8-15gs-4fq1-acbn-021245ghgd05",
									"internalName": "NPO_TEST_pvc_0c8f1r14_a12a_1234_x1v2_b8b12341c1ab",
									"name":         "pvc-03dr24ca-1sf4-acaw-1252-b8b232211244",
									"protocol":     "file",
									"storage.kubernetes.io/csiProvisionerIdentity": "1245742285214-1234-csi.trident.netapp.io",
								},
								"source": coreV1.PersistentVolumeSource{
									CSI: &csiPersistentVolume,
								},
							}}
						assert.EqualValues(t, expected, component)
					},
					func(t *testing.T) {
						relation := <-relationChannel
						expectedRelation := &topology.Relation{
							ExternalID: "urn:kubernetes:/test-cluster-name:persistent-volume/trident-csi-storage-volume->" +
								"urn:kubernetes:external-volume:csi/csi.trident.netapp.io/pvc-03dr24ca-1sf4-acaw-1252-b8b232211244",
							Type:     topology.Type{Name: "exposes"},
							SourceID: "urn:kubernetes:/test-cluster-name:persistent-volume/trident-csi-storage-volume",
							TargetID: "urn:kubernetes:external-volume:csi/csi.trident.netapp.io/pvc-03dr24ca-1sf4-acaw-1252-b8b232211244",
							Data:     map[string]interface{}{},
						}
						assert.EqualValues(t, expectedRelation, relation)
					},
				},
			},
		} {
			t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled), func(t *testing.T) {
				cmc := NewPersistentVolumeCollector(componentChannel, relationChannel, NewTestCommonClusterCollector(tc.apiCollectorClientFactory(), sourcePropertiesEnabled))
				expectedCollectorName := "Persistent Volume Collector"
				RunCollectorTest(t, cmc, expectedCollectorName)

				for _, a := range tc.assertions {
					a(t)
				}
			})
		}
	}
}

func NewTestPV(volumeName string) coreV1.PersistentVolume {
	return coreV1.PersistentVolume{
		TypeMeta: v1.TypeMeta{
			Kind: "",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              volumeName,
			CreationTimestamp: creationTime,
			Namespace:         "test-namespace",
			Labels: map[string]string{
				"test": "label",
			},
			UID:             types.UID(volumeName),
			GenerateName:    "",
			ResourceVersion: "123",
			ManagedFields: []v1.ManagedFieldsEntry{
				{
					Manager:    "ignored",
					Operation:  "Updated",
					APIVersion: "whatever",
					Time:       &v1.Time{Time: time.Now()},
					FieldsType: "whatever",
				},
			},
		},
		Spec: coreV1.PersistentVolumeSpec{
			StorageClassName: "Storage-Class-Name",
		},
		Status: coreV1.PersistentVolumeStatus{
			Phase:   coreV1.VolumeAvailable,
			Message: "Volume is available for use",
		},
	}
}

type MockPersistentVolumeAPICollectorClient struct {
	apiserver.APICollectorClient
	getPersistentVolumes func() ([]coreV1.PersistentVolume, error)
}

func (m MockPersistentVolumeAPICollectorClient) GetPersistentVolumes() ([]coreV1.PersistentVolume, error) {
	return m.getPersistentVolumes()
}
