package storage_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"dbbackup/src/internal/storage"
)

func TestLocalUploadDownload(t *testing.T) {
	dir := t.TempDir()
	provider, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	payload := []byte("test-backup-bytes")
	loc, err := provider.Upload(ctx, "postgres/2024.dump.gz", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(loc); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := provider.Download(ctx, "postgres/2024.dump.gz", &out); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), payload) {
		t.Fatalf("round trip mismatch")
	}
	_ = filepath.Base(loc)
}
