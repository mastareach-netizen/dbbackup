package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "dbbackup",
		Short: "Cross-platform database backup and restore CLI",
		Long: `dbbackup streams compressed full backups to local disk or S3-compatible storage.

Connection (PostgreSQL v1): DATABASE_URL or PGHOST, PGPORT, PGUSER, PGPASSWORD, PGDATABASE, PGSSLMODE.
Storage: local directory (BACKUP_LOCAL_DIR) or S3 (S3_BUCKET, S3_PREFIX, AWS_REGION, optional S3_ENDPOINT).
Notifications: optional WEBHOOK_URL (JSON payload, no secrets).

MySQL and MongoDB adapters are registered but return clear not-implemented errors in v1.`,
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (YAML)")
	root.PersistentFlags().String("db", "postgres", "database type: postgres, mysql, mongodb")
	root.PersistentFlags().String("storage", "local", "storage provider: local, s3")
	root.PersistentFlags().String("output-dir", "", "local backup directory (overrides BACKUP_LOCAL_DIR)")

	_ = viper.BindPFlag("db", root.PersistentFlags().Lookup("db"))
	_ = viper.BindPFlag("storage", root.PersistentFlags().Lookup("storage"))
	_ = viper.BindPFlag("output-dir", root.PersistentFlags().Lookup("output-dir"))

	cobra.OnInitialize(initConfig)

	root.AddCommand(newBackupCmd())
	root.AddCommand(newRestoreCmd())
	root.AddCommand(newCheckCmd())

	return root
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath(os.Getenv("HOME"))
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "using config file: %s\n", viper.ConfigFileUsed())
	}
}

func Execute() {
	if err := NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
