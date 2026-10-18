package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"dbbackup/src/internal/adapters"
	"dbbackup/src/internal/backup"
	"dbbackup/src/internal/config"
	"dbbackup/src/internal/logutil"
	"dbbackup/src/internal/storage"
)

func newBackupCmd() *cobra.Command {
	var mode string
	var key string

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Run a full database backup",
		Long: `Creates a gzip-compressed custom-format PostgreSQL dump (pg_dump -Fc) and uploads it to the configured storage provider.

Flags:
  --mode full   Only full backups are supported in v1.
  --key         Optional object key or relative path under local storage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if mode != "full" {
				return fmt.Errorf("only --mode full is supported in v1")
			}
			dbType := viper.GetString("db")
			conn, err := config.ConnectionFromEnv(dbType)
			if err != nil {
				return err
			}
			adapter, err := adapters.Get(viper.GetString("db"))
			if err != nil {
				return err
			}
			store, err := storage.Get(viper.GetString("storage"), config.StorageFromEnv(viper.GetString("output-dir")))
			if err != nil {
				return err
			}
			runner := &backup.Runner{Log: logutil.New()}
			return runner.Backup(cmd.Context(), adapter, adapters.BackupMode(mode), store, conn, key)
		},
	}

	cmd.Flags().StringVar(&mode, "mode", "full", "backup mode (full only in v1)")
	cmd.Flags().StringVar(&key, "key", "", "backup object key or filename")
	cmd.SetContext(context.Background())
	return cmd
}
