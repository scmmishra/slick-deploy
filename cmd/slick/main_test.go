package main

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDeployCmd_RunE(t *testing.T) {
	oldCmdFunctions := cmdFunctions
	defer func() { cmdFunctions = oldCmdFunctions }()

	cmdFunctions.RunDeploy = func(cmd *cobra.Command, deployer Deployer, configLoader ConfigLoader) error {
		return nil // Simulate successful deployment
	}

	cmd := &cobra.Command{}
	err := deployCmd.RunE(cmd, []string{})

	assert.NoError(t, err)
}

func TestStatusCmd_RunE(t *testing.T) {
	oldCmdFunctions := cmdFunctions
	defer func() { cmdFunctions = oldCmdFunctions }()

	cmdFunctions.RunStatus = func() error {
		return nil // Simulate successful status check
	}

	cmd := &cobra.Command{}
	err := statusCmd.RunE(cmd, []string{})

	assert.NoError(t, err)
}

func TestLogsCmd_RunE(t *testing.T) {
	oldCmdFunctions := cmdFunctions
	defer func() { cmdFunctions = oldCmdFunctions }()

	cmdFunctions.RunLogs = func(cmd *cobra.Command, configLoader ConfigLoader) error {
		return nil // Simulate successful log streaming
	}

	cmd := &cobra.Command{}
	err := logsCmd.RunE(cmd, []string{})

	assert.NoError(t, err)
}

func TestCaddyInspectCmd_RunE(t *testing.T) {
	oldCmdFunctions := cmdFunctions
	defer func() { cmdFunctions = oldCmdFunctions }()

	cmdFunctions.RunCaddyInspect = func(cmd *cobra.Command, configLoader ConfigLoader) error {
		return nil // Simulate successful Caddy inspection
	}

	cmd := &cobra.Command{}
	err := caddyInspectCmd.RunE(cmd, []string{})

	assert.NoError(t, err)
}

func TestCommands_RunE_Error(t *testing.T) {
	testCases := []struct {
		name    string
		cmd     *cobra.Command
		setupFn func()
	}{
		{
			name: "Deploy Error",
			cmd:  deployCmd,
			setupFn: func() {
				cmdFunctions.RunDeploy = func(cmd *cobra.Command, deployer Deployer, configLoader ConfigLoader) error {
					return errors.New("deploy error")
				}
			},
		},
		{
			name: "Status Error",
			cmd:  statusCmd,
			setupFn: func() {
				cmdFunctions.RunStatus = func() error {
					return errors.New("status error")
				}
			},
		},
		{
			name: "Logs Error",
			cmd:  logsCmd,
			setupFn: func() {
				cmdFunctions.RunLogs = func(cmd *cobra.Command, configLoader ConfigLoader) error {
					return errors.New("logs error")
				}
			},
		},
		{
			name: "Caddy Inspect Error",
			cmd:  caddyInspectCmd,
			setupFn: func() {
				cmdFunctions.RunCaddyInspect = func(cmd *cobra.Command, configLoader ConfigLoader) error {
					return errors.New("caddy inspect error")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oldCmdFunctions := cmdFunctions
			defer func() { cmdFunctions = oldCmdFunctions }()

			tc.setupFn()

			err := tc.cmd.RunE(tc.cmd, []string{})
			assert.Error(t, err)
		})
	}
}
