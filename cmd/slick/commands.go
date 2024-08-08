package main

import (
	"fmt"

	"github.com/scmmishra/slick-deploy/internal/caddy"
	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/scmmishra/slick-deploy/internal/deploy"
	"github.com/spf13/cobra"
)

type Deployer interface {
	Deploy(cfg config.DeploymentConfig) error
}
type DefaultDeployer struct{}

func (d DefaultDeployer) Deploy(cfg config.DeploymentConfig) error {
	return deploy.Deploy(cfg)
}

var defaultDeployer Deployer = DefaultDeployer{}

func runDeploy(cmd *cobra.Command, deployer Deployer, configLoader ConfigLoader) error {
	cfg, err := configLoader(cmd)
	if err != nil {
		return err
	}
	return deployer.Deploy(cfg)
}

func runStatus() error {
	dockerService, err := dockerServiceCreator()
	if err != nil {
		return err
	}
	return dockerService.GetStatus()
}

func runLogs(cmd *cobra.Command, configLoader ConfigLoader) error {
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

func runCaddyInspect(cmd *cobra.Command, configLoader ConfigLoader) error {
	cfg, err := configLoader(cmd)
	if err != nil {
		return err
	}

	caddyConfig := caddy.ConvertToCaddyfile(cfg.Caddy, 0) // Use 0 as port since we're just inspecting
	fmt.Println(caddyConfig)
	return nil
}
