package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/registry"
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
func PullImage(imageName string, registryConfig config.RegistryConfig) error {
	// A context in Go is used to define a deadline or a cancellation signal
	// for requests made to external resources, like a Docker Daemon in this case.
	ctx := context.Background()

	// Initialize a new Docker client. It automatically negotiates the API version
	cli, err := NewDockerClient()
	if err != nil {
		return err
	}

	authConfig := registry.AuthConfig{
		Username: registryConfig.Username,
		Password: registryConfig.Password,
	}
	encodedJSON, err := json.Marshal(authConfig)

	if err != nil {
		return err
	}

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	options := types.ImagePullOptions{
		RegistryAuth: authStr,
	}

	out, err := cli.ImagePull(ctx, imageName, options)
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

type Container struct {
	ID   string
	Port int
}

func RunContainer(imageName string, cfg config.DeploymentConfig) (*Container, error) {
	ctx := context.Background()

	cli, err := NewDockerClient()
	if err != nil {
		return nil, err
	}

	portManager := utils.NewPortManager(cfg.App.PortRange.Start, cfg.App.PortRange.End, 1)
	port, err := portManager.AllocatePort()

	if err != nil {
		return nil, err
	}

	envs := []string{}

	for _, env := range cfg.App.ENV {
		envValue, exists := os.LookupEnv(env)
		if exists {
			envs = append(envs, env+"="+envValue)
		}
	}

	containerConfig := &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", cfg.App.ContainerPort)): struct{}{},
		},
		Env: envs,
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

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, err
	}

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("error allocating port: %w", err)
	}

	return &Container{
		ID:   resp.ID,
		Port: port,
	}, nil
}

func FindContainer(imageName string) *Container {
	ctx := context.Background()

	cli, err := NewDockerClient()

	if err != nil {
		return nil
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil
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
			return &Container{
				ID: container.ID,
			}
		}
	}

	return nil
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
