package adapters

import (
	"fmt"
	"strings"

	"dbbackup/src/internal/adapters/mongo"
	"dbbackup/src/internal/adapters/mysql"
	"dbbackup/src/internal/adapters/postgres"
)

func Get(name string) (Adapter, error) {
	switch strings.ToLower(name) {
	case "postgres", "postgresql", "pg":
		return postgres.New(), nil
	case "mysql", "mariadb":
		return mysql.New(), nil
	case "mongodb", "mongo":
		return mongo.New(), nil
	default:
		return nil, fmt.Errorf("unsupported database %q (supported: postgres, mysql, mongodb)", name)
	}
}
