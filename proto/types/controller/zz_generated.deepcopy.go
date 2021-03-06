//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package controller

import (
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Namespace) DeepCopyInto(out *Namespace) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.TypeMeta != nil {
		in, out := &in.TypeMeta, &out.TypeMeta
		*out = new(v1.TypeMeta)
		**out = **in
	}
	if in.ObjectMeta != nil {
		in, out := &in.ObjectMeta, &out.ObjectMeta
		*out = new(v1.ObjectMeta)
		(*in).DeepCopyInto(*out)
	}
	if in.Spec != nil {
		in, out := &in.Spec, &out.Spec
		*out = new(NamespaceSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Status != nil {
		in, out := &in.Status, &out.Status
		*out = new(NamespaceStatus)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Namespace.
func (in *Namespace) DeepCopy() *Namespace {
	if in == nil {
		return nil
	}
	out := new(Namespace)
	in.DeepCopyInto(out)
	return out
}

// GetObjectKind returns the runtime ObjectKind.
func (in *Namespace) GetObjectKind() schema.ObjectKind {
	return in.TypeMeta.GetObjectKind()
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Namespace) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespaceCondition) DeepCopyInto(out *NamespaceCondition) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.LastUpdateTime != nil {
		in, out := &in.LastUpdateTime, &out.LastUpdateTime
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespaceCondition.
func (in *NamespaceCondition) DeepCopy() *NamespaceCondition {
	if in == nil {
		return nil
	}
	out := new(NamespaceCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespaceList) DeepCopyInto(out *NamespaceList) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.TypeMeta != nil {
		in, out := &in.TypeMeta, &out.TypeMeta
		*out = new(v1.TypeMeta)
		out = in
	}
	if in.ListMeta != nil {
		in, out := &in.ListMeta, &out.ListMeta
		*out = new(v1.ListMeta)
		(*in).DeepCopyInto(*out)
	}
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]*Namespace, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Namespace)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespaceList.
func (in *NamespaceList) DeepCopy() *NamespaceList {
	if in == nil {
		return nil
	}
	out := new(NamespaceList)
	in.DeepCopyInto(out)
	return out
}

// GetObjectKind returns the runtime ObjectKind.
func (in *NamespaceList) GetObjectKind() schema.ObjectKind {
	return in.TypeMeta.GetObjectKind()
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NamespaceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespaceSpec) DeepCopyInto(out *NamespaceSpec) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.Init != nil {
		in, out := &in.Init, &out.Init
		*out = make([]*StepTemplate, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepTemplate)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.NamespaceMeta != nil {
		in, out := &in.NamespaceMeta, &out.NamespaceMeta
		*out = new(v1.ObjectMeta)
		(*in).DeepCopyInto(*out)
	}
	if in.PostCreate != nil {
		in, out := &in.PostCreate, &out.PostCreate
		*out = make([]*StepTemplate, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepTemplate)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.PreDelete != nil {
		in, out := &in.PreDelete, &out.PreDelete
		*out = make([]*StepTemplate, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepTemplate)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespaceSpec.
func (in *NamespaceSpec) DeepCopy() *NamespaceSpec {
	if in == nil {
		return nil
	}
	out := new(NamespaceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespaceStatus) DeepCopyInto(out *NamespaceStatus) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]*NamespaceCondition, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(NamespaceCondition)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Init != nil {
		in, out := &in.Init, &out.Init
		*out = make([]*StepStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepStatus)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.NamespaceRef != nil {
		in, out := &in.NamespaceRef, &out.NamespaceRef
		*out = new(corev1.ObjectReference)
		**out = **in
	}
	if in.PostCreate != nil {
		in, out := &in.PostCreate, &out.PostCreate
		*out = make([]*StepStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepStatus)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.PreDelete != nil {
		in, out := &in.PreDelete, &out.PreDelete
		*out = make([]*StepStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepStatus)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespaceStatus.
func (in *NamespaceStatus) DeepCopy() *NamespaceStatus {
	if in == nil {
		return nil
	}
	out := new(NamespaceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StepObject) DeepCopyInto(out *StepObject) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.TypeMeta != nil {
		in, out := &in.TypeMeta, &out.TypeMeta
		*out = new(v1.TypeMeta)
		**out = **in
	}
	if in.ObjectMeta != nil {
		in, out := &in.ObjectMeta, &out.ObjectMeta
		*out = new(v1.ObjectMeta)
		(*in).DeepCopyInto(*out)
	}
	if in.Raw != nil {
		in, out := &in.Raw, &out.Raw
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StepObject.
func (in *StepObject) DeepCopy() *StepObject {
	if in == nil {
		return nil
	}
	out := new(StepObject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StepStatus) DeepCopyInto(out *StepStatus) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.ObjectRef != nil {
		in, out := &in.ObjectRef, &out.ObjectRef
		*out = new(corev1.ObjectReference)
		**out = **in
	}
	if in.JobRef != nil {
		in, out := &in.JobRef, &out.JobRef
		*out = new(corev1.ObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StepStatus.
func (in *StepStatus) DeepCopy() *StepStatus {
	if in == nil {
		return nil
	}
	out := new(StepStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StepTemplate) DeepCopyInto(out *StepTemplate) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.Object != nil {
		in, out := &in.Object, &out.Object
		*out = new(StepObject)
		(*in).DeepCopyInto(*out)
	}
	if in.JobTemplate != nil {
		in, out := &in.JobTemplate, &out.JobTemplate
		*out = new(v1beta1.JobTemplateSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StepTemplate.
func (in *StepTemplate) DeepCopy() *StepTemplate {
	if in == nil {
		return nil
	}
	out := new(StepTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Task) DeepCopyInto(out *Task) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.TypeMeta != nil {
		in, out := &in.TypeMeta, &out.TypeMeta
		*out = new(v1.TypeMeta)
		**out = **in
	}
	if in.ObjectMeta != nil {
		in, out := &in.ObjectMeta, &out.ObjectMeta
		*out = new(v1.ObjectMeta)
		(*in).DeepCopyInto(*out)
	}
	if in.Spec != nil {
		in, out := &in.Spec, &out.Spec
		*out = new(TaskSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Status != nil {
		in, out := &in.Status, &out.Status
		*out = new(TaskStatus)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Task.
func (in *Task) DeepCopy() *Task {
	if in == nil {
		return nil
	}
	out := new(Task)
	in.DeepCopyInto(out)
	return out
}

// GetObjectKind returns the runtime ObjectKind.
func (in *Task) GetObjectKind() schema.ObjectKind {
	return in.TypeMeta.GetObjectKind()
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Task) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TaskCondition) DeepCopyInto(out *TaskCondition) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.LastUpdateTime != nil {
		in, out := &in.LastUpdateTime, &out.LastUpdateTime
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskCondition.
func (in *TaskCondition) DeepCopy() *TaskCondition {
	if in == nil {
		return nil
	}
	out := new(TaskCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TaskList) DeepCopyInto(out *TaskList) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.TypeMeta != nil {
		in, out := &in.TypeMeta, &out.TypeMeta
		*out = new(v1.TypeMeta)
		**out = **in
	}
	if in.ListMeta != nil {
		in, out := &in.ListMeta, &out.ListMeta
		*out = new(v1.ListMeta)
		(*in).DeepCopyInto(*out)
	}
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]*Task, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Task)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskList.
func (in *TaskList) DeepCopy() *TaskList {
	if in == nil {
		return nil
	}
	out := new(TaskList)
	in.DeepCopyInto(out)
	return out
}

// GetObjectKind returns the runtime ObjectKind.
func (in *TaskList) GetObjectKind() schema.ObjectKind {
	return in.TypeMeta.GetObjectKind()
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TaskList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TaskSpec) DeepCopyInto(out *TaskSpec) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.Init != nil {
		in, out := &in.Init, &out.Init
		*out = make([]*StepTemplate, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepTemplate)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Tasklet != nil {
		in, out := &in.Tasklet, &out.Tasklet
		*out = new(TaskletTemplate)
		(*in).DeepCopyInto(*out)
	}
	if in.PreDelete != nil {
		in, out := &in.PreDelete, &out.PreDelete
		*out = make([]*StepTemplate, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepTemplate)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskSpec.
func (in *TaskSpec) DeepCopy() *TaskSpec {
	if in == nil {
		return nil
	}
	out := new(TaskSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TaskStatus) DeepCopyInto(out *TaskStatus) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]*TaskCondition, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(TaskCondition)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Init != nil {
		in, out := &in.Init, &out.Init
		*out = make([]*StepStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepStatus)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.TaskletRef != nil {
		in, out := &in.TaskletRef, &out.TaskletRef
		*out = new(corev1.ObjectReference)
		**out = **in
	}
	if in.PreDelete != nil {
		in, out := &in.PreDelete, &out.PreDelete
		*out = make([]*StepStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepStatus)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskStatus.
func (in *TaskStatus) DeepCopy() *TaskStatus {
	if in == nil {
		return nil
	}
	out := new(TaskStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Tasklet) DeepCopyInto(out *Tasklet) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.TypeMeta != nil {
		in, out := &in.TypeMeta, &out.TypeMeta
		*out = new(v1.TypeMeta)
		**out = **in
	}
	if in.ObjectMeta != nil {
		in, out := &in.ObjectMeta, &out.ObjectMeta
		*out = new(v1.ObjectMeta)
		(*in).DeepCopyInto(*out)
	}
	if in.Spec != nil {
		in, out := &in.Spec, &out.Spec
		*out = new(TaskletSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Status != nil {
		in, out := &in.Status, &out.Status
		*out = new(TaskletStatus)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Tasklet.
func (in *Tasklet) DeepCopy() *Tasklet {
	if in == nil {
		return nil
	}
	out := new(Tasklet)
	in.DeepCopyInto(out)
	return out
}

// GetObjectKind returns the runtime ObjectKind.
func (in *Tasklet) GetObjectKind() schema.ObjectKind {
	return in.TypeMeta.GetObjectKind()
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Tasklet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TaskletCondition) DeepCopyInto(out *TaskletCondition) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.LastUpdateTime != nil {
		in, out := &in.LastUpdateTime, &out.LastUpdateTime
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskletCondition.
func (in *TaskletCondition) DeepCopy() *TaskletCondition {
	if in == nil {
		return nil
	}
	out := new(TaskletCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TaskletList) DeepCopyInto(out *TaskletList) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.TypeMeta != nil {
		in, out := &in.TypeMeta, &out.TypeMeta
		*out = new(v1.TypeMeta)
		**out = **in
	}
	if in.ListMeta != nil {
		in, out := &in.ListMeta, &out.ListMeta
		*out = new(v1.ListMeta)
		(*in).DeepCopyInto(*out)
	}
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]*Tasklet, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Tasklet)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskletList.
func (in *TaskletList) DeepCopy() *TaskletList {
	if in == nil {
		return nil
	}
	out := new(TaskletList)
	in.DeepCopyInto(out)
	return out
}

// GetObjectKind returns the runtime ObjectKind.
func (in *TaskletList) GetObjectKind() schema.ObjectKind {
	return in.TypeMeta.GetObjectKind()
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TaskletList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TaskletSpec) DeepCopyInto(out *TaskletSpec) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.Init != nil {
		in, out := &in.Init, &out.Init
		*out = make([]*StepTemplate, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepTemplate)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Install != nil {
		in, out := &in.Install, &out.Install
		*out = make([]*StepTemplate, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepTemplate)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.PostInstall != nil {
		in, out := &in.PostInstall, &out.PostInstall
		*out = make([]*StepTemplate, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepTemplate)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.PreDelete != nil {
		in, out := &in.PreDelete, &out.PreDelete
		*out = make([]*StepTemplate, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepTemplate)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskletSpec.
func (in *TaskletSpec) DeepCopy() *TaskletSpec {
	if in == nil {
		return nil
	}
	out := new(TaskletSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TaskletStatus) DeepCopyInto(out *TaskletStatus) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]*TaskletCondition, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(TaskletCondition)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Init != nil {
		in, out := &in.Init, &out.Init
		*out = make([]*StepStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepStatus)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Install != nil {
		in, out := &in.Install, &out.Install
		*out = make([]*StepStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepStatus)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.PostInstall != nil {
		in, out := &in.PostInstall, &out.PostInstall
		*out = make([]*StepStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepStatus)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.PreDelete != nil {
		in, out := &in.PreDelete, &out.PreDelete
		*out = make([]*StepStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StepStatus)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskletStatus.
func (in *TaskletStatus) DeepCopy() *TaskletStatus {
	if in == nil {
		return nil
	}
	out := new(TaskletStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TaskletTemplate) DeepCopyInto(out *TaskletTemplate) {
	*out = *in
	in.state = out.state
	if in.unknownFields != nil {
		in, out := &in.unknownFields, &out.unknownFields
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.ObjectMeta != nil {
		in, out := &in.ObjectMeta, &out.ObjectMeta
		*out = new(v1.ObjectMeta)
		(*in).DeepCopyInto(*out)
	}
	if in.Spec != nil {
		in, out := &in.Spec, &out.Spec
		*out = new(TaskletSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskletTemplate.
func (in *TaskletTemplate) DeepCopy() *TaskletTemplate {
	if in == nil {
		return nil
	}
	out := new(TaskletTemplate)
	in.DeepCopyInto(out)
	return out
}
