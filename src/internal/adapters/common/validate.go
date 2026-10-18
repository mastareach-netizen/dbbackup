package common

import (
	"fmt"
	"strings"
)

func (c ConnectionConfig) ValidateForPostgres() error {
	if c.URL != "" {
		return nil
	}
	var missing []string
	if c.Host == "" {
		missing = append(missing, "host (PGHOST)")
	}
	if c.Port == 0 {
		missing = append(missing, "port (PGPORT)")
	}
	if c.User == "" {
		missing = append(missing, "user (PGUSER)")
	}
	if c.Database == "" {
		missing = append(missing, "database (PGDATABASE)")
	}
	if len(missing) > 0 {
		return fmt.Errorf("postgres connection incomplete: missing %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c ConnectionConfig) ValidateForMySQL() error {
	if c.URL != "" {
		return nil
	}
	var missing []string
	if c.Host == "" {
		missing = append(missing, "host (MYSQL_HOST)")
	}
	if c.Port == 0 {
		missing = append(missing, "port (MYSQL_PORT)")
	}
	if c.User == "" {
		missing = append(missing, "user (MYSQL_USER)")
	}
	if c.Database == "" {
		missing = append(missing, "database (MYSQL_DATABASE)")
	}
	if len(missing) > 0 {
		return fmt.Errorf("mysql connection incomplete: missing %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c ConnectionConfig) ValidateForMongo() error {
	return fmt.Errorf("mongodb adapter is not implemented yet")
}
