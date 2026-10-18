package mongo

import (
	"context"
	"fmt"
	"io"

	"dbbackup/src/internal/adapters/common"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "mongodb" }

func (a *Adapter) ClientTools() []string { return []string{"mongodump", "mongorestore"} }

func (a *Adapter) SupportedBackupModes() []common.BackupMode {
	return nil
}

func (a *Adapter) SelectiveRestoreSupported() bool { return false }

func (a *Adapter) ValidateConnection(_ context.Context, _ common.ConnectionConfig) error {
	return fmt.Errorf("mongodb backup is not implemented yet; use postgres for v1")
}

func (a *Adapter) DumpFull(_ context.Context, _ common.ConnectionConfig, _ io.Writer) error {
	return fmt.Errorf("mongodb backup is not implemented yet")
}

func (a *Adapter) RestoreFull(_ context.Context, _ common.ConnectionConfig, _ io.Reader, _ common.RestoreOptions) error {
	return fmt.Errorf("mongodb restore is not implemented yet")
}
