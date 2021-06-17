/*
Copyright 2021.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CBContainersClusterSpec defines the desired state of CBContainersCluster
type CBContainersClusterSpec struct {
	Account           string                        `json:"account,required"`
	ClusterName       string                        `json:"clusterName,required"`
	ApiGatewaySpec    CBContainersApiGatewaySpec    `json:"apiGatewaySpec,required"`
	EventsGatewaySpec CBContainersEventsGatewaySpec `json:"eventsGatewaySpec,required"`
}

// CBContainersClusterStatus defines the observed state of CBContainersCluster
type CBContainersClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=cbcontainersclusters,scope=Cluster

// CBContainersCluster is the Schema for the cbcontainersclusters API
//+kubebuilder:subresource:status
type CBContainersCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CBContainersClusterSpec   `json:"spec,omitempty"`
	Status CBContainersClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CBContainersClusterList contains a list of CBContainersCluster
type CBContainersClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CBContainersCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CBContainersCluster{}, &CBContainersClusterList{})
}
