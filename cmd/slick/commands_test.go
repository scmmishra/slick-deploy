package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
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
	err := runDeploy(cmd, mockDeployer, mockConfigLoader)

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

func TestDefaultConfigLoader(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary .env file
	envContent := []byte("TEST_ENV_VAR=test_value")
	envPath := filepath.Join(tempDir, ".env")
	if err := os.WriteFile(envPath, envContent, 0644); err != nil {
		t.Fatalf("Failed to create temp .env file: %v", err)
	}

	// Create a temporary config.yaml file
	configContent := []byte(`
app:
  name: "test-app"
  image: "test-image:latest"
  container_port: 8080
caddy:
  admin_api: "http://localhost:2019"
  rules:
    - match: "example.com"
      reverse_proxy:
        - path: "/"
          to: "localhost:8080"
`)
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	// Create a cobra command with the necessary flags
	cmd := &cobra.Command{}
	cmd.Flags().String("config", configPath, "Path to config file")
	cmd.Flags().String("env", envPath, "Path to .env file")

	// Run the defaultConfigLoader
	cfg, err := defaultConfigLoader(cmd)

	// Assert no error occurred
	assert.NoError(t, err)

	// Assert the config was loaded correctly
	assert.Equal(t, "test-app", cfg.App.Name)
	assert.Equal(t, "test-image:latest", cfg.App.ImageName)
	assert.Equal(t, 8080, cfg.App.ContainerPort)
	assert.Equal(t, "http://localhost:2019", cfg.Caddy.AdminAPI)
	assert.Len(t, cfg.Caddy.Rules, 1)
	assert.Equal(t, "example.com", cfg.Caddy.Rules[0].Match)

	// Assert the environment variable was loaded
	assert.Equal(t, "test_value", os.Getenv("TEST_ENV_VAR"))
}

func TestDefaultConfigLoader_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		setupCmd    func() *cobra.Command
		expectedErr string
	}{
		{
			name: "Missing .env file",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("env", "non_existent.env", "Path to .env file")
				return cmd
			},
			expectedErr: "failed to load env file",
		},
		{
			name: "Missing config file",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("config", "non_existent.yaml", "Path to config file")
				cmd.Flags().String("env", os.DevNull, "Path to .env file")
				return cmd
			},
			expectedErr: "failed to load config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			_, err := defaultConfigLoader(cmd)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestNewDockerService(t *testing.T) {
	mockClientCreator := func() (docker.DockerClient, error) {
		return nil, errors.New("test error")
	}

	_, err := newDockerService(mockClientCreator)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create Docker client")
}
