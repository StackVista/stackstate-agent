// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var lastAppliedConfigurationJob = `{"apiVersion":"batch/v1","kind":"Job","metadata":{"annotations":{"argocd.io/hook":"Sync"},"generateName":"job-job","labels":{"app.kubernetes.io/component":"job-job","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/version":"1.0.0","helm.sh/chart":"1.0.0"},"name":"job-job-111111","namespace":"tenant"},"spec":{"backoffLimit":20,"template":{"metadata":{"annotations":{"checksum/str":"111qqq111"},"labels":{"app.kubernetes.io/component":"topic-create","app.kubernetes.io/instance":"test","app.kubernetes.io/name":"test"}},"spec":{"containers":[{"command":["bash","-c"],"env":[{"name":"VAR","value":"STR"}]}]}},"ttlSecondsAfterFinished":86400}}`
var creationTimeJob = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
var creationTimeFormattedJob = creationTimeJob.UTC().Format(time.RFC3339)
var parallelism int32
var backoffLimit int32

func TestJobCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	parallelism = int32(2)
	backoffLimit = int32(5)

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			commonClusterCollector := NewTestCommonClusterCollector(MockJobAPICollectorClient{}, componentChannel, relationChannel, sourcePropertiesEnabled, kubernetesStatusEnabled)
			commonClusterCollector.SetUseRelationCache(false)
			jc := NewJobCollector(commonClusterCollector)
			expectedCollectorName := "Job Collector"
			RunCollectorTest(t, jc, expectedCollectorName)

			for _, tc := range []struct {
				testCase                      string
				expectedComponentSP           *topology.Component
				expectedComponentNoSP         *topology.Component
				expectedComponentSPPlusStatus *topology.Component
				expectedRelations             []*topology.Relation
			}{
				{
					testCase: "Test Job 1 + Cron Job Relations",
					expectedComponentNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-1",
						Type:       topology.Type{Name: "job"},
						Data: topology.Data{
							"name":              "test-job-1",
							"kind":              "Job",
							"creationTimestamp": creationTimeJob,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-job",
								"namespace":      "test-namespace",
							},
							"uid":          types.UID("test-job-1"),
							"backoffLimit": &backoffLimit,
							"parallelism":  &parallelism,
						},
					},
					expectedComponentSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-1",
						Type:       topology.Type{Name: "job"},
						Data: topology.Data{
							"name": "test-job-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-job",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "batch/v1",
							"kind":       "Job",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormattedJob,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-job-1",
								"namespace":         "test-namespace",
								"ownerReferences": []interface{}{
									map[string]interface{}{"apiVersion": "", "kind": "CronJob", "name": "test-cronjob-1", "uid": ""},
								},
								"uid": "test-job-1",
							},
							"spec": map[string]interface{}{
								"backoffLimit": float64(5),
								"parallelism":  float64(2),
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": interface{}(nil),
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
						},
					},
					expectedComponentSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-1",
						Type:       topology.Type{Name: "job"},
						Data: topology.Data{
							"name": "test-job-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-job",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "batch/v1",
							"kind":       "Job",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormattedJob,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-job-1",
								"namespace":         "test-namespace",
								"ownerReferences": []interface{}{
									map[string]interface{}{"apiVersion": "", "kind": "CronJob", "name": "test-cronjob-1", "uid": ""},
								},
								"uid":             "test-job-1",
								"resourceVersion": "123",
								"annotations":     map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationJob},
							},
							"spec": map[string]interface{}{
								"backoffLimit": float64(5),
								"parallelism":  float64(2),
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": interface{}(nil),
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
							"status": map[string]interface{}{
								"completionTime": creationTimeFormattedJob,
								"startTime":      creationTimeFormattedJob,
								"succeeded":      float64(1),
								"conditions": []interface{}{map[string]interface{}{
									"lastProbeTime":      creationTimeFormattedJob,
									"lastTransitionTime": creationTimeFormattedJob,
									"status":             "True",
									"type":               "Complete",
								},
								},
							},
						},
					},
					expectedRelations: []*topology.Relation{
						{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-1->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-1",
							Type:     topology.Type{Name: "creates"},
							SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-1",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-1",
							Data:     map[string]interface{}{},
						},
					},
				},
				{
					testCase: "Test Job 2",
					expectedComponentNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-2",
						Type:       topology.Type{Name: "job"},
						Data: topology.Data{
							"name":              "test-job-2",
							"kind":              "Job",
							"creationTimestamp": creationTimeJob,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-job",
								"namespace":      "test-namespace",
							},
							"uid":          types.UID("test-job-2"),
							"backoffLimit": &backoffLimit,
							"parallelism":  &parallelism,
						},
					},
					expectedComponentSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-2",
						Type:       topology.Type{Name: "job"},
						Data: topology.Data{
							"name": "test-job-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-job",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "batch/v1",
							"kind":       "Job",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormattedJob,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-job-2",
								"namespace":         "test-namespace",
								"uid":               "test-job-2",
							},
							"spec": map[string]interface{}{
								"backoffLimit": float64(5),
								"parallelism":  float64(2),
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": interface{}(nil),
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
						},
					},
					expectedComponentSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-2",
						Type:       topology.Type{Name: "job"},
						Data: topology.Data{
							"name": "test-job-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-job",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "batch/v1",
							"kind":       "Job",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormattedJob,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-job-2",
								"namespace":         "test-namespace",
								"uid":               "test-job-2",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationJob},
							},
							"spec": map[string]interface{}{
								"backoffLimit": float64(5),
								"parallelism":  float64(2),
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": interface{}(nil),
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
							"status": map[string]interface{}{
								"completionTime": creationTimeFormattedJob,
								"startTime":      creationTimeFormattedJob,
								"succeeded":      float64(1),
								"conditions": []interface{}{map[string]interface{}{
									"lastProbeTime":      creationTimeFormattedJob,
									"lastTransitionTime": creationTimeFormattedJob,
									"status":             "True",
									"type":               "Complete",
								},
								},
							},
						},
					},
					expectedRelations: []*topology.Relation{
						{
							ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-2",
							Type:     topology.Type{Name: "encloses"},
							SourceID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-2",
							Data:     map[string]interface{}{},
						},
					},
				},
				{
					testCase: "Test Job 3 - Kind + Generate Name",
					expectedComponentNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-3",
						Type:       topology.Type{Name: "job"},
						Data: topology.Data{
							"name":              "test-job-3",
							"kind":              "Job",
							"creationTimestamp": creationTimeJob,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-job",
								"namespace":      "test-namespace",
							},
							"uid":          types.UID("test-job-3"),
							"generateName": "some-specified-generation",
							"backoffLimit": &backoffLimit,
							"parallelism":  &parallelism,
						},
					},
					expectedComponentSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-3",
						Type:       topology.Type{Name: "job"},
						Data: topology.Data{
							"name": "test-job-3",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-job",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "batch/v1",
							"kind":       "Job",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormattedJob,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-job-3",
								"namespace":         "test-namespace",
								"uid":               "test-job-3",
								"generateName":      "some-specified-generation",
							},
							"spec": map[string]interface{}{
								"backoffLimit": float64(5),
								"parallelism":  float64(2),
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": interface{}(nil),
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
						},
					},
					expectedComponentSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-3",
						Type:       topology.Type{Name: "job"},
						Data: topology.Data{
							"name": "test-job-3",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-job",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "batch/v1",
							"kind":       "Job",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormattedJob,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-job-3",
								"namespace":         "test-namespace",
								"uid":               "test-job-3",
								"generateName":      "some-specified-generation",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationJob},
							},
							"spec": map[string]interface{}{
								"backoffLimit": float64(5),
								"parallelism":  float64(2),
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": interface{}(nil),
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
							"status": map[string]interface{}{
								"completionTime": creationTimeFormattedJob,
								"startTime":      creationTimeFormattedJob,
								"succeeded":      float64(1),
								"conditions": []interface{}{map[string]interface{}{
									"lastProbeTime":      creationTimeFormattedJob,
									"lastTransitionTime": creationTimeFormattedJob,
									"status":             "True",
									"type":               "Complete",
								},
								},
							},
						},
					},
					expectedRelations: []*topology.Relation{
						{
							ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-3",
							Type:     topology.Type{Name: "encloses"},
							SourceID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-3",
							Data:     map[string]interface{}{},
						},
					},
				},
			} {
				t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled, kubernetesStatusEnabled), func(t *testing.T) {
					component := <-componentChannel
					if sourcePropertiesEnabled {
						if kubernetesStatusEnabled {
							assert.EqualValues(t, tc.expectedComponentSPPlusStatus, component)
						} else {
							assert.EqualValues(t, tc.expectedComponentSP, component)
						}
					} else {
						assert.EqualValues(t, tc.expectedComponentNoSP, component)
					}

					for _, expectedRelation := range tc.expectedRelations {
						cronJobRelation := <-relationChannel
						assert.EqualValues(t, expectedRelation, cronJobRelation)
					}

				})
			}
		}
	}
}

type MockJobAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockJobAPICollectorClient) GetJobs() ([]batchV1.Job, error) {
	jobs := make([]batchV1.Job, 0)
	for i := 1; i <= 3; i++ {
		job := batchV1.Job{
			TypeMeta: v1.TypeMeta{
				Kind: "Job",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-job-%d", i),
				CreationTimestamp: creationTimeJob,
				Namespace:         "test-namespace",
				Labels: map[string]string{
					"test": "label",
				},
				UID:             types.UID(fmt.Sprintf("test-job-%d", i)),
				GenerateName:    "",
				ResourceVersion: "123",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationJob,
				},
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
			Spec: batchV1.JobSpec{
				Parallelism:  &parallelism,
				BackoffLimit: &backoffLimit,
			},
			Status: batchV1.JobStatus{
				CompletionTime: &creationTimeJob,
				StartTime:      &creationTimeJob,
				Succeeded:      int32(1),
				Conditions: []batchV1.JobCondition{
					{
						LastProbeTime:      creationTimeJob,
						LastTransitionTime: creationTimeJob,
						Status:             "True",
						Type:               "Complete",
					},
				},
			},
		}

		if i == 1 {
			job.OwnerReferences = []v1.OwnerReference{
				{Kind: CronJob, Name: "test-cronjob-1"},
			}
		}

		if i == 3 {
			job.TypeMeta.Kind = "Job"
			job.ObjectMeta.GenerateName = "some-specified-generation"
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
