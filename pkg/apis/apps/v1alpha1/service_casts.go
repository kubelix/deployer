package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ToEnvVars casts this simple map to a slice of k8s env var fields
// noinspection GoReceiverNames
func (e Environment) ToEnvVars() []corev1.EnvVar {
	envs := make([]corev1.EnvVar, 0)
	for k, v := range e {
		envs = append(envs, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return envs
}

// ToPodPorts casts the PortList to a slice of container ports usable for a pod spec
// noinspection GoReceiverNames
func (p PortList) ToPodPorts() []corev1.ContainerPort {
	ports := make([]corev1.ContainerPort, 0)
	for _, port := range p {
		ports = append(ports, corev1.ContainerPort{
			Name:          port.Name,
			ContainerPort: int32(port.Container),
		})
	}
	return ports
}

// ToServicePorts casts the PortList to a slice of container ports usable for a pod spec
// noinspection GoReceiverNames
func (p PortList) ToServicePorts() []corev1.ServicePort {
	ports := make([]corev1.ServicePort, 0)
	for _, port := range p {
		ports = append(ports, corev1.ServicePort{
			Port:       int32(port.Container),
			Name:       port.Name,
			TargetPort: intstr.FromString(port.Name),
		})
	}
	return ports
}
