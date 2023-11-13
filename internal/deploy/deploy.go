package deploy

import (
	"fmt"

	"github.com/scmmishra/slick-deploy/internal/caddy"
	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/scmmishra/slick-deploy/internal/docker"
)

func Deploy(cfg config.DeploymentConfig) error {
	fmt.Println("Deploying...")

	err := docker.PullImage(cfg.App.ImageName)
	if err != nil {
		fmt.Println("Failed to pull image")
		return err
	}

	fmt.Println("- Looking for existing container")
	oldContainer := docker.FindContainer(cfg.App.ImageName)

	fmt.Println("- Spinning up new container")
	newContainer, err := docker.RunContainer(cfg.App.ImageName, cfg)
	if err != nil {
		return err
	}

	fmt.Println("- Setting up caddy")
	err = caddy.SetupCaddy(newContainer.Port, cfg)
	if err != nil {
		return err
	}

	if oldContainer != nil {
		fmt.Println("- Killing old container")
		docker.StopContainer(oldContainer.ID)
	}

	fmt.Println("Deployed successfully")
	return nil
}
