package topologycollectors

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// K8sVolume is a wrapper around v1.Volume to be used when building up source properties for the component
type K8sVolume struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	v1.Volume         `json:"volume,omitempty" protobuf:"bytes,2,opt,name=volume"`
}

// String defers to the Volume.String function
func (k *K8sVolume) String() string {
	return k.Volume.String()
}

// Reset defers to the Volume.Reset function
func (k *K8sVolume) Reset() {
	k.Volume.Reset()
}

// ProtoMessage defers to the Volume.ProtoMessage function
func (k *K8sVolume) ProtoMessage() {
	k.Volume.ProtoMessage()
}

// DeepCopyObject is needed in order to fulfill the runtime.Object interface
func (k *K8sVolume) DeepCopyObject() runtime.Object {
	return nil
}

// K8sVolumeSource is a wrapper around v1.PersistentVolumeSource to be used when building up source properties for the component
type K8sVolumeSource struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta         `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	v1.PersistentVolumeSource `json:"source,omitempty" protobuf:"bytes,2,opt,name=source"`
}

// String defers to the PersistentVolumeSource.String function
func (k *K8sVolumeSource) String() string {
	return k.PersistentVolumeSource.String()
}

// Reset defers to the PersistentVolumeSource.Reset function
func (k *K8sVolumeSource) Reset() {
	k.PersistentVolumeSource.Reset()
}

// ProtoMessage defers to the PersistentVolumeSource.ProtoMessage function
func (k *K8sVolumeSource) ProtoMessage() {
	k.PersistentVolumeSource.ProtoMessage()
}

// DeepCopyObject is needed in order to fulfill the runtime.Object interface
func (k *K8sVolumeSource) DeepCopyObject() runtime.Object {
	return nil
}
