//go:build kubeapiserver

package topologycollectors

import (
	"k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type CronJobInterface interface {
	GetConcurrencyPolicy() string
	GetCreationTimestamp() metav1.Time
	GetGenerateName() string
	GetKind() string
	GetKubernetesObject() MarshalableKubernetesObject
	GetName() string
	GetNamespace() string
	GetObjectMeta() metav1.ObjectMeta
	GetSchedule() string
	GetString() string
	GetUID() types.UID
}

type CronJobV1B1 struct {
	o v1beta1.CronJob
	MarshalableKubernetesObject
}

type CronJobV1 struct {
	o v1.CronJob
	MarshalableKubernetesObject
}

/* CronJobV1B1 */

func (in CronJobV1B1) GetNamespace() string {
	return in.o.Namespace
}

func (in CronJobV1B1) GetGenerateName() string {
	return in.o.GenerateName
}

func (in CronJobV1B1) GetObjectMeta() metav1.ObjectMeta {
	return in.o.ObjectMeta
}

func (in CronJobV1B1) GetString() string {
	return in.o.String()
}

func (in CronJobV1B1) GetName() string {
	return in.o.Name
}

func (in CronJobV1B1) GetKubernetesObject() MarshalableKubernetesObject {
	return &in.o
}

func (in CronJobV1B1) GetKind() string {
	return in.o.Kind
}

func (in CronJobV1B1) GetCreationTimestamp() metav1.Time {
	return in.o.CreationTimestamp
}

func (in CronJobV1B1) GetConcurrencyPolicy() string {
	return string(in.o.Spec.ConcurrencyPolicy)
}

func (in CronJobV1B1) GetSchedule() string {
	return in.o.Spec.Schedule
}

func (in CronJobV1B1) GetUID() types.UID {
	return in.o.UID
}

/* CronJobV1 */

func (in CronJobV1) GetNamespace() string {
	return in.o.Namespace
}

func (in CronJobV1) GetGenerateName() string {
	return in.o.GenerateName
}

func (in CronJobV1) GetObjectMeta() metav1.ObjectMeta {
	return in.o.ObjectMeta
}

func (in CronJobV1) GetString() string {
	return in.o.String()
}

func (in CronJobV1) GetName() string {
	return in.o.Name
}

func (in CronJobV1) GetKubernetesObject() MarshalableKubernetesObject {
	return &in.o
}

func (in CronJobV1) GetKind() string {
	return in.o.Kind
}

func (in CronJobV1) GetCreationTimestamp() metav1.Time {
	return in.o.CreationTimestamp
}

func (in CronJobV1) GetConcurrencyPolicy() string {
	return string(in.o.Spec.ConcurrencyPolicy)
}

func (in CronJobV1) GetSchedule() string {
	return in.o.Spec.Schedule
}

func (in CronJobV1) GetUID() types.UID {
	return in.o.UID
}
