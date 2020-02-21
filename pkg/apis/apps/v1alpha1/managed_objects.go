package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// ReferenceEquals returns true when both ManagedObjects reference the same object
func (in *ManagedObject) ReferenceEquals(obj *ManagedObject) bool {
	return in.Reference.Kind == obj.Reference.Kind &&
		in.Reference.APIVersion == obj.Reference.APIVersion &&
		in.Reference.Name == obj.Reference.Name &&
		in.Reference.Namespace == obj.Reference.Namespace
}

// GroupVersionKind returns the GroupVersionKind of the underlying ObjectReference
func (in *ManagedObject) GroupVersionKind() schema.GroupVersionKind {
	return in.Reference.GroupVersionKind()
}

// NamespacedName returns the NamespacedName representation of the ManagedObject
func (in *ManagedObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: in.Reference.Namespace,
		Name:      in.Reference.Name,
	}
}

// String representation of the managed object, mainly for logging purposes
func (in *ManagedObject) String() string {
	return fmt.Sprintf("%s/%s/%s/%s", in.Reference.APIVersion, in.Reference.Kind, in.Reference.Namespace, in.Reference.Name)
}

// Find a ManagedObject by name and type
func (in *ManagedObjectList) Find(obj runtime.Object, name types.NamespacedName) *ManagedObject {
	for _, m := range *in {
		if referenceIsName(m.Reference, name) && referenceIsObject(m.Reference, obj) {
			return m
		}
	}

	return nil
}

// Contains returns true when the given ManagedObject is part of this list
func (in *ManagedObjectList) Contains(obj *ManagedObject) bool {
	for _, m := range *in {
		if m.ReferenceEquals(obj) {
			return true
		}
	}

	return false
}

// Add a new ManagedObject with the attributes from given object, name and checksum
func (in *ManagedObjectList) Add(obj runtime.Object, name types.NamespacedName, checksum string) *ManagedObject {
	apiVersion, kind := obj.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()

	managedObject := &ManagedObject{
		Checksum: checksum,
		Reference: corev1.ObjectReference{
			APIVersion: apiVersion,
			Kind:       kind,
			Name:       name.Name,
			Namespace:  name.Namespace,
		},
	}

	*in = append(*in, managedObject)

	return managedObject
}

// Remove the ManagedObject from the list
func (in *ManagedObjectList) Remove(obj *ManagedObject) {
	for index, m := range *in {
		if m.ReferenceEquals(obj) {
			in.removeIndex(index)
			return
		}
	}
}

func (in *ManagedObjectList) removeIndex(index int) {
	*in = append((*in)[:index], (*in)[index+1:]...)
}

func referenceIsName(ref corev1.ObjectReference, name types.NamespacedName) bool {
	return ref.Namespace == name.Namespace && ref.Name == name.Name
}

func referenceIsObject(ref corev1.ObjectReference, obj runtime.Object) bool {
	refGVK := ref.GroupVersionKind()
	objGVK := obj.GetObjectKind().GroupVersionKind()

	return refGVK.Group == objGVK.Group &&
		refGVK.Version == objGVK.Version &&
		refGVK.Kind == objGVK.Kind
}
