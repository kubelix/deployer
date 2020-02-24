package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	Singleton bool     `json:"singleton"`
	Image     string   `json:"image"`
	Command   []string `json:"command,omitempty"`
	Args      []string `json:"args,omitempty"`

	Ports              PortList                    `json:"ports,omitempty"`
	Resources          corev1.ResourceRequirements `json:"resources,omitempty"`
	Env                Environment                 `json:"env,omitempty"`
	Files              []File                      `json:"files,omitempty"`
	ServiceAccountName string                      `json:"serviceAccountName,omitempty"`
}

// Environment defines env vars for the app container
type Environment map[string]string

// PortList holds a list of ports
type PortList []Port

// Port defines a port the app opens
type Port struct {
	Name      string        `json:"name"`
	Container uint16        `json:"container"`
	Service   uint16        `json:"service,omitempty"`
	Ingresses []PortIngress `json:"ingresses,omitempty"`
}

// PortIngress defines the ingress config for a port
type PortIngress struct {
	Host string `json:"host"`

	// # +kubebuilder:default={/}
	Paths []string `json:"paths,omitempty"`
}

// File defines a file the app needs
type File struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

// ServiceStatus defines the observed state of Service
type ServiceStatus struct {
	ManagedObjects ManagedObjectList `json:"managedObjects,omitempty"`
}

// ManagedObjectList is a list type for ManagedObject with utility functions
type ManagedObjectList []*ManagedObject

func (in *ManagedObjectList) FromObjectList(objects []runtime.Object) {
	for _, obj := range objects {
		meta, ok := obj.(metav1.Object)
		if !ok {
			panic(fmt.Errorf("failed to convert %s to metav1.Object", obj))
		}

		name := types.NamespacedName{
			Namespace: meta.GetNamespace(),
			Name:      meta.GetName(),
		}

		in.Add(obj, name, "")
	}
}

// ManagedObject references an object
type ManagedObject struct {
	Checksum  string                 `json:"checksum"`
	Reference corev1.ObjectReference `json:"reference"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Service is the Schema for the services API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=services,scope=Namespaced
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceList contains a list of Service
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Service{}, &ServiceList{})
}
