/*
Copyright 2024. rkgcloud.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reconciler.io/runtime/apis"
)

type ContainerInfo struct {
	// ImageURL refers to the URL of the container image
	ImageURL string `json:"imageURL"`

	// ImagePullSecretName refers to a k8s secret used in authenticating to the container image registry
	ImagePullSecretName corev1.LocalObjectReference `json:"imagePullSecretName,omitempty"`

	// Args refer to an array of args used for the container
	Args []string `json:"args,omitempty"`
}

// PodSyncSpec defines the desired state of PodSync
type PodSyncSpec struct {
	// Containers refer to a list of arguments used by the container
	Containers []ContainerInfo `json:"containers"`
}

// PodSyncStatus defines the observed state of PodSync
type PodSyncStatus struct {
	apis.Status `json:",inline"`

	// PodName refers to the generated pod's name
	PodName string `json:"podName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`
// +kubebuilder:printcolumn:name="Detail",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].message`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// PodSync is the Schema for the podsyncs API
type PodSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodSyncSpec   `json:"spec,omitempty"`
	Status PodSyncStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PodSyncList contains a list of PodSync
type PodSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodSync{}, &PodSyncList{})
}
