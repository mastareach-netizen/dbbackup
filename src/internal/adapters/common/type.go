package common

import (
	"context"
	"io"
)

type BackupMode string

const (
	ModeFull BackupMode = "full"
)

type ConnectionConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	URL      string
}

type RestoreOptions struct {
	ArchivePath  string
	DropExisting bool
}

type Adapter interface {
	Name() string
	ClientTools() []string
	ValidateConnection(ctx context.Context, cfg ConnectionConfig) error
	SupportedBackupModes() []BackupMode
	DumpFull(ctx context.Context, cfg ConnectionConfig, w io.Writer) error
	RestoreFull(ctx context.Context, cfg ConnectionConfig, r io.Reader, opts RestoreOptions) error
	SelectiveRestoreSupported() bool
}
