package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

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
