package main

import (
	"fmt"

	"github.com/scmmishra/slick-deploy/internal/docker"
)

type DockerService interface {
	GetStatus() error
	FindContainer(imageName string) *docker.Container
	StreamLogs(containerID string, tail string) error
}

type DockerServiceCreator func() (DockerService, error)

var dockerServiceCreator DockerServiceCreator = func() (DockerService, error) {
	return newDockerService(docker.NewDockerClient)
}

type DockerClientCreator func() (docker.DockerClient, error)

func newDockerService(clientCreator DockerClientCreator) (DockerService, error) {
	cli, err := clientCreator()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return docker.NewDockerService(cli), nil
}
