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

type CBContainersRuntimeResolverSpec struct {
	EventsGatewaySpec CBContainersEventsGatewaySpec `json:"eventsGatewaySpec,required"`
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DeploymentAnnotations map[string]string `json:"deploymentAnnotations,omitempty"`
	// +kubebuilder:default:={prometheus.io/scrape: "false", prometheus.io/port: "7071"}
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:=1
	ReplicasCount *int32 `json:"replicasCount,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/runtime-kubernetes-resolver"}
	Image CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "200m"}, limits: {memory: "1024Mi", cpu: "900m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes CBContainersHTTPProbesSpec `json:"probes,omitempty"`
	// +kubebuilder:default:=<>
	Prometheus CBContainersPrometheusSpec `json:"prometheus,omitempty"`
}

type CBContainersRuntimeSensorSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DaemonSetAnnotations map[string]string `json:"daemonSetAnnotations,omitempty"`
	// +kubebuilder:default:={prometheus.io/scrape: "false", prometheus.io/port: "7071"}
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/runtime-kubernetes-sensor"}
	Image CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "1024Mi", cpu: "500m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes CBContainersFileProbesSpec `json:"probes,omitempty"`
	// +kubebuilder:default:=<>
	Prometheus CBContainersPrometheusSpec `json:"prometheus,omitempty"`
	// +kubebuilder:default:=2
	VerbosityLevel *int `json:"verbosity_level,omitempty"`
}

// CBContainersRuntimeSpec defines the desired state of CBContainersRuntime
type CBContainersRuntimeSpec struct {
	Version string `json:"version,required"`
	// +kubebuilder:default:="cbcontainers-access-token"
	AccessTokenSecretName string                          `json:"accessTokenSecretName,omitempty"`
	ResolverSpec          CBContainersRuntimeResolverSpec `json:"resolverSpec,omitempty"`
	// +kubebuilder:default:=<>
	SensorSpec CBContainersRuntimeSensorSpec `json:"sensorSpec,omitempty"`
	// +kubebuilder:default:=443
	InternalGrpcPort int32 `json:"internalGrpcPort,omitempty"`
}

// CBContainersRuntimeStatus defines the observed state of CBContainersRuntime
type CBContainersRuntimeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// CBContainersRuntime is the Schema for the cbcontainersruntimes API
type CBContainersRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CBContainersRuntimeSpec   `json:"spec,omitempty"`
	Status CBContainersRuntimeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CBContainersRuntimeList contains a list of CBContainersRuntime
type CBContainersRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CBContainersRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CBContainersRuntime{}, &CBContainersRuntimeList{})
}
