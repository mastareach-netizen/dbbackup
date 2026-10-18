package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

type UploadResult struct {
	Location string
	Bytes    int64
}

type ObjectInfo struct {
	Key      string
	Size     int64
	Modified time.Time
}

type Provider interface {
	Name() string
	Upload(ctx context.Context, key string, r io.Reader) (string, error)
	Download(ctx context.Context, key string, w io.Writer) error
	List(ctx context.Context, prefix string) ([]ObjectInfo, error)
	Delete(ctx context.Context, key string) error
}

func Get(name string, cfg Config) (Provider, error) {
	switch strings.ToLower(name) {
	case "local", "filesystem", "fs":
		return NewLocal(cfg.LocalDir)
	case "s3", "s3-compatible":
		return NewS3(cfg.S3)
	default:
		return nil, fmt.Errorf("unsupported storage %q (supported: local, s3)", name)
	}
}

type Config struct {
	LocalDir string
	S3       S3Config
}

type S3Config struct {
	Bucket   string
	Prefix   string
	Region   string
	Endpoint string
}
