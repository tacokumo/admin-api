package scenario_test

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func createPostgreSQLContainer(
	ctx context.Context,
) (*testcontainers.DockerContainer, error) {
	container, err := testcontainers.Run(
		ctx, "postgres:18.0",
		testcontainers.WithLogger(&SlogForTestContainers{}),
		testcontainers.WithExposedPorts("5432/tcp"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_PASSWORD": "password",
		}),
		testcontainers.WithFiles(
			testcontainers.ContainerFile{
				HostFilePath:      "./../../develop/tacokumo_admin_db/00-create-database.sql",
				ContainerFilePath: "/docker-entrypoint-initdb.d/00-create-database.sql",
				FileMode:          0644,
			},
			testcontainers.ContainerFile{
				HostFilePath:      "./../../develop/tacokumo_admin_db/10-create-user.sql",
				ContainerFilePath: "/docker-entrypoint-initdb.d/10-create-user.sql",
				FileMode:          0644,
			},
		),
		testcontainers.WithWaitStrategy(
			wait.ForExec([]string{"pg_isready", "-U", "admin_api", "-d", "tacokumo_admin_db"}).
				WithPollInterval(5*time.Second).
				WithStartupTimeout(25*time.Second),
		),
	)
	if err != nil {
		return nil, err
	}
	return container, nil
}

func applySchemaWithAtlas(
	ctx context.Context,
	postgresContainer *testcontainers.DockerContainer,
) (err error) {
	// Get network to connect atlas container to the same network as postgres
	networks, err := postgresContainer.Networks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get container networks: %w", err)
	}
	if len(networks) == 0 {
		return fmt.Errorf("postgres container has no networks")
	}

	// Get the container name for network connection
	inspect, err := postgresContainer.Inspect(ctx)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %w", err)
	}
	containerName := inspect.Name
	// Container name usually starts with '/', remove it
	if len(containerName) > 0 && containerName[0] == '/' {
		containerName = containerName[1:]
	}

	// Run atlas schema apply using GenericContainer for more control
	req := testcontainers.ContainerRequest{
		Image:    "arigaio/atlas:latest",
		Networks: networks,
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      "./../../sql/schema.sql",
				ContainerFilePath: "/schema.sql",
				FileMode:          0644,
			},
		},
		Cmd: []string{
			"schema", "apply",
			"--url", fmt.Sprintf("postgres://admin_api:password@%s:5432/tacokumo_admin_db?sslmode=disable", containerName),
			"--dev-url", fmt.Sprintf("postgres://postgres:password@%s:5432/postgres?sslmode=disable", containerName),
			"--to", "file:///schema.sql",
			"--auto-approve",
		},
	}

	atlasContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return fmt.Errorf("failed to run atlas: %w", err)
	}
	defer func() {
		err = atlasContainer.Terminate(ctx)
	}()

	return nil
}
