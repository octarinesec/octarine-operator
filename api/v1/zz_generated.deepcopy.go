// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersCluster) DeepCopyInto(out *CBContainersCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersCluster.
func (in *CBContainersCluster) DeepCopy() *CBContainersCluster {
	if in == nil {
		return nil
	}
	out := new(CBContainersCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CBContainersCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersClusterApiGatewaySpec) DeepCopyInto(out *CBContainersClusterApiGatewaySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersClusterApiGatewaySpec.
func (in *CBContainersClusterApiGatewaySpec) DeepCopy() *CBContainersClusterApiGatewaySpec {
	if in == nil {
		return nil
	}
	out := new(CBContainersClusterApiGatewaySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersClusterEventsGatewaySpec) DeepCopyInto(out *CBContainersClusterEventsGatewaySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersClusterEventsGatewaySpec.
func (in *CBContainersClusterEventsGatewaySpec) DeepCopy() *CBContainersClusterEventsGatewaySpec {
	if in == nil {
		return nil
	}
	out := new(CBContainersClusterEventsGatewaySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersClusterList) DeepCopyInto(out *CBContainersClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CBContainersCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersClusterList.
func (in *CBContainersClusterList) DeepCopy() *CBContainersClusterList {
	if in == nil {
		return nil
	}
	out := new(CBContainersClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CBContainersClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersClusterSpec) DeepCopyInto(out *CBContainersClusterSpec) {
	*out = *in
	out.ApiGatewaySpec = in.ApiGatewaySpec
	out.EventsGatewaySpec = in.EventsGatewaySpec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersClusterSpec.
func (in *CBContainersClusterSpec) DeepCopy() *CBContainersClusterSpec {
	if in == nil {
		return nil
	}
	out := new(CBContainersClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersClusterStatus) DeepCopyInto(out *CBContainersClusterStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersClusterStatus.
func (in *CBContainersClusterStatus) DeepCopy() *CBContainersClusterStatus {
	if in == nil {
		return nil
	}
	out := new(CBContainersClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersHardening) DeepCopyInto(out *CBContainersHardening) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersHardening.
func (in *CBContainersHardening) DeepCopy() *CBContainersHardening {
	if in == nil {
		return nil
	}
	out := new(CBContainersHardening)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CBContainersHardening) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersHardeningEnforcerSpec) DeepCopyInto(out *CBContainersHardeningEnforcerSpec) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.DeploymentAnnotations != nil {
		in, out := &in.DeploymentAnnotations, &out.DeploymentAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PodTemplateAnnotations != nil {
		in, out := &in.PodTemplateAnnotations, &out.PodTemplateAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.Image = in.Image
	in.Resources.DeepCopyInto(&out.Resources)
	out.Probes = in.Probes
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersHardeningEnforcerSpec.
func (in *CBContainersHardeningEnforcerSpec) DeepCopy() *CBContainersHardeningEnforcerSpec {
	if in == nil {
		return nil
	}
	out := new(CBContainersHardeningEnforcerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersHardeningEventsGatewaySpec) DeepCopyInto(out *CBContainersHardeningEventsGatewaySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersHardeningEventsGatewaySpec.
func (in *CBContainersHardeningEventsGatewaySpec) DeepCopy() *CBContainersHardeningEventsGatewaySpec {
	if in == nil {
		return nil
	}
	out := new(CBContainersHardeningEventsGatewaySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersHardeningImageSpec) DeepCopyInto(out *CBContainersHardeningImageSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersHardeningImageSpec.
func (in *CBContainersHardeningImageSpec) DeepCopy() *CBContainersHardeningImageSpec {
	if in == nil {
		return nil
	}
	out := new(CBContainersHardeningImageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersHardeningList) DeepCopyInto(out *CBContainersHardeningList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CBContainersHardening, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersHardeningList.
func (in *CBContainersHardeningList) DeepCopy() *CBContainersHardeningList {
	if in == nil {
		return nil
	}
	out := new(CBContainersHardeningList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CBContainersHardeningList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersHardeningProbesSpec) DeepCopyInto(out *CBContainersHardeningProbesSpec) {
	*out = *in
	out.Port = in.Port
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersHardeningProbesSpec.
func (in *CBContainersHardeningProbesSpec) DeepCopy() *CBContainersHardeningProbesSpec {
	if in == nil {
		return nil
	}
	out := new(CBContainersHardeningProbesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersHardeningSpec) DeepCopyInto(out *CBContainersHardeningSpec) {
	*out = *in
	in.EnforcerSpec.DeepCopyInto(&out.EnforcerSpec)
	in.StateReporterSpec.DeepCopyInto(&out.StateReporterSpec)
	out.EventsGatewaySpec = in.EventsGatewaySpec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersHardeningSpec.
func (in *CBContainersHardeningSpec) DeepCopy() *CBContainersHardeningSpec {
	if in == nil {
		return nil
	}
	out := new(CBContainersHardeningSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersHardeningStateReporterSpec) DeepCopyInto(out *CBContainersHardeningStateReporterSpec) {
	*out = *in
	if in.DeploymentLabels != nil {
		in, out := &in.DeploymentLabels, &out.DeploymentLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PodTemplateLabels != nil {
		in, out := &in.PodTemplateLabels, &out.PodTemplateLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.DeploymentAnnotations != nil {
		in, out := &in.DeploymentAnnotations, &out.DeploymentAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PodTemplateAnnotations != nil {
		in, out := &in.PodTemplateAnnotations, &out.PodTemplateAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.Image = in.Image
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.Resources.DeepCopyInto(&out.Resources)
	out.Probes = in.Probes
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersHardeningStateReporterSpec.
func (in *CBContainersHardeningStateReporterSpec) DeepCopy() *CBContainersHardeningStateReporterSpec {
	if in == nil {
		return nil
	}
	out := new(CBContainersHardeningStateReporterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CBContainersHardeningStatus) DeepCopyInto(out *CBContainersHardeningStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CBContainersHardeningStatus.
func (in *CBContainersHardeningStatus) DeepCopy() *CBContainersHardeningStatus {
	if in == nil {
		return nil
	}
	out := new(CBContainersHardeningStatus)
	in.DeepCopyInto(out)
	return out
}
