package cmd

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	PushGateway  `yaml:"push-gateway"`
	NodeExporter `yaml:"node-exporter"`
}

type NodeExporter struct {
	Namespace  string `yaml:"namespace"`
	ScrapeTime int    `yaml:"scrape-every"`
}

type PushGateway struct {
	Endpoint       string `yaml:"endpoint"`
	Authentication Auth   `yaml:"auth,omitempty"`
	Project        string `yaml:"project"`
}

type Auth struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

func ReadConfig(configPath string) (Config, error) {
	if configPath == "" {
		configPath = "config.yaml"
	}
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal YAML: %v", err)
	}
	return config, nil
}
