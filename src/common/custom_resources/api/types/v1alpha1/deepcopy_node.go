package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func (in *MayastorNode) DeepCopyInto(out *MayastorNode) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
}

func (in *MayastorNode) DeepCopy() *MayastorNode {
	if in == nil {
		return nil
	}
	out := new(MayastorNode)
	in.DeepCopyInto(out)
	return out
}

func (in *MayastorNode) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *MayastorNodeList) DeepCopyInto(out *MayastorNodeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MayastorNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *MayastorNodeList) DeepCopy() *MayastorNodeList {
	if in == nil {
		return nil
	}
	out := new(MayastorNodeList)
	in.DeepCopyInto(out)
	return out
}

func (in *MayastorNodeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
