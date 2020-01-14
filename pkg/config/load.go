package config

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
)

// Load the config from a given file
func Load(path string) (*RootConfig, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	yb, err := yaml.JSONToYAML(b)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	cfg := NewConfig()
	if err := yaml.Unmarshal(yb, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config: %v", err)
	}

	return cfg, nil
}
