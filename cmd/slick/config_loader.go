package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/spf13/cobra"
)

type ConfigLoader func(*cobra.Command) (config.DeploymentConfig, error)
func defaultConfigLoader(cmd *cobra.Command) (config.DeploymentConfig, error) {
	cfgPath, _ := cmd.Flags().GetString("config")
	envPath, _ := cmd.Flags().GetString("env")

	if err := godotenv.Load(envPath); err != nil {
		return config.DeploymentConfig{}, fmt.Errorf("failed to load env file: %w", err)
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return config.DeploymentConfig{}, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}
