package postgres

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"dbbackup/src/internal/adapters/common"
	"dbbackup/src/internal/logutil"
)

type Adapter struct {
	PgDumpPath    string
	PgRestorePath string
	PgIsReadyPath string
}

func New() *Adapter {
	return &Adapter{
		PgDumpPath:    envOr("PG_DUMP", "pg_dump"),
		PgRestorePath: envOr("PG_RESTORE", "pg_restore"),
		PgIsReadyPath: envOr("PG_ISREADY", "pg_isready"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (a *Adapter) Name() string { return "postgres" }

func (a *Adapter) ClientTools() []string {
	return []string{a.PgDumpPath, a.PgRestorePath, a.PgIsReadyPath}
}

func (a *Adapter) SupportedBackupModes() []common.BackupMode {
	return []common.BackupMode{common.ModeFull}
}

func (a *Adapter) SelectiveRestoreSupported() bool { return false }

func (a *Adapter) ValidateConnection(ctx context.Context, cfg common.ConnectionConfig) error {
	if err := cfg.ValidateForPostgres(); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, a.PgIsReadyPath, a.isReadyArgs(cfg)...)
	cmd.Env = a.toolEnv(cfg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("postgres connection check failed: %s", logutil.Sanitize(strings.TrimSpace(string(out))))
	}
	return nil
}

func (a *Adapter) DumpFull(ctx context.Context, cfg common.ConnectionConfig, w io.Writer) error {
	if err := cfg.ValidateForPostgres(); err != nil {
		return err
	}
	args := []string{"--format=custom", "--no-owner", "--no-acl"}
	args = append(args, a.connectionArgs(cfg)...)
	cmd := exec.CommandContext(ctx, a.PgDumpPath, args...)
	cmd.Env = a.toolEnv(cfg)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failed: %w", err)
	}
	return nil
}

func (a *Adapter) RestoreFull(ctx context.Context, cfg common.ConnectionConfig, r io.Reader, opts common.RestoreOptions) error {
	if err := cfg.ValidateForPostgres(); err != nil {
		return err
	}
	args := []string{"--no-owner", "--no-acl"}
	if opts.DropExisting {
		args = append(args, "--clean", "--if-exists")
	}
	args = append(args, a.connectionArgs(cfg)...)
	args = append(args, "-")
	cmd := exec.CommandContext(ctx, a.PgRestorePath, args...)
	cmd.Env = a.toolEnv(cfg)
	cmd.Stdin = r
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_restore failed: %w", err)
	}
	return nil
}

func (a *Adapter) connectionArgs(cfg common.ConnectionConfig) []string {
	if cfg.URL != "" {
		return []string{"--dbname", cfg.URL}
	}
	args := []string{
		"--host", cfg.Host,
		"--port", strconv.Itoa(cfg.Port),
		"--username", cfg.User,
		"--dbname", cfg.Database,
	}
	return args
}

func (a *Adapter) isReadyArgs(cfg common.ConnectionConfig) []string {
	if cfg.URL != "" {
		return []string{"-d", cfg.URL, "-q"}
	}
	return []string{
		"-h", cfg.Host,
		"-p", strconv.Itoa(cfg.Port),
		"-U", cfg.User,
		"-d", cfg.Database,
		"-q",
	}
}

func (a *Adapter) toolEnv(cfg common.ConnectionConfig) []string {
	env := os.Environ()
	if cfg.Password != "" {
		env = append(env, "PGPASSWORD="+cfg.Password)
	}
	if cfg.SSLMode != "" {
		env = append(env, "PGSSLMODE="+cfg.SSLMode)
	}
	return env
}

func ParseDatabaseURL(raw string) (common.ConnectionConfig, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return common.ConnectionConfig{}, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return common.ConnectionConfig{}, fmt.Errorf("DATABASE_URL scheme must be postgres or postgresql")
	}
	cfg := common.ConnectionConfig{URL: raw}
	if u.User != nil {
		cfg.User = u.User.Username()
		if p, ok := u.User.Password(); ok {
			cfg.Password = p
		}
	}
	cfg.Host = u.Hostname()
	if port := u.Port(); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return common.ConnectionConfig{}, fmt.Errorf("invalid port in DATABASE_URL")
		}
		cfg.Port = p
	} else {
		cfg.Port = 5432
	}
	cfg.Database = strings.TrimPrefix(u.Path, "/")
	q := u.Query()
	if ssl := q.Get("sslmode"); ssl != "" {
		cfg.SSLMode = ssl
	}
	return cfg, nil
}
