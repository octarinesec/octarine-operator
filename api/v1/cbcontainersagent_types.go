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
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CBContainersAgentSpec defines the desired state of CBContainersAgentSpec
type CBContainersAgentSpec struct {
	Account     string                   `json:"account,required"`
	ClusterName string                   `json:"clusterName,required"`
	Version     string                   `json:"version,required"`
	Gateways    CBContainersGatewaysSpec `json:"gateways,required"`
	// The field below remains to avoid moving the CRD from v1 to v2.
	// It MUST not be used as agent namespace should be controlled outside the operator itself.
	// This is because a custom namespace in the CRD requires high privileges by the operator across the whole cluster to be able to "switch" namespaces on demand.

	// +kubebuilder:default:="cbcontainers-dataplane"
	// Namespace is deprecated and the value has no effect. Do not use.
	// Deprecated: The operator and agent always run in the same namespace. See documentation for ways to customize this namespace.
	Namespace string `json:"namespace,omitempty"`
	// +kubebuilder:default:="cbcontainers-access-token"
	AccessTokenSecretName string `json:"accessTokenSecretName,omitempty"`
	// +kubebuilder:default:=<>
	Components CBContainersComponentsSpec `json:"components,omitempty"`
}

type CBContainersComponentsSpec struct {
	// +kubebuilder:default:=<>
	Basic CBContainersBasicSpec `json:"basic,omitempty"`
	// +kubebuilder:default:=<>
	RuntimeProtection CBContainersRuntimeProtectionSpec `json:"runtimeProtection,omitempty"`
	Cndr              *CBContainersCndrSpec             `json:"cndr,omitempty"`
	// +kubebuilder:default:=<>
	ClusterScanning CBContainersClusterScanningSpec `json:"clusterScanning,omitempty"`
	// +kubebuilder:default:=<>
	Settings CBContainersComponentsSettings `json:"settings,omitempty"`
}

type CBContainersComponentsSettings struct {
	// +kubebuilder:default:=[{key: "operator", value: "Exists"}]
	DaemonSetsTolerations []coreV1.Toleration `json:"daemonSetsTolerations,omitempty"`
	// CreateDefaultImagePullSecrets controls whether or not to create the secrets
	// needed to pull the containers images from the default repository.
	//
	// This should be set to true if the user does not override the default images for the services.
	//
	// +kubebuilder:default:=true
	CreateDefaultImagePullSecrets *bool `json:"createDefaultImagePullSecrets,omitempty"`
	// ImagePullSecrets is a list of image pull secret names, which will be used to pull the container image(s)
	// for the Agent deployment.
	//
	// These secrets will be shared for all containers.
	//
	// The secrets must already exist.
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
}

func (s CBContainersComponentsSettings) ShouldCreateDefaultImagePullSecrets() bool {
	// this field is TRUE by default, so if the user did not set it, return true
	if s.CreateDefaultImagePullSecrets == nil {
		return true
	}
	return *s.CreateDefaultImagePullSecrets
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
// +kubebuilder:subresource:status
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
