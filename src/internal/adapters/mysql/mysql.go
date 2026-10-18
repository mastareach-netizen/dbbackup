package mysql

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
	DumpPath   string
	ClientPath string
}

func New() *Adapter {
	return &Adapter{
		DumpPath:   envOr("MYSQLDUMP", "mysqldump"),
		ClientPath: envOr("MYSQL", "mysql"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (a *Adapter) Name() string { return "mysql" }

func (a *Adapter) ClientTools() []string {
	return []string{a.DumpPath, a.ClientPath}
}

func (a *Adapter) SupportedBackupModes() []common.BackupMode {
	return []common.BackupMode{common.ModeFull}
}

func (a *Adapter) SelectiveRestoreSupported() bool { return false }

func (a *Adapter) ValidateConnection(ctx context.Context, cfg common.ConnectionConfig) error {
	if err := cfg.ValidateForMySQL(); err != nil {
		return err
	}
	args := append(a.baseArgs(cfg), "-e", "SELECT 1")
	cmd := exec.CommandContext(ctx, a.ClientPath, args...)
	cmd.Env = a.toolEnv(cfg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mysql connection check failed: %s", logutil.Sanitize(strings.TrimSpace(string(out))))
	}
	return nil
}

func (a *Adapter) DumpFull(ctx context.Context, cfg common.ConnectionConfig, w io.Writer) error {
	if err := cfg.ValidateForMySQL(); err != nil {
		return err
	}
	args := append(a.baseArgs(cfg),
		"--single-transaction",
		"--quick",
		"--routines",
		"--triggers",
	)
	cmd := exec.CommandContext(ctx, a.DumpPath, args...)
	cmd.Env = a.toolEnv(cfg)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysqldump failed: %w", err)
	}
	return nil
}

func (a *Adapter) RestoreFull(ctx context.Context, cfg common.ConnectionConfig, r io.Reader, _ common.RestoreOptions) error {
	if err := cfg.ValidateForMySQL(); err != nil {
		return err
	}
	args := a.baseArgs(cfg)
	cmd := exec.CommandContext(ctx, a.ClientPath, args...)
	cmd.Env = a.toolEnv(cfg)
	cmd.Stdin = r
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysql restore failed: %w", err)
	}
	return nil
}

func (a *Adapter) baseArgs(cfg common.ConnectionConfig) []string {
	if cfg.URL != "" {
		parsed, err := ParseDatabaseURL(cfg.URL)
		if err == nil {
			cfg = parsed
		}
	}
	return []string{
		"--host=" + cfg.Host,
		"--port=" + strconv.Itoa(cfg.Port),
		"--user=" + cfg.User,
		cfg.Database,
	}
}

func (a *Adapter) toolEnv(cfg common.ConnectionConfig) []string {
	env := os.Environ()
	if cfg.Password != "" {
		env = append(env, "MYSQL_PWD="+cfg.Password)
	}
	return env
}

func ParseDatabaseURL(raw string) (common.ConnectionConfig, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return common.ConnectionConfig{}, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}
	if u.Scheme != "mysql" && u.Scheme != "mariadb" {
		return common.ConnectionConfig{}, fmt.Errorf("DATABASE_URL scheme must be mysql or mariadb")
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
		cfg.Port = 3306
	}
	cfg.Database = strings.TrimPrefix(u.Path, "/")
	if tls := u.Query().Get("tls"); tls != "" {
		cfg.SSLMode = tls
	}
	return cfg, nil
}
