package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Name         string         `json:"name"`
	ReverseProxy []ReverseProxy `json:"reverse_proxy"`
}

type ReverseProxy struct {
	Path string `json:"path"`
	To   string `json:"to"`
}

type DeploymentConfig struct {
	Deployment struct {
		ImageName     string `yaml:"image_name"`
		ContainerPort int    `yaml:"container_port"`
		PortIncrement int    `yaml:"port_increment"`
		PortRange     struct {
			Start int `yaml:"start"`
			End   int `yaml:"end"`
		} `yaml:"port_range"`
	} `yaml:"deployment"`
	Caddy struct {
		AdminAPI string   `yaml:"admin_api"`
		Servers  []Server `json:"servers"`
	} `yaml:"caddy"`
	HealthCheck struct {
		Endpoint       string `yaml:"endpoint"`
		TimeoutSeconds int    `yaml:"timeout_seconds"`
	} `yaml:"health_check"`
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

	if config.Deployment.PortIncrement == 0 {
		config.Deployment.PortIncrement = 1
	}

	if config.Deployment.PortRange.Start == 0 {
		config.Deployment.PortRange.Start = 8000
	}

	if config.Deployment.PortRange.End == 0 {
		config.Deployment.PortRange.End = 9999
	}

	if config.Caddy.AdminAPI == "" {
		config.Caddy.AdminAPI = "http://localhost:2019"
	}

	if config.HealthCheck.Endpoint == "" {
		config.HealthCheck.Endpoint = "/"
	}

	return config, nil
}
