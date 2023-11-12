package deploy

import (
	"fmt"

	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/scmmishra/slick-deploy/internal/docker"
)

func Deploy(cfg config.DeploymentConfig) (int, error) {
	fmt.Println("Deploying...")

	err := docker.PullImage(cfg.Deployment.ImageName)
	if err != nil {
		fmt.Println("Failed to pull image")
		return 0, err
	}

	_, port, err := docker.RunContainer(cfg.Deployment.ImageName, cfg)
	if err != nil {
		fmt.Println("Failed to run container")
		return 0, err
	}

	fmt.Println("Deployed successfully")
	return port, nil
}
