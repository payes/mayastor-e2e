package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func (in *MayastorPool) DeepCopyInto(out *MayastorPool) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
}

func (in *MayastorPool) DeepCopy() *MayastorPool {
	if in == nil {
		return nil
	}
	out := new(MayastorPool)
	in.DeepCopyInto(out)
	return out
}

func (in *MayastorPool) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *MayastorPoolList) DeepCopyInto(out *MayastorPoolList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MayastorPool, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *MayastorPoolList) DeepCopy() *MayastorPoolList {
	if in == nil {
		return nil
	}
	out := new(MayastorPoolList)
	in.DeepCopyInto(out)
	return out
}

func (in *MayastorPoolList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
