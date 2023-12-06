package config

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type PortRange struct {
	Start int `yaml:"start"`
	End   int `yaml:"end"`
}

type RegistryConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type App struct {
	Name          string         `yaml:"name"`
	ImageName     string         `yaml:"image"`
	Registry      RegistryConfig `yaml:"registry"`
	ContainerPort int            `yaml:"container_port"`
	Network       string         `yaml:"network"`
	ENV           []string       `yaml:"env"`
	PortRange     PortRange      `yaml:"port_range"`
}

type ReverseProxy struct {
	Path string `yaml:"path"`
	To   string `yaml:"to"`
}

type Rule struct {
	Match        string         `yaml:"match"`
	Tls          string         `yaml:"tls"`
	ReverseProxy []ReverseProxy `yaml:"reverse_proxy"`
}

type CaddyConfig struct {
	AdminAPI string `yaml:"admin_api"`
	Rules    []Rule `yaml:"rules"`
}

type HealthCheck struct {
	Endpoint        string `yaml:"endpoint"`
	TimeoutSeconds  int    `yaml:"timeout_seconds"`
	IntervalSeconds int    `yaml:"interval_seconds"`
	MaxRetries      int    `yaml:"max_retries"`
}

type DeploymentConfig struct {
	App         App         `yaml:"app"`
	Caddy       CaddyConfig `yaml:"caddy"`
	HealthCheck HealthCheck `yaml:"health_check"`
}

func replaceEnvVariables(input string) (string, error) {
	re := regexp.MustCompile(`\{env\.([a-zA-Z_][a-zA-Z0-9_]*)\}`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract the variable name
		varName := strings.TrimPrefix(match, "{env.")
		varName = strings.TrimSuffix(varName, "}")

		// Get the environment variable value
		return os.Getenv(varName)
	}), nil
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

	if config.HealthCheck.TimeoutSeconds == 0 {
		config.HealthCheck.TimeoutSeconds = 5
	}

	if config.HealthCheck.IntervalSeconds == 0 {
		config.HealthCheck.IntervalSeconds = 5
	}

	if config.HealthCheck.MaxRetries == 0 {
		config.HealthCheck.MaxRetries = 3
	}

	for _, rule := range config.Caddy.Rules {
		newTlsValue, err := replaceEnvVariables(rule.Tls)
		if err != nil {
			return config, err
		}

		rule.Tls = newTlsValue
	}

	if config.App.Registry.Username != "" && config.App.Registry.Password != "" {
		envValue, exists := os.LookupEnv(config.App.Registry.Password)
		if exists {
			config.App.Registry.Password = envValue
		}
	}

	return config, nil
}
