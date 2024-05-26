package docker

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
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

// DockerService holds the client used to interact with Docker.
type DockerService struct {
	Client DockerClient
}

// NewDockerService creates a new instance of DockerService with the given DockerClient.
func NewDockerService(cli DockerClient) *DockerService {
	return &DockerService{Client: cli}
}

type DockerClient interface {
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
	ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	Close() error
}

func NewDockerClient() (DockerClient, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// PullImage is a function that pulls a Docker image from a Docker registry.
// This is similar to running `docker pull <image>` from the command line.
func (ds *DockerService) PullImage(imageName string, registryConfig config.RegistryConfig) error {
	// A context in Go is used to define a deadline or a cancellation signal
	// for requests made to external resources, like a Docker Daemon in this case.
	ctx := context.Background()

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

	out, err := ds.Client.ImagePull(ctx, imageName, options)
	if err != nil {
		return err
	}

	// Ensure the response body is closed after this function ends.
	// This is important for resource management and to prevent memory leaks.
	// skipcq: GO-S2307
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

func (ds *DockerService) RunContainer(imageName string, appCfg config.App) (*Container, error) {
	ctx := context.Background()

	portManager := utils.NewPortManager(appCfg.PortRange.Start, appCfg.PortRange.End, 1)
	port, err := portManager.AllocatePort()

	if err != nil {
		return nil, err
	}

	// skipcq: GO-W1027
	envs := []string{}

	for _, env := range appCfg.ENV {
		envValue, exists := os.LookupEnv(env)
		if exists {
			envs = append(envs, env+"="+envValue)
		}
	}

	containerConfig := &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", appCfg.ContainerPort)): struct{}{},
		},
		Env: envs,
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", appCfg.ContainerPort)): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: fmt.Sprintf("%d", port),
				},
			},
		},
	}

	if len(appCfg.Volumes) > 0 {
		hostConfig.Binds = appCfg.Volumes
	}

	if appCfg.Network != "" {
		hostConfig.NetworkMode = container.NetworkMode(appCfg.Network)
	}

	resp, err := ds.Client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, err
	}

	err = ds.Client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("error allocating port: %w", err)
	}

	return &Container{
		ID:   resp.ID,
		Port: port,
	}, nil
}

func (ds *DockerService) FindContainer(imageName string) *Container {
	ctx := context.Background()

	containers, err := ds.Client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil
	}

	baseImageName := strings.Split(imageName, ":")[0]

	for _, container := range containers {
		// Inspect each container to get detailed information
		cont, err := ds.Client.ContainerInspect(ctx, container.ID)
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

func (ds *DockerService) StopContainer(containerID string) error {
	ctx := context.Background()

	defer ds.Client.Close()

	timeout := 15
	stopOptions := container.StopOptions{
		Timeout: &timeout,
	}

	err := ds.Client.ContainerStop(ctx, containerID, stopOptions)
	if err != nil {
		return err
	}

	err = ds.Client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (ds *DockerService) StreamLogs(container string, tail string) error {
	ctx := context.Background()
	defer ds.Client.Close()

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       tail,
		Details:    true,
	}

	out, err := ds.Client.ContainerLogs(ctx, container, options)
	if err != nil {
		return err
	}
	// skipcq: GO-S2307

	defer out.Close()

	reader := bufio.NewReader(out)

	for {
		// Docker log lines have a 8 byte header, 4 byte big endian timestamp, and then the log message
		header := make([]byte, 8)
		_, err := reader.Read(header)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Read the rest of the line as the log message
		message, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		// Print the log message
		fmt.Print(message)
	}

	return nil
}

func (ds *DockerService) GetStatus() error {
	containers, err := ds.Client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "CONTAINER ID\tIMAGE\tCREATED\tSTATUS\tPORTS\tNAMES")

	for _, container := range containers {
		ports := ""

		for _, port := range container.Ports {
			ports += fmt.Sprintf("%s:%d->%d/%s ", port.IP, port.PublicPort, port.PrivatePort, port.Type)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			container.ID[:10],
			container.Image,
			time.Since(time.Unix(container.Created, 0)),
			container.State,
			ports,
			container.Names)
	}

	w.Flush()

	return nil
}
