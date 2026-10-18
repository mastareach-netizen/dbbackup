package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"dbbackup/src/internal/adapters"
	"dbbackup/src/internal/backup"
	"dbbackup/src/internal/config"
	"dbbackup/src/internal/logutil"
	"dbbackup/src/internal/storage"
)

func newRestoreCmd() *cobra.Command {
	var key string
	var dropExisting bool

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore a full backup",
		Long: `Restores a gzip-compressed pg_dump custom archive using pg_restore.

Selective table restore is not supported for PostgreSQL in v1.`,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			opts := adapters.RestoreOptions{
				ArchivePath:  key,
				DropExisting: dropExisting,
			}
			return runner.Restore(cmd.Context(), adapter, store, conn, key, opts)
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "backup object key or filename (required)")
	cmd.Flags().BoolVar(&dropExisting, "clean", false, "drop existing objects before restore (pg_restore --clean)")
	_ = cmd.MarkFlagRequired("key")
	cmd.SetContext(context.Background())
	return cmd
}
