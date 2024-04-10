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

	batchV1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/version"

	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ComponentExpected struct {
	testCase             string
	expectedNoSP         *topology.Component
	expectedSP           *topology.Component
	expectedSPPlusStatus *topology.Component
}

var cronJobTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
var cronJobTimeFormatted = cronJobTime.UTC().Format(time.RFC3339)
var lastAppliedConfigurationCron = `{"apiVersion":"batch/v1","kind":"CronJob","metadata":{"labels":{"app.kubernetes.io/component":"backup"}},"spec":{"concurrencyPolicy":"Forbid","failedJobsHistoryLimit":1,"jobTemplate":{"metadata":{"labels":{"app.kubernetes.io/component":"backup"}},"spec":{"backoffLimit":1,"template":{"metadata":{"labels":{"app.kubernetes.io/component":"backup"}},"spec":{}}}},"schedule":"0 4 * * *","successfulJobsHistoryLimit":1}}`

func TestCronJobCollector_20(t *testing.T) {
	// Version 1.20 only supports batch/v1beta1
	k8sVersion := version.Info{
		Major: "1",
		Minor: "20",
	}
	expectedComponents := []ComponentExpected{
		{
			testCase:             "Test Cron Job 1 - Kind + Generate Name",
			expectedNoSP:         cronJobV1B1NoSP1(),
			expectedSP:           cronJobV1B1SP1(),
			expectedSPPlusStatus: cronJobV1B1SPPlusStatus1(),
		},
		{
			testCase:             "Test Cron Job 2 - Minimal",
			expectedNoSP:         cronJobV1B1NoSP2(),
			expectedSP:           cronJobV1B1SP2(),
			expectedSPPlusStatus: cronJobV1B1SPPlusStatus2(),
		},
	}
	testCronJobsWithK8SVersion(t, k8sVersion, expectedComponents)
}

func TestCronJobCollector_21(t *testing.T) {
	// Version 1.21 supports both batch/v1beta1 and batch/v1
	k8sVersion := version.Info{
		Major: "1",
		Minor: "21",
	}
	expectedComponents := []ComponentExpected{
		{
			testCase:             "Test Cron Job v1 - Minimal",
			expectedNoSP:         cronJobV1NoSP(),
			expectedSP:           cronJobV1SP(),
			expectedSPPlusStatus: cronJobV1SPPlusStatus(),
		},
	}
	testCronJobsWithK8SVersion(t, k8sVersion, expectedComponents)
}

func TestCronJobCollector_25(t *testing.T) {
	// Version 1.25 supports only batch/v1
	k8sVersion := version.Info{
		Major: "1",
		Minor: "25",
	}
	expectedComponents := []ComponentExpected{
		{
			testCase:             "Test Cron Job v1 1 - Kind + Generate Name",
			expectedNoSP:         cronJobV1NoSP(),
			expectedSP:           cronJobV1SP(),
			expectedSPPlusStatus: cronJobV1SPPlusStatus(),
		},
	}
	testCronJobsWithK8SVersion(t, k8sVersion, expectedComponents)
}

func testCronJobsWithK8SVersion(t *testing.T, k8sVersion version.Info, componentsExpected []ComponentExpected) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			commonClusterCollector := NewTestCommonClusterCollectorWithVersion(MockCronJobAPICollectorClient{}, sourcePropertiesEnabled, componentChannel, relationChannel, &k8sVersion, kubernetesStatusEnabled)
			commonClusterCollector.SetUseRelationCache(false)
			cjc := NewCronJobCollector(commonClusterCollector)
			expectedCollectorName := "CronJob Collector"
			RunCollectorTest(t, cjc, expectedCollectorName)

			for _, tc := range componentsExpected {
				t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled, kubernetesStatusEnabled), func(t *testing.T) {
					cronJob := <-componentChannel
					if sourcePropertiesEnabled {
						if kubernetesStatusEnabled {
							assert.EqualValues(t, tc.expectedSPPlusStatus, cronJob)
						} else {
							assert.EqualValues(t, tc.expectedSP, cronJob)
						}
					} else {
						assert.EqualValues(t, tc.expectedNoSP, cronJob)
					}

					actualRelation := <-relationChannel
					expectedRelation := &topology.Relation{
						ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" + cronJob.ExternalID,
						Type:       topology.Type{Name: "encloses"},
						SourceID:   "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
						TargetID:   cronJob.ExternalID,
						Data:       map[string]interface{}{},
					}
					assert.EqualValues(t, expectedRelation, actualRelation)
				})
			}
		}
	}
}

