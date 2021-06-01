package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func (in *MayastorVolume) DeepCopyInto(out *MayastorVolume) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
}

func (in *MayastorVolume) DeepCopy() *MayastorVolume {
	if in == nil {
		return nil
	}
	out := new(MayastorVolume)
	in.DeepCopyInto(out)
	return out
}

func (in *MayastorVolume) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *MayastorVolumeList) DeepCopyInto(out *MayastorVolumeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MayastorVolume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *MayastorVolumeList) DeepCopy() *MayastorVolumeList {
	if in == nil {
		return nil
	}
	out := new(MayastorVolumeList)
	in.DeepCopyInto(out)
	return out
}

func (in *MayastorVolumeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
