package db

import (
	"context"
	"fmt"
	"os"
	"skyfox/config"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type DatabaseContainer struct {
	Container  testcontainers.Container
	MappedPort nat.Port
}

func CreateTestContainer(ctx context.Context, cfg config.DbConfig) (*DatabaseContainer, error) {

	var env = map[string]string{
		"POSTGRES_PASSWORD": cfg.Password,
		"POSTGRES_USER":     cfg.User,
		"POSTGRES_DB":       cfg.Name,
	}
	var port = "5432/tcp"
	dbURL := func(host string, port nat.Port) string {
		return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", cfg.User, cfg.Password, port.Port(), cfg.Name)
	}
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:latest",
			ExposedPorts: []string{port},
			Cmd:          []string{"postgres", "-c", "fsync=off"},
			Env:          env,
			WaitingFor:   wait.ForSQL(nat.Port(port), "postgres", dbURL).WithStartupTimeout(time.Second * 5),
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %s", err)
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return nil, fmt.Errorf("failed to get container external port: %s", err)
	}

	return &DatabaseContainer{
		Container:  container,
		MappedPort: mappedPort,
	}, nil

}

func SetupTestContainerEnv() {
	os.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", "artifactory.idfcfirstbank.com/neev-docker/")
	os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock")
	os.Setenv("DOCKER_HOST", fmt.Sprintf("unix://%s/.colima/docker.sock", os.Getenv("HOME")))
}
