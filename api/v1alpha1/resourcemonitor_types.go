/*
Copyright 2024.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)



// ResourceMonitorSpec defines the desired state of ResourceMonitor
type ResourceMonitorSpec struct {
	// Number of pods to trigger alert
    PodThreshold int `json:"podThreshold"`
}

// ResourceMonitorStatus defines the observed state of ResourceMonitor
type ResourceMonitorStatus struct {
	// Define observed state of the cluster
    PodsCount int `json:"podsCount"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ResourceMonitor is the Schema for the resourcemonitors API
type ResourceMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceMonitorSpec   `json:"spec,omitempty"`
	Status ResourceMonitorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ResourceMonitorList contains a list of ResourceMonitor
type ResourceMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceMonitor{}, &ResourceMonitorList{})
}
