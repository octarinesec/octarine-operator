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

// CBContainersAgentSpec defines the desired state of CBContainersAgentSpec
type CBContainersAgentSpec struct {
	Account        string                     `json:"account,required"`
	ClusterName    string                     `json:"clusterName,required"`
	Version        string                     `json:"version,required"`
	Features       AgentFeaturesList          `json:"features,required"`
	ApiGatewaySpec CBContainersApiGatewaySpec `json:"apiGatewaySpec,required"`
	// +kubebuilder:default:=<>
	GatewayTLS CBContainersGatewayTLS `json:"gatewayTLS,omitempty"`
	// +kubebuilder:default:=<>
	CoreSpec CBContainersCoreSpec `json:"coreSpec,omitempty"`
	// +kubebuilder:default:=<>
	HardeningSpec CBContainersHardeningSpec `json:"hardeningSpec,omitempty"`
	// +kubebuilder:default:=<>
	RuntimeSpec CBContainersRuntimeSpec `json:"runtimeSpec,omitempty"`
}

// CBContainersAgentStatus defines the observed state of CBContainersAgent
type CBContainersAgentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=cbcontainersagents,scope=Cluster

// CBContainersAgent is the Schema for the cbcontainersagents API
//+kubebuilder:subresource:status
type CBContainersAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CBContainersAgentSpec   `json:"spec,omitempty"`
	Status CBContainersAgentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CBContainersAgentList contains a list of CBContainersAgent
type CBContainersAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CBContainersAgent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CBContainersAgent{}, &CBContainersAgentList{})
}
