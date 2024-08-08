package main

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/scmmishra/slick-deploy/internal/docker"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockLoadConfig func(*cobra.Command) (config.DeploymentConfig, error)

type MockDeployer struct {
	mock.Mock
}

func (m *MockDeployer) Deploy(cfg config.DeploymentConfig) error {
	args := m.Called(cfg)
	return args.Error(0)
}

type MockDockerService struct {
	mock.Mock
}

func (m *MockDockerService) GetStatus() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDockerService) FindContainer(imageName string) *docker.Container {
	args := m.Called(imageName)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*docker.Container)
}

func (m *MockDockerService) StreamLogs(containerID, tail string) error {
	args := m.Called(containerID, tail)
	return args.Error(0)
}

// Helper function to create a cobra command for testing
func createTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.Flags().String("config", "", "Path to the configuration file")
	cmd.Flags().String("env", "", "Path to the env file")
	return cmd
}

func TestRunDeploy(t *testing.T) {
	mockDeployer := new(MockDeployer)
	mockDeployer.On("Deploy", mock.Anything).Return(nil)

	mockConfigLoader := func(*cobra.Command) (config.DeploymentConfig, error) {
		return config.DeploymentConfig{}, nil
	}

	cmd := createTestCommand()
	err := runDeploy(cmd, mockDeployer, mockConfigLoader)

	assert.NoError(t, err)
	mockDeployer.AssertExpectations(t)
}

func TestRunDeploy_ConfigLoaderError(t *testing.T) {
	mockDeployer := new(MockDeployer)
	mockConfigLoader := func(*cobra.Command) (config.DeploymentConfig, error) {
		return config.DeploymentConfig{}, errors.New("config load error")
	}

	cmd := createTestCommand()
	err := runDeploy(cmd, mockDeployer, mockConfigLoader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config load error")
	mockDeployer.AssertNotCalled(t, "Deploy")
}

func TestRunLogs(t *testing.T) {
	mockDockerService := new(MockDockerService)
	mockDockerService.On("FindContainer", mock.Anything).Return(&docker.Container{ID: "test-container"})
	mockDockerService.On("StreamLogs", "test-container", "all").Return(nil)

	mockConfigLoader := func(*cobra.Command) (config.DeploymentConfig, error) {
		return config.DeploymentConfig{
			App: config.App{ImageName: "test-image"},
		}, nil
	}

	originalDockerServiceCreator := dockerServiceCreator
	dockerServiceCreator = func() (DockerService, error) {
		return mockDockerService, nil
	}
	defer func() { dockerServiceCreator = originalDockerServiceCreator }()

	cmd := createTestCommand()
	cmd.Flags().String("tail", "all", "")
	err := runLogs(cmd, mockConfigLoader)

	assert.NoError(t, err)
	mockDockerService.AssertExpectations(t)
}

func TestRunLogs_NoContainer(t *testing.T) {
	mockDockerService := new(MockDockerService)
	mockDockerService.On("FindContainer", mock.Anything).Return(nil)

	mockConfigLoader := func(*cobra.Command) (config.DeploymentConfig, error) {
		return config.DeploymentConfig{
			App: config.App{ImageName: "test-image"},
		}, nil
	}

	originalDockerServiceCreator := dockerServiceCreator
	dockerServiceCreator = func() (DockerService, error) {
		return mockDockerService, nil
	}
	defer func() { dockerServiceCreator = originalDockerServiceCreator }()

	cmd := createTestCommand()
	cmd.Flags().String("tail", "all", "")
	err := runLogs(cmd, mockConfigLoader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no container found")
	mockDockerService.AssertExpectations(t)
}

func TestRunLogs_DockerServiceCreatorFails(t *testing.T) {
	mockConfigLoader := func(*cobra.Command) (config.DeploymentConfig, error) {
		return config.DeploymentConfig{
			App: config.App{ImageName: "test-image"},
		}, nil
	}

	originalDockerServiceCreator := dockerServiceCreator
	dockerServiceCreator = func() (DockerService, error) {
		return nil, errors.New("failed to create Docker service")
	}
	defer func() { dockerServiceCreator = originalDockerServiceCreator }()

	cmd := createTestCommand()
	cmd.Flags().String("tail", "all", "")
	err := runLogs(cmd, mockConfigLoader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create Docker service")
}

func TestRunLogs_ConfigLoaderFails(t *testing.T) {
	mockConfigLoader := func(*cobra.Command) (config.DeploymentConfig, error) {
		return config.DeploymentConfig{}, errors.New("config loading failed")
	}

	mockDockerService := new(MockDockerService)

	originalDockerServiceCreator := dockerServiceCreator
	dockerServiceCreator = func() (DockerService, error) {
		return mockDockerService, nil
	}
	defer func() { dockerServiceCreator = originalDockerServiceCreator }()

	cmd := createTestCommand()
	cmd.Flags().String("tail", "all", "")
	err := runLogs(cmd, mockConfigLoader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config loading failed")

	// Ensure that no methods on mockDockerService were called
	mockDockerService.AssertNotCalled(t, "FindContainer")
	mockDockerService.AssertNotCalled(t, "StreamLogs")
}

func TestRunCaddyInspect(t *testing.T) {
	mockConfigLoader := func(*cobra.Command) (config.DeploymentConfig, error) {
		return config.DeploymentConfig{
			Caddy: config.CaddyConfig{
				Rules: []config.Rule{
					{Match: "http://example.com"},
				},
			},
		}, nil
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := createTestCommand()
	err := runCaddyInspect(cmd, mockConfigLoader)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "http://")
}

func TestRunCaddyInspect_ConfigError(t *testing.T) {
	mockConfigLoader := func(*cobra.Command) (config.DeploymentConfig, error) {
		return config.DeploymentConfig{}, errors.New("config load error")
	}

	cmd := createTestCommand()
	err := runCaddyInspect(cmd, mockConfigLoader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config load error")
}

func TestRunStatus(t *testing.T) {
	mockDockerService := new(MockDockerService)
	mockDockerService.On("GetStatus").Return(nil)

	// Save the original dockerServiceCreator
	originalDockerServiceCreator := dockerServiceCreator

	// Replace dockerServiceCreator with a mock version
	dockerServiceCreator = func() (DockerService, error) {
		return mockDockerService, nil
	}

	// Restore the original dockerServiceCreator after the test
	defer func() {
		dockerServiceCreator = originalDockerServiceCreator
	}()

	err := runStatus()

	assert.NoError(t, err)
	mockDockerService.AssertExpectations(t)
}

func TestRunStatus_Error(t *testing.T) {
	mockDockerService := new(MockDockerService)
	mockDockerService.On("GetStatus").Return(errors.New("status error"))

	originalDockerServiceCreator := dockerServiceCreator
	dockerServiceCreator = func() (DockerService, error) {
		return mockDockerService, nil
	}
	defer func() { dockerServiceCreator = originalDockerServiceCreator }()

	err := runStatus()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status error")
	mockDockerService.AssertExpectations(t)
}

func TestRunStatus_DockerServiceCreatorFails(t *testing.T) {
	originalDockerServiceCreator := dockerServiceCreator
	dockerServiceCreator = func() (DockerService, error) {
		return nil, errors.New("failed to create Docker service")
	}
	defer func() { dockerServiceCreator = originalDockerServiceCreator }()

	err := runStatus()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create Docker service")
}