func cronJobV1B1SP2() *topology.Component {
	return &topology.Component{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-2",
		Type:       topology.Type{Name: "cronjob"},
		Data: topology.Data{
			"name": "test-cronjob-2",
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-cronjob",
				"namespace":      "test-namespace",
			},
		},
		SourceProperties: map[string]interface{}{
			"apiVersion": "batch/v1beta1",
			"kind":       "CronJob",
			"metadata": map[string]interface{}{
				"creationTimestamp": cronJobTimeFormatted,
				"labels":            map[string]interface{}{"test": "label"},
				"name":              "test-cronjob-2",
				"namespace":         "test-namespace",
				"uid":               "test-cronjob-2",
			},
			"spec": map[string]interface{}{
				"concurrencyPolicy": "Allow",
				"jobTemplate": map[string]interface{}{
					"metadata": map[string]interface{}{"creationTimestamp": interface{}(nil)},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"creationTimestamp": interface{}(nil),
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"image":     "busybox",
										"name":      "job",
										"resources": map[string]interface{}{},
									},
								},
								"restartPolicy": "OnFailure",
							},
						},
					},
				},
				"schedule": "0 0 * * *",
			},
		},
	}
}

func cronJobV1B1SPPlusStatus2() *topology.Component {
	return &topology.Component{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-2",
		Type:       topology.Type{Name: "cronjob"},
		Data: topology.Data{
			"name": "test-cronjob-2",
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-cronjob",
				"namespace":      "test-namespace",
			},
		},
		SourceProperties: map[string]interface{}{
			"apiVersion": "batch/v1beta1",
			"kind":       "CronJob",
			"metadata": map[string]interface{}{
				"creationTimestamp": cronJobTimeFormatted,
				"labels":            map[string]interface{}{"test": "label"},
				"name":              "test-cronjob-2",
				"namespace":         "test-namespace",
				"uid":               "test-cronjob-2",
				"resourceVersion":   "123",
				"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationCron},
			},
			"spec": map[string]interface{}{
				"concurrencyPolicy": "Allow",
				"jobTemplate": map[string]interface{}{
					"metadata": map[string]interface{}{"creationTimestamp": interface{}(nil)},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"creationTimestamp": interface{}(nil),
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"image":     "busybox",
										"name":      "job",
										"resources": map[string]interface{}{},
									},
								},
								"restartPolicy": "OnFailure",
							},
						},
					},
				},
				"schedule": "0 0 * * *",
			},
			"status": map[string]interface{}{
				"active":             []interface{}{map[string]interface{}{"kind": "Job", "name": "cronjob-job-1"}},
				"lastScheduleTime":   cronJobTimeFormatted,
				"lastSuccessfulTime": cronJobTimeFormatted,
			},
		},
	}
}

func cronJobV1B1NoSP2() *topology.Component {
	return &topology.Component{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-2",
		Type:       topology.Type{Name: "cronjob"},
		Data: topology.Data{
			"schedule":          "0 0 * * *",
			"name":              "test-cronjob-2",
			"kind":              "CronJob",
			"creationTimestamp": cronJobTime,
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-cronjob",
				"namespace":      "test-namespace",
			},
			"uid":               types.UID("test-cronjob-2"),
			"concurrencyPolicy": "Allow",
		},
	}
}

func cronJobV1B1SP1() *topology.Component {
	return &topology.Component{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-1",
		Type:       topology.Type{Name: "cronjob"},
		Data: topology.Data{
			"name": "test-cronjob-1",
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-cronjob",
				"namespace":      "test-namespace",
			},
		},
		SourceProperties: map[string]interface{}{
			"apiVersion": "batch/v1beta1",
			"kind":       "CronJob",
			"metadata": map[string]interface{}{
				"creationTimestamp": cronJobTimeFormatted,
				"labels":            map[string]interface{}{"test": "label"},
				"name":              "test-cronjob-1",
				"namespace":         "test-namespace",
				"uid":               "test-cronjob-1",
				"generateName":      "some-specified-generation",
			},
			"spec": map[string]interface{}{
				"concurrencyPolicy": "Allow",
				"jobTemplate": map[string]interface{}{
					"metadata": map[string]interface{}{"creationTimestamp": interface{}(nil)},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"creationTimestamp": interface{}(nil),
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"image":     "busybox",
										"name":      "job",
										"resources": map[string]interface{}{},
									},
								},
								"restartPolicy": "OnFailure",
							},
						},
					},
				},
				"schedule": "0 0 * * *",
			},
		},
	}
}

