//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2020 The RUIJIE Authors.

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

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PatroniCluster) DeepCopyInto(out *PatroniCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.PatroniClusterSpec = in.PatroniClusterSpec
	out.PatroniClusterStatus = in.PatroniClusterStatus
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PatroniCluster.
func (in *PatroniCluster) DeepCopy() *PatroniCluster {
	if in == nil {
		return nil
	}
	out := new(PatroniCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PatroniCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PatroniClusterList) DeepCopyInto(out *PatroniClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]*PatroniCluster, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(PatroniCluster)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PatroniClusterList.
func (in *PatroniClusterList) DeepCopy() *PatroniClusterList {
	if in == nil {
		return nil
	}
	out := new(PatroniClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PatroniClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PatroniClusterSpec) DeepCopyInto(out *PatroniClusterSpec) {
	*out = *in
	if in.NodeList != nil {
		in, out := &in.NodeList, &out.NodeList
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PatroniClusterSpec.
func (in *PatroniClusterSpec) DeepCopy() *PatroniClusterSpec {
	if in == nil {
		return nil
	}
	out := new(PatroniClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PatroniClusterStatus) DeepCopyInto(out *PatroniClusterStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PatroniClusterStatus.
func (in *PatroniClusterStatus) DeepCopy() *PatroniClusterStatus {
	if in == nil {
		return nil
	}
	out := new(PatroniClusterStatus)
	in.DeepCopyInto(out)
	return out
}
