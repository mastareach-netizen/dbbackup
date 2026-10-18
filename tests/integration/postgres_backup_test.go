//go:build integration

package integration_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPostgresBackupCheckIntegration(t *testing.T) {
	if _, err := exec.LookPath("pg_dump"); err != nil {
		t.Skip("pg_dump not on PATH")
	}
	ctx := context.Background()
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("app"),
		postgres.WithUsername("app"),
		postgres.WithPassword("secret"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Skipf("docker/testcontainers unavailable: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("PGHOST", host)
	t.Setenv("PGPORT", port.Port())
	t.Setenv("PGUSER", "app")
	t.Setenv("PGPASSWORD", "secret")
	t.Setenv("PGDATABASE", "app")

	bin := buildCLIBinary(t)
	cmd := exec.Command(bin, "check", "--db", "postgres")
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check failed: %v\n%s", err, out)
	}
}

func buildCLIBinary(t *testing.T) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "dbbackup")
	cmd := exec.Command("go", "build", "-o", out, "./src/cmd/dbbackup")
	cmd.Dir = mustRepoRoot(t)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build cli: %v\n%s", err, b)
	}
	return out
}

func mustRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}
