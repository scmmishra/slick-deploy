package docker

import (
	"errors"
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
	mockClient.On("ContainerRemove", mock.Anything, containerID, types.ContainerRemoveOptions{}).Return(nil)

	err := dockerService.StopContainer(containerID)
	assert.NoError(t, err)

	mockClient.AssertCalled(t, "ContainerStop", mock.Anything, containerID, container.StopOptions{
		Timeout: &timeout,
	})
	mockClient.AssertCalled(t, "ContainerRemove", mock.Anything, containerID, types.ContainerRemoveOptions{})
	mockClient.AssertExpectations(t)
}

func TestDockerService_StopContainer_StopError(t *testing.T) {
	mockClient := new(MockDockerClient)
	dockerService := NewDockerService(mockClient)

	containerID := "container123"
	timeout := 15

	mockClient.On("ContainerStop", mock.Anything, containerID, container.StopOptions{
		Timeout: &timeout,
	}).Return(errors.New("stop error"))

	err := dockerService.StopContainer(containerID)
	assert.Error(t, err)

	mockClient.AssertCalled(t, "ContainerStop", mock.Anything, containerID, container.StopOptions{
		Timeout: &timeout,
	})
	mockClient.AssertExpectations(t)
}

func TestDockerService_StopContainer_RemoveError(t *testing.T) {
	mockClient := new(MockDockerClient)
	dockerService := NewDockerService(mockClient)

	containerID := "container123"
	timeout := 15

	mockClient.On("ContainerStop", mock.Anything, containerID, container.StopOptions{
		Timeout: &timeout,
	}).Return(nil)
	mockClient.On("ContainerRemove", mock.Anything, containerID, types.ContainerRemoveOptions{}).Return(errors.New("remove error"))

	err := dockerService.StopContainer(containerID)
	assert.Error(t, err)

	mockClient.AssertCalled(t, "ContainerStop", mock.Anything, containerID, container.StopOptions{
		Timeout: &timeout,
	})
	mockClient.AssertCalled(t, "ContainerRemove", mock.Anything, containerID, types.ContainerRemoveOptions{})
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

func TestDockerService_FindContainer(t *testing.T) {
	mockClient := new(MockDockerClient)
	dockerService := NewDockerService(mockClient)

	imageName := "example/image:latest"
	// baseImageName := strings.Split(imageName, ":")[0]

	containerID := "container123"
	containerList := []types.Container{
		{
			ID: containerID,
			Names: []string{
				"test-container",
			},
		},
	}

	containerJSON := types.ContainerJSON{
		Config: &container.Config{
			Image: imageName,
		},
	}

	mockClient.On("ContainerList", mock.Anything, mock.AnythingOfType("types.ContainerListOptions")).Return(containerList, nil)
	mockClient.On("ContainerInspect", mock.Anything, containerID).Return(containerJSON, nil)

	container := dockerService.FindContainer(imageName)

	assert.NotNil(t, container)
	assert.Equal(t, containerID, container.ID)

	mockClient.AssertCalled(t, "ContainerList", mock.Anything, mock.AnythingOfType("types.ContainerListOptions"))
	mockClient.AssertCalled(t, "ContainerInspect", mock.Anything, containerID)
	mockClient.AssertExpectations(t)
}

func TestDockerService_FindContainer_NoMatch(t *testing.T) {
	mockClient := new(MockDockerClient)
	dockerService := NewDockerService(mockClient)

	imageName := "example/image:latest"
	// baseImageName := strings.Split(imageName, ":")[0]

	containerID := "container123"
	containerList := []types.Container{
		{
			ID: containerID,
			Names: []string{
				"test-container",
			},
		},
	}

	containerJSON := types.ContainerJSON{
		Config: &container.Config{
			Image: "different/image:latest",
		},
	}

	mockClient.On("ContainerList", mock.Anything, mock.AnythingOfType("types.ContainerListOptions")).Return(containerList, nil)
	mockClient.On("ContainerInspect", mock.Anything, containerID).Return(containerJSON, nil)

	container := dockerService.FindContainer(imageName)

	assert.Nil(t, container)

	mockClient.AssertCalled(t, "ContainerList", mock.Anything, mock.AnythingOfType("types.ContainerListOptions"))
	mockClient.AssertCalled(t, "ContainerInspect", mock.Anything, containerID)
	mockClient.AssertExpectations(t)
}

func TestDockerService_GetStatus(t *testing.T) {
	mockClient := new(MockDockerClient)
	dockerService := NewDockerService(mockClient)

	containerID := "container123"
	containerList := []types.Container{
		{
			ID: containerID,
			Names: []string{
				"test-container",
			},
			Ports: []types.Port{
				{
					IP:          "127.0.0.1",
					PrivatePort: 8080,
					PublicPort:  5000,
					Type:        "tcp",
				},
			},
		},
	}

	mockClient.On("ContainerList", mock.Anything, mock.AnythingOfType("types.ContainerListOptions")).Return(containerList, nil)

	dockerService.GetStatus()

	mockClient.AssertCalled(t, "ContainerList", mock.Anything, mock.AnythingOfType("types.ContainerListOptions"))
	mockClient.AssertExpectations(t)
}

func TestDockerService_GetStatus_Error(t *testing.T) {
	mockClient := new(MockDockerClient)
	dockerService := NewDockerService(mockClient)

	mockClient.On("ContainerList", mock.Anything, mock.AnythingOfType("types.ContainerListOptions")).Return(nil, errors.New("mock error"))

	err := dockerService.GetStatus()
	assert.Error(t, err)
	assert.Equal(t, "mock error", err.Error())

	mockClient.AssertCalled(t, "ContainerList", mock.Anything, mock.AnythingOfType("types.ContainerListOptions"))
	mockClient.AssertExpectations(t)
}
