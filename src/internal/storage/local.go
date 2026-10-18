package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Local struct {
	baseDir string
}

func NewLocal(baseDir string) (*Local, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("local storage requires an output directory")
	}
	return &Local{baseDir: baseDir}, nil
}

func (l *Local) Name() string { return "local" }

func (l *Local) Upload(_ context.Context, key string, r io.Reader) (string, error) {
	path := filepath.Join(l.baseDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create backup file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("write backup file: %w", err)
	}
	return path, nil
}

func (l *Local) Download(_ context.Context, key string, w io.Writer) error {
	path := filepath.Join(l.baseDir, filepath.FromSlash(key))
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open backup file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("read backup file: %w", err)
	}
	return nil
}

func (l *Local) List(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	var result []ObjectInfo
	root := filepath.Join(l.baseDir, filepath.FromSlash(prefix))
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(l.baseDir, path)
		if err != nil {
			return err
		}
		result = append(result, ObjectInfo{
			Key:      filepath.ToSlash(rel),
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list backups: %w", err)
	}
	return result, nil
}

func (l *Local) Delete(ctx context.Context, key string) error {
	path := filepath.Join(l.baseDir, filepath.FromSlash(key))
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete backup file: %w", err)
	}
	return nil
}
