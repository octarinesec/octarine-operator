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
	"github.com/vmware/cbcontainers-operator/api/v1/common_specs"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type CBContainersHardeningSpec struct {
	Version           string                                     `json:"version,required"`
	EventsGatewaySpec common_specs.CBContainersEventsGatewaySpec `json:"eventsGatewaySpec,required"`
	// +kubebuilder:default:="cbcontainers-access-token"
	AccessTokenSecretName string `json:"accessTokenSecretName,omitempty"`
	// +kubebuilder:default:=<>
	EnforcerSpec CBContainersHardeningEnforcerSpec `json:"enforcerSpec,omitempty"`
	// +kubebuilder:default:=<>
	StateReporterSpec CBContainersHardeningStateReporterSpec `json:"stateReporterSpec,omitempty"`
}

type CBContainersHardeningStateReporterSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DeploymentAnnotations map[string]string `json:"deploymentAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/guardrails-state-reporter"}
	Image common_specs.CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes common_specs.CBContainersHTTPProbesSpec `json:"probes,omitempty"`
}

type CBContainersHardeningEnforcerSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DeploymentAnnotations map[string]string `json:"deploymentAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:=1
	ReplicasCount *int32 `json:"replicasCount,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:=<>
	Prometheus common_specs.CBContainersPrometheusSpec `json:"prometheus,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/guardrails-enforcer"}
	Image common_specs.CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes common_specs.CBContainersHTTPProbesSpec `json:"probes,omitempty"`
	// +kubebuilder:default:=5
	WebhookTimeoutSeconds int32 `json:"webhookTimeoutSeconds,omitempty"`
}

// CBContainersHardeningStatus defines the observed state of CBContainersHardening
type CBContainersHardeningStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=cbcontainershardenings,scope=Cluster

// CBContainersHardening is the Schema for the cbcontainershardenings API
type CBContainersHardening struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CBContainersHardeningSpec   `json:"spec,required"`
	Status CBContainersHardeningStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CBContainersHardeningList contains a list of CBContainersHardening
type CBContainersHardeningList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CBContainersHardening `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CBContainersHardening{}, &CBContainersHardeningList{})
}
