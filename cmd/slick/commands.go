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

func runDeploy(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}
	return deploy.Deploy(cfg)
}

func runStatus(cmd *cobra.Command, args []string) error {
	dockerService, err := newDockerService()
	if err != nil {
		return err
	}
	return dockerService.GetStatus()
}

func runLogs(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	dockerService, err := newDockerService()
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

func runCaddyInspect(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	caddyConfig := caddy.ConvertToCaddyfile(cfg.Caddy, 0) // Use 0 as port since we're just inspecting
	fmt.Println(caddyConfig)
	return nil
}

func loadConfig(cmd *cobra.Command) (config.DeploymentConfig, error) {
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

func newDockerService() (*docker.DockerService, error) {
	cli, err := docker.NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return docker.NewDockerService(cli), nil
}
