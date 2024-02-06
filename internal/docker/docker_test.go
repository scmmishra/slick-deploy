package docker

import (
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDockerService_PullImage(t *testing.T) {
	mockClient := new(MockDockerClient)
	dockerService := NewDockerService(mockClient)

	imageName := "example/image:latest"
	registryConfig := config.RegistryConfig{
		Username: "testuser",
		Password: "testpass",
	}

	mockClient.On("ImagePull", mock.Anything, imageName, mock.AnythingOfType("types.ImagePullOptions")).Return(io.NopCloser(strings.NewReader("")), nil)

	err := dockerService.PullImage(imageName, registryConfig)
	assert.NoError(t, err)

	mockClient.AssertCalled(t, "ImagePull", mock.Anything, imageName, mock.AnythingOfType("types.ImagePullOptions"))
	mockClient.AssertExpectations(t)
}

func TestDockerService_RunContainer(t *testing.T) {
	mockClient := new(MockDockerClient)
	dockerService := NewDockerService(mockClient)

	imageName := "example/image:latest"
	cfg := config.App{
		Name:      "test-app",
		ImageName: "example/image:latest",
		ENV: []string{
			"TEST_ENV",
			"ANOTHER_ENV",
		},
		Network:       "slick-test",
		ContainerPort: 8080,
		PortRange: config.PortRange{
			Start: 5000,
			End:   6000,
		},
	}

	containerID := "container123"
	mockClient.On("ContainerCreate", mock.Anything, mock.AnythingOfType("*container.Config"), mock.AnythingOfType("*container.HostConfig"), mock.Anything, mock.Anything, "").Return(container.CreateResponse{ID: containerID}, nil)
	mockClient.On("ContainerStart", mock.Anything, containerID, types.ContainerStartOptions{}).Return(nil)

	newContainer, err := dockerService.RunContainer(imageName, cfg)
	assert.NoError(t, err)
	assert.Equal(t, containerID, newContainer.ID)

	mockClient.AssertExpectations(t)
}

func TestDockerService_StopContainer(t *testing.T) {
	mockClient := new(MockDockerClient)
	dockerService := NewDockerService(mockClient)

	containerID := "container123"
	timeout := 15
	mockClient.On("ContainerStop", mock.Anything, containerID, container.StopOptions{
		Timeout: &timeout,
	}).Return(nil)

	err := dockerService.StopContainer(containerID)
	assert.NoError(t, err)

	mockClient.AssertCalled(t, "ContainerStop", mock.Anything, containerID, container.StopOptions{
		Timeout: &timeout,
	})
	mockClient.AssertExpectations(t)
}

func TestDockerService_StreamLogs(t *testing.T) {
	mockClient := new(MockDockerClient)
	dockerService := NewDockerService(mockClient)

	containerID := "container123"
	logStream := "test log stream\nmore logs\n"
	mockClient.On("ContainerLogs", mock.Anything, containerID, mock.AnythingOfType("types.ContainerLogsOptions")).Return(io.NopCloser(strings.NewReader(logStream)), nil)

	err := dockerService.StreamLogs(containerID, "all")
	assert.NoError(t, err)

	mockClient.AssertCalled(t, "ContainerLogs", mock.Anything, containerID, mock.AnythingOfType("types.ContainerLogsOptions"))
	mockClient.AssertExpectations(t)
}
