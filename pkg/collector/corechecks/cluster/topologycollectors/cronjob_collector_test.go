// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	batchV1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	"testing"
	"time"

	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestCronJobCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		cjc := NewCronJobCollector(componentChannel, relationChannel, NewTestCommonClusterCollector(MockCronJobAPICollectorClient{}, sourcePropertiesEnabled))
		expectedCollectorName := "CronJob Collector"
		RunCollectorTest(t, cjc, expectedCollectorName)

		for _, tc := range []struct {
			testCase     string
			expectedNoSP *topology.Component
			expectedSP   *topology.Component
		}{
			{
				testCase: "Test Cron Job 1 - Kind + Generate Name",
				expectedNoSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-1",
					Type:       topology.Type{Name: "cronjob"},
					Data: topology.Data{
						"schedule":          "0 0 * * *",
						"name":              "test-cronjob-1",
						"creationTimestamp": creationTime,
						"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
						"uid":               types.UID("test-cronjob-1"),
						"concurrencyPolicy": v1beta1.AllowConcurrent,
						"kind":              "some-specified-kind",
						"generateName":      "some-specified-generation",
					},
				},
				expectedSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-1",
					Type:       topology.Type{Name: "cronjob"},
					Data: topology.Data{
						"name": "test-cronjob-1",
						"tags": map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					},
					SourceProperties: map[string]interface{}{
						"metadata": map[string]interface{}{
							"creationTimestamp": creationTimeFormatted,
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
				},
			},
			{
				testCase: "Test Cron Job 2 - Minimal",
				expectedNoSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-2",
					Type:       topology.Type{Name: "cronjob"},
					Data: topology.Data{
						"schedule":          "0 0 * * *",
						"name":              "test-cronjob-2",
						"creationTimestamp": creationTime,
						"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
						"uid":               types.UID("test-cronjob-2"),
						"concurrencyPolicy": v1beta1.AllowConcurrent,
					},
				},
				expectedSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-2",
					Type:       topology.Type{Name: "cronjob"},
					Data: topology.Data{
						"name": "test-cronjob-2",
						"tags": map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					},
					SourceProperties: map[string]interface{}{
						"metadata": map[string]interface{}{
							"creationTimestamp": creationTimeFormatted,
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
				},
			},
		} {
			t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled), func(t *testing.T) {
				cronJob := <-componentChannel
				if sourcePropertiesEnabled {
					assert.EqualValues(t, tc.expectedSP, cronJob)
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

type MockCronJobAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockCronJobAPICollectorClient) GetCronJobs() ([]v1beta1.CronJob, error) {
	cronJobs := make([]v1beta1.CronJob, 0)
	for i := 1; i <= 2; i++ {
		job := v1beta1.CronJob{
			TypeMeta: v1.TypeMeta{
				Kind: "",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-cronjob-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				Labels: map[string]string{
					"test": "label",
				},
				UID:             types.UID(fmt.Sprintf("test-cronjob-%d", i)),
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
				LastScheduleTime: &creationTime,
			},
		}

		if i == 1 {
			job.TypeMeta.Kind = "some-specified-kind"
			job.ObjectMeta.GenerateName = "some-specified-generation"
		}

		cronJobs = append(cronJobs, job)
	}

	return cronJobs, nil
}
