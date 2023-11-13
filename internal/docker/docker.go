package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/scmmishra/slick-deploy/pkg/utils"
)

type ImagePullResponse struct {
	Status         string `json:"status"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
	Progress string `json:"progress"`
	ID       string `json:"id"`
}

func NewDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return cli, nil
}

// PullImage is a function that pulls a Docker image from a Docker registry.
// This is similar to running `docker pull <image>` from the command line.
func PullImage(imageName string) error {
	// A context in Go is used to define a deadline or a cancellation signal
	// for requests made to external resources, like a Docker Daemon in this case.
	ctx := context.Background()

	// Initialize a new Docker client. It automatically negotiates the API version
	cli, err := NewDockerClient()
	if err != nil {
		return err
	}

	fmt.Println("Pulling image...")

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	// Ensure the response body is closed after this function ends.
	// This is important for resource management and to prevent memory leaks.
	defer out.Close()

	// Process the output from ImagePull to show progress.
	dec := json.NewDecoder(out)
	var response ImagePullResponse
	for {
		if err := dec.Decode(&response); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if response.Progress != "" {
			fmt.Printf("\r%-70s", "Progress: "+response.Progress)
		} else {
			fmt.Printf("\r%-70s", response.Status)
		}
		os.Stdout.Sync()
	}

	fmt.Println() // Print a new line at the end
	// If everything goes well, return nil indicating the pull was successful.
	return nil
}

func RunContainer(imageName string, cfg config.DeploymentConfig) (string, int, error) {
	ctx := context.Background()

	cli, err := NewDockerClient()
	if err != nil {
		return "", 0, err
	}

	containerId, _ := FindContainer(cli, imageName)

	portManager := utils.NewPortManager(cfg.App.PortRange.Start, cfg.App.PortRange.End, 1)
	port, err := portManager.AllocatePort()

	if err != nil {
		return "", 0, err
	}

	containerConfig := &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", cfg.App.ContainerPort)): struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", cfg.App.ContainerPort)): []nat.PortBinding{
				{
					HostIP:   "127.0.0.1",
					HostPort: fmt.Sprintf("%d", port),
				},
			},
		},
	}

	fmt.Printf("Starting new container on port %d\n", port)
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", 0, err
	}

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", 0, fmt.Errorf("error allocating port: %w", err)
	}

	// assume health check worked
	fmt.Println("Container started")

	if containerId != "" {
		fmt.Printf("Stopping existing container with ID: %s\n", containerId)
		StopContainer(containerId)
	}

	return resp.ID, port, nil
}

func FindContainer(cli *client.Client, imageName string) (string, error) {
	ctx := context.Background()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return "", err
	}

	baseImageName := strings.Split(imageName, ":")[0]

	for _, container := range containers {
		// Inspect each container to get detailed information
		cont, err := cli.ContainerInspect(ctx, container.ID)
		if err != nil {
			continue // Skip to next container on error
		}

		containerBaseImageName := strings.Split(cont.Config.Image, ":")[0]
		// Check if the container's image matches the specified image name
		if containerBaseImageName == baseImageName {
			return container.ID, nil
		}
	}

	return "", nil
}

func StopContainer(containerID string) error {
	ctx := context.Background()
	cli, err := NewDockerClient()
	if err != nil {
		return err
	}

	defer cli.Close()

	timeout := 15
	stopOptions := container.StopOptions{
		Timeout: &timeout,
	}

	err = cli.ContainerStop(ctx, containerID, stopOptions)
	if err != nil {
		return err
	}

	return nil
}
