// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com/).
// Copyright 2019-present StackState

//go:build kubeapiserver
// +build kubeapiserver

package apiserver

import (
	appsV1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	batchV1B1 "k8s.io/api/batch/v1beta1"
	coreV1 "k8s.io/api/core/v1"
	extensionsV1B "k8s.io/api/extensions/v1beta1"
	netV1 "k8s.io/api/networking/v1"
	storageV1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/version"
)

type APICollectorClient interface {
	GetDaemonSets() ([]appsV1.DaemonSet, error)
	GetReplicaSets() ([]appsV1.ReplicaSet, error)
	GetDeployments() ([]appsV1.Deployment, error)
	GetStatefulSets() ([]appsV1.StatefulSet, error)
	GetJobs() ([]batchV1.Job, error)
	GetCronJobsV1B1() ([]batchV1B1.CronJob, error)
	GetCronJobsV1() ([]batchV1.CronJob, error)
	GetEndpoints() ([]coreV1.Endpoints, error)
	GetNodes() ([]coreV1.Node, error)
	GetPods() ([]coreV1.Pod, error)
	GetServices() ([]coreV1.Service, error)
	GetIngressesExtV1B1() ([]extensionsV1B.Ingress, error)
	GetIngressesNetV1() ([]netV1.Ingress, error)
	GetConfigMaps() ([]coreV1.ConfigMap, error)
	GetSecrets() ([]coreV1.Secret, error)
	GetNamespaces() ([]coreV1.Namespace, error)
	GetPersistentVolumes() ([]coreV1.PersistentVolume, error)
	GetPersistentVolumeClaims() ([]coreV1.PersistentVolumeClaim, error)
	GetVolumeAttachments() ([]storageV1.VolumeAttachment, error)
	GetVersion() (*version.Info, error)
}
