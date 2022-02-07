package owner

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func HasOwnerRef(owner metav1.Object, obj metav1.Object) bool {
	ownerRefs := obj.GetOwnerReferences()
	for i := 0; i < len(ownerRefs); i++ {
		if ownerRefs[i].UID == owner.GetUID() {
			return true
		}
	}
	return false
}

func HasGVKOwnerRef(obj metav1.Object, gvk schema.GroupVersionKind) bool {
	ownerRefs := obj.GetOwnerReferences()
	for i := 0; i < len(ownerRefs); i++ {
		if ownerRefs[i].APIVersion == gvk.GroupVersion().String() &&
			ownerRefs[i].Kind == gvk.Kind {
			return true
		}
	}
	return false
}

func DelOwnerRef(owner metav1.Object, obj metav1.Object) bool {
	// 没有引用返回
	if !HasOwnerRef(owner, obj) {
		return false
	}
	var newOwnerRefs []metav1.OwnerReference
	ownerRefs := obj.GetOwnerReferences()
	for i := 0; i < len(ownerRefs); i++ {
		if ownerRefs[i].UID == owner.GetUID() {
			continue
		}
		newOwnerRefs = append(newOwnerRefs, ownerRefs[i])
	}

	obj.SetOwnerReferences(newOwnerRefs)
	return true
}

func AddOwnerRef(owner metav1.Object, obj metav1.Object, gvk schema.GroupVersionKind) bool {
	if HasOwnerRef(owner, obj) {
		return false
	}

	isController := true
	BlockOwnerDeletion := true

	newOwnerRef := metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		Controller:         &isController,
		BlockOwnerDeletion: &BlockOwnerDeletion,
	}

	ownerRefs := obj.GetOwnerReferences()
	ownerRefs = append(ownerRefs, newOwnerRef)
	obj.SetOwnerReferences(ownerRefs)
	return true
}
