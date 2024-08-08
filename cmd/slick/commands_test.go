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
	return args.Get(0).(*docker.Container)
}

func (m *MockDockerService) StreamLogs(containerID string, tail string) error {
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
	err := runDeploy(cmd, []string{}, mockDeployer, mockConfigLoader)

	assert.NoError(t, err)
	mockDeployer.AssertExpectations(t)
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
	err := runLogs(cmd, []string{}, mockConfigLoader)

	assert.NoError(t, err)
	mockDockerService.AssertExpectations(t)
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
	err := runCaddyInspect(cmd, []string{}, mockConfigLoader)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	assert.Contains(t, output, "http://")
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

	cmd := createTestCommand()
	err := runStatus(cmd, []string{})

	assert.NoError(t, err)
	mockDockerService.AssertExpectations(t)
}

func TestNewDockerService(t *testing.T) {
	mockClientCreator := func() (docker.DockerClient, error) {
		return nil, errors.New("test error")
	}

	_, err := newDockerService(mockClientCreator)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create Docker client")
}
