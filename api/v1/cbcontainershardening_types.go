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
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type CBContainersHardeningSpec struct {
	Version      string                            `json:"version,omitempty"`
	EnforcerSpec CBContainersHardeningEnforcerSpec `json:"enforcerSpec,omitempty"`
}

type CBContainersHardeningEnforcerSpec struct {
	DeploymentLabels       map[string]string                                `json:"deploymentLabels,omitempty"`
	PodTemplateLabels      map[string]string                                `json:"podTemplateLabels,omitempty"`
	DeploymentAnnotations  map[string]string                                `json:"deploymentAnnotations,omitempty"`
	PodTemplateAnnotations map[string]string                                `json:"podTemplateAnnotations,omitempty"`
	ReplicasCount          int32                                            `json:"replicasCount,omitempty"`
	ServiceAccountName     string                                           `json:"serviceAccountName,omitempty"`
	PriorityClassName      string                                           `json:"priorityClassName,omitempty"`
	Env                    map[string]string                                `json:"env,omitempty"`
	Image                  CBContainersHardeningEnforcerImageSpec           `json:"image,omitempty"`
	SecurityContext        CBContainersHardeningEnforcerSecurityContextSpec `json:"securityContext,omitempty"`
	Resources              coreV1.ResourceRequirements                      `json:"resources,omitempty"`
	Probes                 CBContainersHardeningEnforcerProbesSpec          `json:"probes,omitempty"`
}

type CBContainersHardeningEnforcerImageSpec struct {
	Repository string            `json:"repository,omitempty"`
	Tag        string            `json:"tag,omitempty"`
	PullPolicy coreV1.PullPolicy `json:"pullPolicy,omitempty"`
}

type CBContainersHardeningEnforcerSecurityContextSpec struct {
	AllowPrivilegeEscalation bool                `json:"allowPrivilegeEscalation,omitempty"`
	ReadOnlyRootFilesystem   bool                `json:"readOnlyRootFilesystem,omitempty"`
	RunAsUser                int64               `json:"runAsUser,omitempty"`
	CapabilitiesToAdd        []coreV1.Capability `json:"capabilitiesToAdd,omitempty"`
	CapabilitiesToDrop       []coreV1.Capability `json:"capabilitiesToDrop,omitempty"`
}

type CBContainersHardeningEnforcerProbesSpec struct {
	LivenessPath        string             `json:"livenessPath,omitempty"`
	ReadinessPath       string             `json:"readinessPath,omitempty"`
	Port                intstr.IntOrString `json:"port"`
	Scheme              coreV1.URIScheme   `json:"scheme,omitempty"`
	InitialDelaySeconds int32              `json:"initialDelaySeconds,omitempty"`
	TimeoutSeconds      int32              `json:"timeoutSeconds,omitempty"`
	PeriodSeconds       int32              `json:"periodSeconds,omitempty"`
	SuccessThreshold    int32              `json:"successThreshold,omitempty"`
	FailureThreshold    int32              `json:"failureThreshold,omitempty"`
}

// CBContainersHardeningStatus defines the observed state of CBContainersHardening
type CBContainersHardeningStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CBContainersHardening is the Schema for the cbcontainershardenings API
type CBContainersHardening struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CBContainersHardeningSpec   `json:"spec,omitempty"`
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
