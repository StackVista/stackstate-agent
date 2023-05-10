// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var lastAppliedConfigurationSecret = `{"apiVersion":"v1","data":{"EXTRA":"123"},"kind":"Secret","metadata":{"annotations":{"argocd.io/tracking-id":"api"},"labels":{"app.kubernetes.io/instance":"test","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"app","app.kubernetes.io/version":"1.0.0","helm.sh/chart":"1.0.0"},"name":"api","namespace":"tenant"},"stringData":{"app.conf":"text"},"type":"Opaque"}`

func TestSecretCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			cmc := NewSecretCollector(NewTestCommonClusterCollector(MockSecretAPICollectorClient{}, componentChannel, relationChannel, sourcePropertiesEnabled, kubernetesStatusEnabled))
			expectedCollectorName := "Secret Collector"
			RunCollectorTest(t, cmc, expectedCollectorName)

			for _, tc := range []struct {
				testCase             string
				expectedSP           *topology.Component
				expectedNoSP         *topology.Component
				expectedSPPlusStatus *topology.Component
			}{
				{
					testCase: "Test Secret 1 - Complete",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-1",
						Type:       topology.Type{Name: "secret"},
						Data: topology.Data{
							"name":              "test-secret-1",
							"kind":              "Secret",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-secret",
								"namespace":      "test-namespace",
							},
							"uid":         types.UID("test-secret-1"),
							"data":        "c20ca49dcb76feaaa1c14a2725263bf2290d0e5f3dc98d208b249f080fa64b45",
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-1"},
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-1",
						Type:       topology.Type{Name: "secret"},
						Data: topology.Data{
							"name": "test-secret-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-secret",
								"namespace":      "test-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-1"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"annotations":       map[string]interface{}{"openshift.io/token-secret.value": "<redacted>"},
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-secret-1",
								"namespace":         "test-namespace",
								"uid":               "test-secret-1"},
							"data": map[string]interface{}{
								"<data hash>": "YzIwY2E0OWRjYjc2ZmVhYWExYzE0YTI3MjUyNjNiZjIyOTBkMGU1ZjNkYzk4ZDIwOGIyNDlmMDgwZmE2NGI0NQ==",
							},
						},
					},
					expectedSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-1",
						Type:       topology.Type{Name: "secret"},
						Data: topology.Data{
							"name": "test-secret-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-secret",
								"namespace":      "test-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-1"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"annotations":       map[string]interface{}{"openshift.io/token-secret.value": "<redacted>"},
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-secret-1",
								"namespace":         "test-namespace",
								"uid":               "test-secret-1",
								"resourceVersion":   "123",
							},
							"data": map[string]interface{}{
								"<data hash>": "YzIwY2E0OWRjYjc2ZmVhYWExYzE0YTI3MjUyNjNiZjIyOTBkMGU1ZjNkYzk4ZDIwOGIyNDlmMDgwZmE2NGI0NQ==",
							},
						},
					},
				},
				{
					testCase: "Test Secret 2 - Without Data",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-2",
						Type:       topology.Type{Name: "secret"},
						Data: topology.Data{
							"name":              "test-secret-2",
							"kind":              "Secret",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-secret",
								"namespace":      "test-namespace",
							},
							"uid":         types.UID("test-secret-2"),
							"data":        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // Empty data is represented as a hash to obscure it
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-2"},
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-2",
						Type:       topology.Type{Name: "secret"},
						Data: topology.Data{
							"name": "test-secret-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-secret",
								"namespace":      "test-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-2"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"annotations":       map[string]interface{}{"openshift.io/token-secret.value": "<redacted>"},
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-secret-2",
								"namespace":         "test-namespace",
								"uid":               "test-secret-2"},
							"data": map[string]interface{}{
								"<data hash>": "ZTNiMGM0NDI5OGZjMWMxNDlhZmJmNGM4OTk2ZmI5MjQyN2FlNDFlNDY0OWI5MzRjYTQ5NTk5MWI3ODUyYjg1NQ==",
							},
						},
					},
					expectedSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-2",
						Type:       topology.Type{Name: "secret"},
						Data: topology.Data{
							"name": "test-secret-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-secret",
								"namespace":      "test-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-2"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"annotations":       map[string]interface{}{"openshift.io/token-secret.value": "<redacted>"},
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-secret-2",
								"namespace":         "test-namespace",
								"uid":               "test-secret-2",
								"resourceVersion":   "123",
							},
							"data": map[string]interface{}{
								"<data hash>": "ZTNiMGM0NDI5OGZjMWMxNDlhZmJmNGM4OTk2ZmI5MjQyN2FlNDFlNDY0OWI5MzRjYTQ5NTk5MWI3ODUyYjg1NQ==",
							},
						},
					},
				},
				{
					testCase: "Test Secret 3 - Minimal",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-3",
						Type:       topology.Type{Name: "secret"},
						Data: topology.Data{
							"name":              "test-secret-3",
							"kind":              "Secret",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-secret",
								"namespace":      "test-namespace",
							},
							"uid":         types.UID("test-secret-3"),
							"data":        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // Empty data is represented as a hash to obscure it
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-3"},
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-3",
						Type:       topology.Type{Name: "secret"},
						Data: topology.Data{
							"name": "test-secret-3",
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-secret",
								"namespace":      "test-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-3"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"annotations":       map[string]interface{}{"openshift.io/token-secret.value": "<redacted>"},
								"creationTimestamp": creationTimeFormatted,
								"name":              "test-secret-3",
								"namespace":         "test-namespace",
								"uid":               "test-secret-3"},
							"data": map[string]interface{}{
								"<data hash>": "ZTNiMGM0NDI5OGZjMWMxNDlhZmJmNGM4OTk2ZmI5MjQyN2FlNDFlNDY0OWI5MzRjYTQ5NTk5MWI3ODUyYjg1NQ==",
							},
						},
					},
					expectedSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-3",
						Type:       topology.Type{Name: "secret"},
						Data: topology.Data{
							"name": "test-secret-3",
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-secret",
								"namespace":      "test-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:secret/test-secret-3"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"annotations":       map[string]interface{}{"openshift.io/token-secret.value": "<redacted>"},
								"creationTimestamp": creationTimeFormatted,
								"name":              "test-secret-3",
								"namespace":         "test-namespace",
								"uid":               "test-secret-3",
								"resourceVersion":   "123",
							},
							"data": map[string]interface{}{
								"<data hash>": "ZTNiMGM0NDI5OGZjMWMxNDlhZmJmNGM4OTk2ZmI5MjQyN2FlNDFlNDY0OWI5MzRjYTQ5NTk5MWI3ODUyYjg1NQ==",
							},
						},
					},
				},
			} {
				t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled, kubernetesStatusEnabled), func(t *testing.T) {
					component := <-componentChannel
					if sourcePropertiesEnabled {
						if kubernetesStatusEnabled {
							assert.EqualValues(t, tc.expectedSPPlusStatus, component)
						} else {
							assert.EqualValues(t, tc.expectedSP, component)
						}
					} else {
						assert.EqualValues(t, tc.expectedNoSP, component)
					}
				})
			}
		}
	}
}

type MockSecretAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockSecretAPICollectorClient) GetSecrets() ([]coreV1.Secret, error) {
	secrets := make([]coreV1.Secret, 0)
	for i := 1; i <= 3; i++ {

		secret := coreV1.Secret{
			TypeMeta: v1.TypeMeta{
				Kind: "Secret",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-secret-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				UID:               types.UID(fmt.Sprintf("test-secret-%d", i)),
				GenerateName:      "",
				ResourceVersion:   "123",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationSecret,
					"openshift.io/token-secret.value":                  `{"secret":"data"`,
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
		}

		if i == 1 {
			secret.Data = map[string][]byte{
				"key1": asBase64("value1"),
				"key2": asBase64("longersecretvalue2"),
			}
		}

		if i != 3 {
			secret.Labels = map[string]string{
				"test": "label",
			}
		}

		secrets = append(secrets, secret)
	}

	return secrets, nil
}

func asBase64(s string) []byte {
	return []byte(base64.StdEncoding.EncodeToString([]byte(s)))
}
