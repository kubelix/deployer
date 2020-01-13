package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	// Name        string   `json:"name"`
	ProjectName string   `json:"projectName"`
	Singleton   bool     `json:"singleton"`
	Image       string   `json:"image"`
	Command     []string `json:"command,omitempty"`
	Args        []string `json:"args,omitempty"`

	Ports     PortList                      `json:"ports,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	Env       Environment                 `json:"env,omitempty"`
	Files     []File                      `json:"files,omitempty"`
}

// Environment defines env vars for the app container
type Environment map[string]string

// PortList holds a list of ports
type PortList []Port

// Port defines a port the app opens
type Port struct {
	Name      string       `json:"name"`
	Container uint16       `json:"container"`
	Ingress   *PortIngress `json:"ingress,omitempty"`
}

// PortIngress defines the ingress config for a port
type PortIngress struct {
	Paths []string `json:"paths"`
	Hosts []string `json:"hosts"`
}

// File defines a file the app needs
type File struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// ServiceStatus defines the observed state of Service
type ServiceStatus struct {}

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
