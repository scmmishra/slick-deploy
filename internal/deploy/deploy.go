package deploy

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/scmmishra/slick-deploy/internal/caddy"
	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/scmmishra/slick-deploy/internal/docker"
	"github.com/scmmishra/slick-deploy/internal/health"
)

func Deploy(cfg config.DeploymentConfig) error {
	fmt.Println("Deploying...")

	// Initialize Docker client
	cli, err := docker.NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Create DockerService instance
	dockerService := docker.NewDockerService(cli)

	err = dockerService.PullImage(cfg.App.ImageName, cfg.App.Registry)
	if err != nil {
		return err
	}

	fmt.Println("- Looking for existing container")
	oldContainer := dockerService.FindContainer(cfg.App.ImageName)

	fmt.Println("- Spinning up new container")
	newContainer, err := dockerService.RunContainer(cfg.App.ImageName, cfg.App)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure resources are freed on exit

	go handleSignals(ctx, cancel, dockerService, newContainer.ID)

	fmt.Println("- Waiting for container to be healthy")
	host := fmt.Sprintf("http://localhost:%d", newContainer.Port)
	err = health.CheckHealth(host, &cfg.HealthCheck)
	if err != nil {
		fmt.Println("Container is unhealthy, rolling back")
		dockerService.StopContainer(newContainer.ID)
		return err
	}

	fmt.Println("- Setting up caddy")
	err = caddy.SetupCaddy(newContainer.Port, cfg)
	if err != nil {
		fmt.Println("Unable to setup caddy, rolling back")
		dockerService.StopContainer(newContainer.ID)
		return err
	}

	if oldContainer != nil {
		fmt.Println("- Killing old container")
		dockerService.StopContainer(oldContainer.ID)
	}

	fmt.Println("Deployed successfully")
	return nil
}

func handleSignals(ctx context.Context, cancelFunc context.CancelFunc, dockerService *docker.DockerService, newContainerID string) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
		fmt.Println("Received interrupt signal, rolling back")
		dockerService.StopContainer(newContainerID)
	case <-ctx.Done():
		// Context cancelled, stop listening for signals
	}

	// Clean up
	signal.Stop(sigs)
	close(sigs)
}
