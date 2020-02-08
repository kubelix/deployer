package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	ConfigFile string `split_words:"true" default:"config.yaml"`
}

// Env holds the environment config
var Env envConfig

// Config file content
var Config RootConfig

func Init() {
	envconfig.MustProcess("", &Env)

	cfg, err := Load(Env.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	Config = *cfg
}
