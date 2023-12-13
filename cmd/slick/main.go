package main

import (
	"fmt"
	"os"

	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/scmmishra/slick-deploy/internal/deploy"
	"github.com/scmmishra/slick-deploy/internal/status"

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
		// load the .env file
		envPath, _ := cmd.Flags().GetString("env")
		if envPath == "" {
			envPath = ".env"
		}

		err := godotenv.Load(envPath)
		if err != nil {
			cmd.PrintErrf("Failed to load env file: %v", err)
			os.Exit(1)
		}

		cfgPath, _ := cmd.Flags().GetString("config")
		// if cfgPath is not preset, use the slick.yml in the current directory
		if cfgPath == "" {
			cfgPath = "slick.yml"
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
		status.GetStatus()
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
	var cfgFile string
	var envFile string
	deployCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Path to the configuration file")
	deployCmd.Flags().StringVarP(&envFile, "env", "e", "", "Path to the env file")
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(statusCmd)
}

func main() {
	Execute()
}
