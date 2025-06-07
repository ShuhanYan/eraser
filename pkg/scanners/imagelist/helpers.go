package main

import (
	"os"

	"k8s.io/apimachinery/pkg/util/yaml"
)

func loadConfig(filename string) (Config, error) {
	cfg := *DefaultConfig()

	b, err := os.ReadFile(filename)
	if err != nil {
		log.Error(err, "unable to read image list scanner config")
		return cfg, err
	}

	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		log.Error(err, "unable to unmarshal image list scanner config")
	}

	return cfg, nil
}
