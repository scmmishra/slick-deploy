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
	cfgPath, _ := cmd.Flags().GetString("config")
	envPath, _ := cmd.Flags().GetString("env")

	if err := godotenv.Load(envPath); err != nil {
		return fmt.Errorf("failed to load env file: %w", err)
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return deploy.Deploy(cfg)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cli, err := docker.NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	dockerService := docker.NewDockerService(cli)
	return dockerService.GetStatus()
}

func runLogs(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")
	tail, _ := cmd.Flags().GetString("tail")

	cli, err := docker.NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	dockerService := docker.NewDockerService(cli)

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	container := dockerService.FindContainer(cfg.App.ImageName)
	if container == nil {
		return fmt.Errorf("no container found")
	}

	return dockerService.StreamLogs(container.ID, tail)
}

func runCaddyInspect(cmd *cobra.Command, args []string) error {
	cfgPath, _ := cmd.Flags().GetString("config")

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	caddyConfig := caddy.ConvertToCaddyfile(cfg.Caddy, 0) // Use 0 as port since we're just inspecting
	fmt.Println(caddyConfig)
	return nil
}
