package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/scmmishra/slick-deploy/internal/caddy"
	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/scmmishra/slick-deploy/internal/deploy"
	"github.com/scmmishra/slick-deploy/internal/docker"
	"github.com/spf13/cobra"
)

type Deployer interface {
	Deploy(cfg config.DeploymentConfig) error
}

type DockerService interface {
	GetStatus() error
	FindContainer(imageName string) *docker.Container
	StreamLogs(containerID string, tail string) error
}

type DockerServiceCreator func() (DockerService, error)

var dockerServiceCreator DockerServiceCreator = func() (DockerService, error) {
	return newDockerService(docker.NewDockerClient)
}

type DefaultDeployer struct{}

func (d DefaultDeployer) Deploy(cfg config.DeploymentConfig) error {
	return deploy.Deploy(cfg)
}

var defaultDeployer Deployer = DefaultDeployer{}

type ConfigLoader func(*cobra.Command) (config.DeploymentConfig, error)

func runDeploy(cmd *cobra.Command, deployer Deployer, configLoader ConfigLoader) error {
	cfg, err := configLoader(cmd)
	if err != nil {
		return err
	}
	return deployer.Deploy(cfg)
}

func runStatus(cmd *cobra.Command, args []string) error {
	dockerService, err := dockerServiceCreator()
	if err != nil {
		return err
	}
	return dockerService.GetStatus()
}

func runLogs(cmd *cobra.Command, args []string, configLoader ConfigLoader) error {
	cfg, err := configLoader(cmd)
	if err != nil {
		return err
	}

	dockerService, err := dockerServiceCreator()
	if err != nil {
		return err
	}

	container := dockerService.FindContainer(cfg.App.ImageName)
	if container == nil {
		return fmt.Errorf("no container found")
	}

	tail, _ := cmd.Flags().GetString("tail")
	return dockerService.StreamLogs(container.ID, tail)
}

func runCaddyInspect(cmd *cobra.Command, args []string, configLoader ConfigLoader) error {
	cfg, err := configLoader(cmd)
	if err != nil {
		return err
	}

	caddyConfig := caddy.ConvertToCaddyfile(cfg.Caddy, 0) // Use 0 as port since we're just inspecting
	fmt.Println(caddyConfig)
	return nil
}

func defaultConfigLoader(cmd *cobra.Command) (config.DeploymentConfig, error) {
	cfgPath, _ := cmd.Flags().GetString("config")
	envPath, _ := cmd.Flags().GetString("env")

	if err := godotenv.Load(envPath); err != nil {
		return config.DeploymentConfig{}, fmt.Errorf("failed to load env file: %w", err)
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return config.DeploymentConfig{}, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}

type DockerClientCreator func() (docker.DockerClient, error)

func newDockerService(clientCreator DockerClientCreator) (DockerService, error) {
	cli, err := clientCreator()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return docker.NewDockerService(cli), nil
}
