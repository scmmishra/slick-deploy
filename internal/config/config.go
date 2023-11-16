package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type PortRange struct {
	Start int `yaml:"start"`
	End   int `yaml:"end"`
}

type App struct {
	Name          string    `yaml:"name"`
	ImageName     string    `yaml:"image_name"`
	ContainerPort int       `yaml:"container_port"`
	ENV           []string  `yaml:"env"`
	PortRange     PortRange `yaml:"port_range"`
}

type ReverseProxy struct {
	Path string `yaml:"path"`
	To   string `yaml:"to"`
}

type Rule struct {
	Match        string         `yaml:"match"`
	ReverseProxy []ReverseProxy `yaml:"reverse_proxy"`
}

type CaddyConfig struct {
	AdminAPI string `yaml:"admin_api"`
	Rules    []Rule `yaml:"rules"`
}

type HealthCheck struct {
	Endpoint       string `yaml:"endpoint"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

type DeploymentConfig struct {
	App         App         `yaml:"app"`
	Caddy       CaddyConfig `yaml:"caddy"`
	HealthCheck HealthCheck `yaml:"health_check"`
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
	yamlData, _ := io.ReadAll(yamlFile)

	// Unmarshal the YAML into the config struct
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		return config, err
	}

	if config.App.PortRange.Start == 0 {
		config.App.PortRange.Start = 8000
	}

	if config.App.PortRange.End == 0 {
		config.App.PortRange.End = 9000
	}

	if config.Caddy.AdminAPI == "" {
		config.Caddy.AdminAPI = "http://localhost:2019"
	}

	return config, nil
}
