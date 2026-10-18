package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"dbbackup/src/internal/adapters"
	"dbbackup/src/internal/config"
	"dbbackup/src/internal/logutil"
)

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Validate database connectivity",
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
			if err := adapter.ValidateConnection(cmd.Context(), conn); err != nil {
				logutil.New().Error(err.Error())
				return err
			}
			fmt.Println("connection ok")
			return nil
		},
	}
	cmd.SetContext(context.Background())
	return cmd
}
