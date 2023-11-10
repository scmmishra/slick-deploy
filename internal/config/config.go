package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type DeploymentConfig struct {
	Deployment struct {
		ImageName         string `yaml:"image_name"`
		ContainerBaseName string `yaml:"container_base_name"`
	} `yaml:"deployment"`
	Caddy struct {
		AdminAPI     string `yaml:"admin_api"`
		ReverseProxy struct {
			ProxyMatcher string   `yaml:"proxy_matcher"`
			ToPort       int      `yaml:"to_port"`
			Rewrite      []string `yaml:"rewrite"`
		} `yaml:"reverse_proxy"`
	} `yaml:"caddy"`
	HealthCheck struct {
		Endpoint       string `yaml:"endpoint"`
		TimeoutSeconds int    `yaml:"timeout_seconds"`
	} `yaml:"health_check"`
	Network struct {
		StartPort     int `yaml:"start_port"`
		PortIncrement int `yaml:"port_increment"`
	} `yaml:"network"`
}

func LoadConfig(path string) (DeploymentConfig, error) {
	var config DeploymentConfig
	yamlFile, err := os.Open(path)
	if err != nil {
		return config, fmt.Errorf("error opening config file: %v", err)
	}

	// close the file when we're done
	defer yamlFile.Close()

	// Read the file content
	yamlData, err := io.ReadAll(yamlFile)
	if err != nil {
		return config, err
	}

	// Unmarshal the YAML into the config struct
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
