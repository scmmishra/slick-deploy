package main

import (
	"fmt"
	"os"

	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/scmmishra/slick-deploy/internal/deploy"
	"github.com/scmmishra/slick-deploy/internal/docker"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "slick",
	Short: "Slick is a CLI tool for zero-downtime deployment using Docker and Caddy",
	Long:  "Slick is designed to simplify your deployment process ensuring zero downtime\nand easy configuration management.",
}

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your application with zero downtime",
	Long: `The deploy command starts a new deployment process
ensuring that your application is updated with no service interruption.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, _ := cmd.Flags().GetString("config")

		// load the .env file
		envPath, _ := cmd.Flags().GetString("env")

		err := godotenv.Load(envPath)
		if err != nil {
			cmd.PrintErrf("Failed to load env file: %v", err)
			os.Exit(1)
		}

		// Load configuration
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			cmd.PrintErrf("Failed to load config: %v", err)
			os.Exit(1)
		}

		err = deploy.Deploy(cfg)
		if err != nil {
			cmd.PrintErrf("Failed to deploy image: %v", err)
			os.Exit(1)
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of your application",
	Long:  `The status command shows the status of your application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize Docker client
		cli, err := docker.NewDockerClient()
		if err != nil {
			cmd.PrintErrf("Failed to create Docker client: %v", err)
			os.Exit(1)
		}

		// Create DockerService instance
		dockerService := docker.NewDockerService(cli)

		err = dockerService.GetStatus()
		if err != nil {
			cmd.PrintErrf("Failed to get status: %v", err)
			os.Exit(1)
		}
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Tail and follow app logs",
	Long:  `The logs command shows the logs output of your application. It is similar to running 'docker logs -f <container-id>'`,
	Run: func(cmd *cobra.Command, args []string) {
		// if cfgPath is not preset, use the slick.yml in the current directory
		cfgPath, _ := cmd.Flags().GetString("config")
		tail, _ := cmd.Flags().GetString("tail")

		// Initialize Docker client
		cli, err := docker.NewDockerClient()
		if err != nil {
			cmd.PrintErrf("Failed to create Docker client: %v", err)
			os.Exit(1)
		}

		// Create DockerService instance
		dockerService := docker.NewDockerService(cli)

		// Load configuration
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			cmd.PrintErrf("Failed to load config: %v", err)
			os.Exit(1)
		}

		container := dockerService.FindContainer(cfg.App.ImageName)
		if container == nil {
			cmd.PrintErrf("No container found")
			os.Exit(1)
		}

		dockerService.StreamLogs(container.ID, tail)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// register deploy command
	deployCmd.Flags().StringP("config", "c", "slick.yml", "Path to the configuration file")
	deployCmd.Flags().StringP("env", "e", ".env", "Path to the env file")
	rootCmd.AddCommand(deployCmd)

	// register status command
	rootCmd.AddCommand(statusCmd)

	logsCmd.Flags().StringP("config", "c", "slick.yml", "Path to the configuration file")
	logsCmd.Flags().StringP("tail", "t", "all", "Tail logs")
	rootCmd.AddCommand(logsCmd)
}

func main() {
	Execute()
}
