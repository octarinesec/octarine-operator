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
	// +kubebuilder:default:={{operator: "Exists"}}
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

	// DefaultImagesRegistry is the default registry to use with the agent images
	DefaultImagesRegistry string `json:"defaultImagesRegistry,omitempty"`

	// Proxy controls the optional centralized HTTP & HTTPS proxy settings, that can be applied
	// to all components at once. One can still have a per-component proxy settings by using the
	// good old environment variables. However, here we have an additional advantage of taking
	// care of determining the necessary `NO_PROXY` settings.
	//
	Proxy *CBContainersProxySettings `json:"proxy,omitempty"`
}

func (s CBContainersComponentsSettings) ShouldCreateDefaultImagePullSecrets() bool {
	// this field is TRUE by default, so if the user did not set it, return true
	if s.CreateDefaultImagePullSecrets == nil {
		return true
	}
	return *s.CreateDefaultImagePullSecrets
}

type CBContainersProxySettings struct {
	// Enabled controls if the proxy settings are applied or not
	//
	// +kubebuilder:default:=false
	Enabled *bool `json:"enabled,omitempty"`

	// HttpProxy points to the URL of the HTTP proxy.
	// If set, it'll result in an additional HTTP_PROXY environment variable defined for all
	// components. When a component already has HTTP_PROXY defined through the CRD, HttpProxy won't
	// be used, using the component's original HTTP_PROXY value instead.
	HttpProxy *string `json:"httpProxy,omitempty"`

	// HttpsProxy points to the URL of the HTTPS proxy.
	// If set, it'll result in an additional HTTPS_PROXY environment variable defined for all
	// components. When a component already has HTTPS_PROXY defined through the CRD, HttpsProxy won't
	// be used, using the component's original HTTPS_PROXY value instead.
	HttpsProxy *string `json:"httpsProxy,omitempty"`

	// NoProxy can contain a comma separated list of hosts to which all components can connect
	// without using a proxy. If set, it'll result in an additional NO_PROXY environment variable
	// defined for all components. When a component already has NO_PROXY defined through the CRD, NoProxy won't
	// be used, using the component's original NO_PROXY value instead.
	NoProxy *string `json:"noProxy,omitempty"`

	// NoProxySuffix can be an empty string or contain a comma separated list of hosts which can
	// be safely appended to the `NoProxy` values or to the component specific NO_PROXY environment
	// variable. NoProxySuffix is defaulted to a list of the Kubernetes API server IP addresses and
	// to the service domain suffix of the installation namespace (usually
	// cbcontainers-dataplane.svc.cluster.local). It's exposed more as a means by which to control
	// the defaults.
	NoProxySuffix *string `json:"noProxySuffix,omitempty"`
}

// CBContainersAgentStatus defines the observed state of CBContainersAgent
type CBContainersAgentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ObservedGeneration is the last Custom resource generation that was fully reconciled.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=cbcontainersagents,scope=Cluster
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version",description="Version of the deployed agent"
// +kubebuilder:printcolumn:name="Cluster image scanning",type="boolean",JSONPath=".spec.components.clusterScanning.enabled",description="Whether cluster image scanning is enabled"
// +kubebuilder:printcolumn:name="Runtime protection",type="string",JSONPath=".spec.components.runtimeProtection.enabled",description="Whether runtime protection is enabled"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

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