func cronJobV1B1SPPlusStatus1() *topology.Component {
	return &topology.Component{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-1",
		Type:       topology.Type{Name: "cronjob"},
		Data: topology.Data{
			"name": "test-cronjob-1",
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-cronjob",
				"namespace":      "test-namespace",
			},
		},
		SourceProperties: map[string]interface{}{
			"apiVersion": "batch/v1beta1",
			"kind":       "CronJob",
			"metadata": map[string]interface{}{
				"creationTimestamp": cronJobTimeFormatted,
				"labels":            map[string]interface{}{"test": "label"},
				"name":              "test-cronjob-1",
				"namespace":         "test-namespace",
				"uid":               "test-cronjob-1",
				"generateName":      "some-specified-generation",
				"resourceVersion":   "123",
				"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationCron},
			},
			"spec": map[string]interface{}{
				"concurrencyPolicy": "Allow",
				"jobTemplate": map[string]interface{}{
					"metadata": map[string]interface{}{"creationTimestamp": interface{}(nil)},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"creationTimestamp": interface{}(nil),
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"image":     "busybox",
										"name":      "job",
										"resources": map[string]interface{}{},
									},
								},
								"restartPolicy": "OnFailure",
							},
						},
					},
				},
				"schedule": "0 0 * * *",
			},
			"status": map[string]interface{}{
				"active":             []interface{}{map[string]interface{}{"kind": "Job", "name": "cronjob-job-1"}},
				"lastScheduleTime":   cronJobTimeFormatted,
				"lastSuccessfulTime": cronJobTimeFormatted,
			},
		},
	}
}

func cronJobV1B1NoSP1() *topology.Component {
	return &topology.Component{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-1",
		Type:       topology.Type{Name: "cronjob"},
		Data: topology.Data{
			"schedule":          "0 0 * * *",
			"name":              "test-cronjob-1",
			"creationTimestamp": cronJobTime,
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-cronjob",
				"namespace":      "test-namespace",
			},
			"uid":               types.UID("test-cronjob-1"),
			"concurrencyPolicy": "Allow",
			"kind":              "CronJob",
			"generateName":      "some-specified-generation",
		},
	}
}

func cronJobV1SP() *topology.Component {
	return &topology.Component{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-v1",
		Type:       topology.Type{Name: "cronjob"},
		Data: topology.Data{
			"name": "test-cronjob-v1",
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-cronjob",
				"namespace":      "test-namespace",
			},
		},
		SourceProperties: map[string]interface{}{
			"apiVersion": "batch/v1",
			"kind":       "CronJob",
			"metadata": map[string]interface{}{
				"creationTimestamp": cronJobTimeFormatted,
				"labels":            map[string]interface{}{"test": "label"},
				"name":              "test-cronjob-v1",
				"namespace":         "test-namespace",
				"uid":               "test-cronjob-v1",
				"generateName":      "some-specified-generation-v1",
			},
			"spec": map[string]interface{}{
				"concurrencyPolicy": "Allow",
				"jobTemplate": map[string]interface{}{
					"metadata": map[string]interface{}{"creationTimestamp": interface{}(nil)},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"creationTimestamp": interface{}(nil),
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"image":     "busybox",
										"name":      "job",
										"resources": map[string]interface{}{},
									},
								},
								"restartPolicy": "OnFailure",
							},
						},
					},
				},
				"schedule": "0 1 * * *",
			},
		},
	}
}

func cronJobV1SPPlusStatus() *topology.Component {
	return &topology.Component{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-v1",
		Type:       topology.Type{Name: "cronjob"},
		Data: topology.Data{
			"name": "test-cronjob-v1",
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-cronjob",
				"namespace":      "test-namespace",
			},
		},
		SourceProperties: map[string]interface{}{
			"apiVersion": "batch/v1",
			"kind":       "CronJob",
			"metadata": map[string]interface{}{
				"creationTimestamp": cronJobTimeFormatted,
				"labels":            map[string]interface{}{"test": "label"},
				"name":              "test-cronjob-v1",
				"namespace":         "test-namespace",
				"uid":               "test-cronjob-v1",
				"generateName":      "some-specified-generation-v1",
				"resourceVersion":   "123",
				"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationCron},
			},
			"spec": map[string]interface{}{
				"concurrencyPolicy": "Allow",
				"jobTemplate": map[string]interface{}{
					"metadata": map[string]interface{}{"creationTimestamp": interface{}(nil)},
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{
								"creationTimestamp": interface{}(nil),
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"image":     "busybox",
										"name":      "job",
										"resources": map[string]interface{}{},
									},
								},
								"restartPolicy": "OnFailure",
							},
						},
					},
				},
				"schedule": "0 1 * * *",
			},
			"status": map[string]interface{}{
				"active":             []interface{}{map[string]interface{}{"kind": "Job", "name": "cronjob-job-v1"}},
				"lastScheduleTime":   cronJobTimeFormatted,
				"lastSuccessfulTime": cronJobTimeFormatted,
			},
		},
	}
}

