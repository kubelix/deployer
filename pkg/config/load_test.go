package config

import (
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	cfg, err := Load("./test/config.yaml")
	if err != nil {
		t.Fatalf("Got error from Load(): %v", err)
	}

	if cfg == nil {
		t.Fatal("Got nil as config")
	}

	if !reflect.DeepEqual(cfg, testConfig) {
		t.Errorf("Expected to get %#v got %#v", testConfig, cfg)
	}
}

var testConfig = &RootConfig{
	CoreService: CoreServiceConfig{
		Annotations: map[string]string{},
	},
	Deployment: DeploymentConfig{
		Annotations: map[string]string{},
	},
	Ingress: IngressConfig{
		Annotations: map[string]string{
			"cert-manager.io/cluster-issuer": "letsencrypt",
			"kubernetes.io/ingress.class":    "nginx",
		},
	},
	DockerPullSecretes: []DockerPullSecret{
		{
			Registry: "gitlab.com",
			Username: "test",
			Password: "testpw",
		},
	},
}
