package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "slick",
	Short: "Slick is a CLI tool for zero-downtime deployment using Docker and Caddy",
	Long:  "Slick is designed to simplify your deployment process ensuring zero downtime and easy configuration management.",
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your application with zero downtime",
	Long:  "The deploy command starts a new deployment process ensuring that your application is updated with no service interruption.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDeploy(cmd, args, defaultDeployer)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of your application",
	Long:  "The status command shows the status of your application.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus(cmd, args, defaultDockerServiceCreator)
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Tail and follow app logs",
	Long:  "The logs command shows the logs output of your application. It is similar to running 'docker logs -f <container-id>'",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLogs(cmd, args, defaultDockerServiceCreator)
	},
}

var caddyInspectCmd = &cobra.Command{
	Use:   "caddy-inspect",
	Short: "Inspect the current Caddy configuration",
	Long:  "The caddy-inspect command retrieves and displays the current Caddy configuration.",
	RunE:  runCaddyInspect,
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "slick.yml", "Path to the configuration file")
	rootCmd.PersistentFlags().StringP("env", "e", ".env", "Path to the env file")

	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(caddyInspectCmd)

	logsCmd.Flags().StringP("tail", "t", "all", "Tail logs")
}
