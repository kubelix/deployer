package config

func NewConfig() *RootConfig {
	return &RootConfig{
		CoreService: CoreServiceConfig{
			Annotations: map[string]string{},
		},
		Deployment: DeploymentConfig{
			Annotations: map[string]string{},
		},
		Ingress: IngressConfig{
			Annotations: map[string]string{},
		},
		DockerPullSecretes: []DockerPullSecret{},
	}
}

// RootConfig configures the behavior of the operator
type RootConfig struct {
	CoreService        CoreServiceConfig  `json:"coreService"`
	Deployment         DeploymentConfig   `json:"deployment"`
	Ingress            IngressConfig      `json:"ingress"`
	DockerPullSecretes []DockerPullSecret `json:"dockerPullSecretes"`
}

// IngressConfig specifies additional information for ingress creation
type IngressConfig struct {
	Annotations map[string]string `json:"annotations"`
}

// DeploymentConfig specifies additional information for deployment creation
type DeploymentConfig struct {
	Annotations map[string]string `json:"annotations"`
}

// CoreServiceConfig specifies additional information for core/v1 Service creation
type CoreServiceConfig struct {
	Annotations map[string]string `json:"annotations"`
}

// DockerPullSecret defines secrets used to pull docker images from a registry
type DockerPullSecret struct {
	Registry string `json:"registry"`
	Username string `json:"username"`
	Password string `json:"password"`
}