func cronJobV1NoSP() *topology.Component {
	return &topology.Component{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-v1",
		Type:       topology.Type{Name: "cronjob"},
		Data: topology.Data{
			"schedule":          "0 1 * * *",
			"name":              "test-cronjob-v1",
			"creationTimestamp": cronJobTime,
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-cronjob",
				"namespace":      "test-namespace",
			},
			"uid":               types.UID("test-cronjob-v1"),
			"concurrencyPolicy": "Allow",
			"kind":              "CronJob",
			"generateName":      "some-specified-generation-v1",
		},
	}
}

type MockCronJobAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockCronJobAPICollectorClient) GetCronJobsV1B1() ([]v1beta1.CronJob, error) {
	cronJobs := make([]v1beta1.CronJob, 0)
	for i := 1; i <= 2; i++ {
		job := v1beta1.CronJob{
			TypeMeta: v1.TypeMeta{
				Kind: "CronJob",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-cronjob-%d", i),
				CreationTimestamp: cronJobTime,
				Namespace:         "test-namespace",
				Labels: map[string]string{
					"test": "label",
				},
				UID:             types.UID(fmt.Sprintf("test-cronjob-%d", i)),
				GenerateName:    "",
				ResourceVersion: "123",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationCron,
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
			Spec: v1beta1.CronJobSpec{
				Schedule:          "0 0 * * *",
				ConcurrencyPolicy: v1beta1.AllowConcurrent,
				JobTemplate: v1beta1.JobTemplateSpec{
					Spec: batchV1.JobSpec{
						Template: apiv1.PodTemplateSpec{
							Spec: apiv1.PodSpec{
								Containers: []apiv1.Container{
									{
										Name:  "job",
										Image: "busybox",
									},
								},
								RestartPolicy: apiv1.RestartPolicyOnFailure,
							},
						},
					},
				},
			},
			Status: v1beta1.CronJobStatus{
				Active: []apiv1.ObjectReference{
					{Kind: "Job", Name: "cronjob-job-1"},
				},
				LastScheduleTime:   &cronJobTime,
				LastSuccessfulTime: &cronJobTime,
			},
		}

		if i == 1 {
			job.TypeMeta.Kind = "CronJob"
			job.ObjectMeta.GenerateName = "some-specified-generation"
		}

		cronJobs = append(cronJobs, job)
	}

	return cronJobs, nil
}

func (m MockCronJobAPICollectorClient) GetCronJobsV1() ([]batchV1.CronJob, error) {
	cronJobs := make([]batchV1.CronJob, 0)
	job := batchV1.CronJob{
		TypeMeta: v1.TypeMeta{
			Kind: "CronJob",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              "test-cronjob-v1",
			CreationTimestamp: cronJobTime,
			Namespace:         "test-namespace",
			Labels: map[string]string{
				"test": "label",
			},
			UID:             types.UID("test-cronjob-v1"),
			GenerateName:    "some-specified-generation-v1",
			ResourceVersion: "123",
			Annotations: map[string]string{
				"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationCron,
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
		Spec: batchV1.CronJobSpec{
			Schedule:          "0 1 * * *",
			ConcurrencyPolicy: batchV1.AllowConcurrent,
			JobTemplate: batchV1.JobTemplateSpec{
				Spec: batchV1.JobSpec{
					Template: apiv1.PodTemplateSpec{
						Spec: apiv1.PodSpec{
							Containers: []apiv1.Container{
								{
									Name:  "job",
									Image: "busybox",
								},
							},
							RestartPolicy: apiv1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
		Status: batchV1.CronJobStatus{
			Active: []apiv1.ObjectReference{
				{Kind: "Job", Name: "cronjob-job-v1"},
			},
			LastScheduleTime:   &cronJobTime,
			LastSuccessfulTime: &cronJobTime,
		},
	}

	cronJobs = append(cronJobs, job)

	return cronJobs, nil
}
