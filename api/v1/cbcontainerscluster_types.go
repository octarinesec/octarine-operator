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

// CBContainersClusterSpec defines the desired state of CBContainersCluster
type CBContainersClusterSpec struct {
	Account           string                        `json:"account,required"`
	ClusterName       string                        `json:"clusterName,required"`
	Version           string                        `json:"version,required"`
	ApiGatewaySpec    CBContainersApiGatewaySpec    `json:"apiGatewaySpec,required"`
	EventsGatewaySpec CBContainersEventsGatewaySpec `json:"eventsGatewaySpec,required"`
	// +kubebuilder:default:=<>
	GatewayTLS CBContainersGatewayTLS `json:"gatewayTls,omitempty"`
	// +kubebuilder:default:=<>
	MonitorSpec CBContainersClusterMonitorSpec `json:"monitorSpec,omitempty"`
}

type CBContainersClusterMonitorSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DeploymentAnnotations map[string]string `json:"deploymentAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/monitor"}
	Image CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes CBContainersHTTPProbesSpec `json:"probes,omitempty"`
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
