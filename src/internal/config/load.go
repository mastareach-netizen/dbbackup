package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"dbbackup/src/internal/adapters"
	"dbbackup/src/internal/adapters/mysql"
	"dbbackup/src/internal/adapters/postgres"
	"dbbackup/src/internal/storage"
)

func ConnectionFromEnv(dbType string) (adapters.ConnectionConfig, error) {
	dbType = strings.ToLower(dbType)
	if raw := os.Getenv("DATABASE_URL"); raw != "" {
		switch {
		case strings.HasPrefix(raw, "postgres://") || strings.HasPrefix(raw, "postgresql://"):
			if dbType != "postgres" && dbType != "postgresql" && dbType != "pg" {
				return adapters.ConnectionConfig{}, fmt.Errorf("DATABASE_URL is postgres but --db is %q", dbType)
			}
			return postgres.ParseDatabaseURL(raw)
		case strings.HasPrefix(raw, "mysql://") || strings.HasPrefix(raw, "mariadb://"):
			if dbType != "mysql" && dbType != "mariadb" {
				return adapters.ConnectionConfig{}, fmt.Errorf("DATABASE_URL is mysql but --db is %q", dbType)
			}
			return mysql.ParseDatabaseURL(raw)
		default:
			return adapters.ConnectionConfig{}, fmt.Errorf("unsupported DATABASE_URL scheme")
		}
	}

	switch dbType {
	case "postgres", "postgresql", "pg":
		return postgresFromEnv(), nil
	case "mysql", "mariadb":
		return mysqlFromEnv(), nil
	case "mongodb", "mongo":
		return adapters.ConnectionConfig{
			Host: os.Getenv("MONGO_HOST"),
		}, nil
	default:
		return adapters.ConnectionConfig{}, fmt.Errorf("unsupported database %q", dbType)
	}
}

func postgresFromEnv() adapters.ConnectionConfig {
	port := 5432
	if p := os.Getenv("PGPORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}
	return adapters.ConnectionConfig{
		Host:     os.Getenv("PGHOST"),
		Port:     port,
		User:     os.Getenv("PGUSER"),
		Password: os.Getenv("PGPASSWORD"),
		Database: os.Getenv("PGDATABASE"),
		SSLMode:  os.Getenv("PGSSLMODE"),
	}
}

func mysqlFromEnv() adapters.ConnectionConfig {
	port := 3306
	if p := os.Getenv("MYSQL_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}
	return adapters.ConnectionConfig{
		Host:     envFirst("MYSQL_HOST", "MYSQLHOST"),
		Port:     port,
		User:     envFirst("MYSQL_USER", "MYSQLUSER"),
		Password: envFirst("MYSQL_PASSWORD", "MYSQLPWD"),
		Database: envFirst("MYSQL_DATABASE", "MYSQLDATABASE"),
		SSLMode:  os.Getenv("MYSQL_TLS"),
	}
}

func envFirst(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func StorageFromEnv(localOverride string) storage.Config {
	dir := localOverride
	if dir == "" {
		dir = os.Getenv("BACKUP_LOCAL_DIR")
	}
	return storage.Config{
		LocalDir: dir,
		S3: storage.S3Config{
			Bucket:   os.Getenv("S3_BUCKET"),
			Prefix:   os.Getenv("S3_PREFIX"),
			Region:   os.Getenv("AWS_REGION"),
			Endpoint: os.Getenv("S3_ENDPOINT"),
		},
	}
}
