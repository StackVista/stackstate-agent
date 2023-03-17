// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.

//go:build kubeapiserver
// +build kubeapiserver

package apiserver

import (
	"context"
	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetConfigMaps() retrieves all the ConfigMaps in the Kubernetes / OpenShift cluster across all namespaces.
func (c *APIClient) GetConfigMaps() ([]coreV1.ConfigMap, error) {
	cmList, err := c.Cl.CoreV1().ConfigMaps(metaV1.NamespaceAll).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []coreV1.ConfigMap{}, err
	}

	return cmList.Items, nil
}

// GetSecrets() retrieves all the Secrets in the Kubernetes / OpenShift cluster across all namespaces.
func (c *APIClient) GetSecrets() ([]coreV1.Secret, error) {
	secretList, err := c.Cl.CoreV1().Secrets(metaV1.NamespaceAll).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []coreV1.Secret{}, err
	}

	return secretList.Items, nil
}

// GetNamespaces() retrieves all the ConfigMaps in the Kubernetes / OpenShift cluster across all namespaces.
func (c *APIClient) GetNamespaces() ([]coreV1.Namespace, error) {
	cmList, err := c.Cl.CoreV1().Namespaces().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []coreV1.Namespace{}, err
	}

	return cmList.Items, nil
}

// GetPersistentVolumes() retrieves all the PersistentVolumes in the Kubernetes / OpenShift cluster across all namespaces.
func (c *APIClient) GetPersistentVolumes() ([]coreV1.PersistentVolume, error) {
	pvList, err := c.Cl.CoreV1().PersistentVolumes().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []coreV1.PersistentVolume{}, err
	}

	return pvList.Items, nil
}

// GetPersistentVolumeClaims() retrieves all the PersistentVolumeClaims in the Kubernetes / OpenShift cluster across all namespaces.
func (c *APIClient) GetPersistentVolumeClaims() ([]coreV1.PersistentVolumeClaim, error) {
	pvList, err := c.Cl.CoreV1().PersistentVolumeClaims(metaV1.NamespaceAll).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []coreV1.PersistentVolumeClaim{}, err
	}

	return pvList.Items, nil
}

// GetVolumeAttachments() retrieves all the VolumeAttachments in the Kubernetes / OpenShift cluster across all namespaces.
func (c *APIClient) GetVolumeAttachments() ([]storageV1.VolumeAttachment, error) {
	vaList, err := c.Cl.StorageV1().VolumeAttachments().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []storageV1.VolumeAttachment{}, err
	}

	return vaList.Items, nil
}
