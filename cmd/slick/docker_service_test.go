package main

import (
	"errors"
	"testing"

	"github.com/scmmishra/slick-deploy/internal/docker"
	"github.com/stretchr/testify/assert"
)

func TestNewDockerService(t *testing.T) {
	mockClientCreator := func() (docker.DockerClient, error) {
		return nil, errors.New("test error")
	}

	_, err := newDockerService(mockClientCreator)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create Docker client")
}
