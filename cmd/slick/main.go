package main

import (
	"fmt"

	"os"

	"github.com/scmmishra/slick-deploy/internal/caddy"
	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/scmmishra/slick-deploy/internal/deploy"
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
		// if cfgPath is not preset, use the slick.yml in the current directory
		if cfgPath == "" {
			cfgPath = "slick.yml"
		}

		// Load configuration
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			cmd.PrintErrf("Failed to pull image: %v", err)
			os.Exit(1)
		}

		port, err := deploy.Deploy(cfg)
		if err != nil {
			cmd.PrintErrf("Failed to pull image: %v", err)
			os.Exit(1)
		}

		err = caddy.SetupCaddy(port, cfg)
		if err != nil {
			cmd.PrintErrf("Failed to setup Caddy: %v", err)
			os.Exit(1)
		}
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
	deployCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Path to the configuration file")
	rootCmd.AddCommand(deployCmd)
}

func main() {
	Execute()
}
