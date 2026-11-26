package container_builders

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	postgresContainer   *PostgresContainerBuilder
	postgresContainerMu sync.Mutex
)

type PostgresContainerBuilder struct {
	container       testcontainers.Container
	connStrTemplate string
	host            string
	port            int
	containerMu     sync.Mutex
	databasesMu     sync.Mutex
	databases       []string
}

func GetPostgresContainerBuilder() *PostgresContainerBuilder {
	postgresContainerMu.Lock()
	defer postgresContainerMu.Unlock()

	if postgresContainer == nil {
		postgresContainer = &PostgresContainerBuilder{}
	}
	return postgresContainer
}

func (b *PostgresContainerBuilder) start() string {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container == nil {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "postgres:18-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "postgres",
			},
			// postgres logs ready twice: the init server, then the real one.
			WaitingFor: wait.ForAll(
				wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
				wait.ForListeningPort("5432/tcp"),
			),
		}

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			log.Fatalf("failed to start postgres container: %v", err)
		}
		b.container = container

		host, _ := container.Host(ctx)
		port, _ := container.MappedPort(ctx, "5432")
		b.host = host
		b.port, _ = strconv.Atoi(port.Port())

		b.connStrTemplate = fmt.Sprintf(
			"host=%s port=%s dbname=%s user=postgres password=postgres sslmode=disable",
			host, port.Port(), "%s",
		)
		log.Println("PostgreSQL container ready")
	}
	return b.connStrTemplate
}

func (b *PostgresContainerBuilder) stop() {
	b.containerMu.Lock()
	defer b.containerMu.Unlock()

	if b.container != nil {
		_ = b.container.Terminate(context.Background())
		b.container = nil
		log.Println("PostgreSQL container stopped")
	}
	postgresContainerMu.Lock()
	postgresContainer = nil
	postgresContainerMu.Unlock()
}

// Create returns (dbName, connectionString) for a scratch DB in the shared container, so tests run in parallel.
func (b *PostgresContainerBuilder) Create() (string, string) {
	dnsTemplate := b.start()
	mainDsn := fmt.Sprintf(dnsTemplate, "postgres")

	conn, err := sql.Open("postgres", mainDsn)
	if err != nil {
		log.Fatalf("failed to open postgres: %v", err)
	}
	defer conn.Close()

	dbName := "test_" + uuid.New().String()[:8]
	if _, err := conn.Exec(fmt.Sprintf(`CREATE DATABASE "%s";`, dbName)); err != nil {
		log.Fatalf("failed to create db %s: %v", dbName, err)
	}

	b.databasesMu.Lock()
	b.databases = append(b.databases, dbName)
	b.databasesMu.Unlock()

	return dbName, fmt.Sprintf(dnsTemplate, dbName)
}

// AdminEndpoint returns where the shared server listens; its admin is postgres/postgres.
func (b *PostgresContainerBuilder) AdminEndpoint() (string, int) {
	b.start()
	return b.host, b.port
}

func (b *PostgresContainerBuilder) DropDatabase(dbName string) {
	b.databasesMu.Lock()
	defer b.databasesMu.Unlock()
	for i, db := range b.databases {
		if db == dbName {
			b.databases = append(b.databases[:i], b.databases[i+1:]...)
			break
		}
	}
}
