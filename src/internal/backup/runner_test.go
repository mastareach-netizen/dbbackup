package backup

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"math/rand"
	"testing"
	"time"

	"dbbackup/src/internal/adapters"
	"dbbackup/src/internal/logutil"
	"dbbackup/src/internal/storage"
)

type testAdapter struct {
	dump    func(context.Context, io.Writer) error
	restore func(context.Context, io.Reader) error
}

func (a testAdapter) Name() string { return "test" }
func (a testAdapter) ClientTools() []string { return nil }
func (a testAdapter) ValidateConnection(context.Context, adapters.ConnectionConfig) error {
	return nil
}
func (a testAdapter) SupportedBackupModes() []adapters.BackupMode {
	return []adapters.BackupMode{adapters.ModeFull}
}
func (a testAdapter) DumpFull(ctx context.Context, _ adapters.ConnectionConfig, w io.Writer) error {
	return a.dump(ctx, w)
}
func (a testAdapter) RestoreFull(ctx context.Context, _ adapters.ConnectionConfig, r io.Reader, _ adapters.RestoreOptions) error {
	return a.restore(ctx, r)
}
func (testAdapter) SelectiveRestoreSupported() bool { return false }

type testStorage struct {
	upload   func(context.Context, io.Reader) (string, error)
	download func(context.Context, io.Writer) error
}

func (testStorage) Name() string { return "test" }
func (s testStorage) Upload(ctx context.Context, _ string, r io.Reader) (string, error) {
	return s.upload(ctx, r)
}
func (s testStorage) Download(ctx context.Context, _ string, w io.Writer) error {
	return s.download(ctx, w)
}
func (testStorage) List(context.Context, string) ([]storage.ObjectInfo, error) {
	return nil, nil
}
func (testStorage) Delete(context.Context, string) error { return nil }

func TestBackupStopsDumpWhenUploadFailsEarly(t *testing.T) {
	t.Setenv("WEBHOOK_URL", "")
	uploadErr := errors.New("upload failed")
	payload := randomPayload(1 << 20)
	store := testStorage{
		upload: func(context.Context, io.Reader) (string, error) {
			return "", uploadErr
		},
	}
	adapter := testAdapter{dump: func(_ context.Context, w io.Writer) error {
		_, err := w.Write(payload)
		return err
	}}

	runWithTimeout(t, func() error {
		return (&Runner{Log: logutil.New()}).Backup(
			context.Background(), adapter, adapters.ModeFull, store,
			adapters.ConnectionConfig{}, "test.dump.gz",
		)
	}, uploadErr)
}

func TestRestoreStopsDownloadWhenRestoreFailsEarly(t *testing.T) {
	t.Setenv("WEBHOOK_URL", "")
	var compressed bytes.Buffer
	zipper := gzip.NewWriter(&compressed)
	if _, err := zipper.Write(randomPayload(1 << 20)); err != nil {
		t.Fatal(err)
	}
	if err := zipper.Close(); err != nil {
		t.Fatal(err)
	}
	downloadDone := make(chan struct{})
	store := testStorage{
		download: func(_ context.Context, w io.Writer) error {
			defer close(downloadDone)
			_, err := w.Write(compressed.Bytes())
			return err
		},
	}
	restoreErr := errors.New("restore failed")
	adapter := testAdapter{restore: func(context.Context, io.Reader) error {
		return restoreErr
	}}

	runWithTimeout(t, func() error {
		return (&Runner{Log: logutil.New()}).Restore(
			context.Background(), adapter, store, adapters.ConnectionConfig{},
			"test.dump.gz", adapters.RestoreOptions{},
		)
	}, restoreErr)

	select {
	case <-downloadDone:
	case <-time.After(2 * time.Second):
		t.Fatal("download worker remained blocked after restore failed")
	}
}

func randomPayload(size int) []byte {
	payload := make([]byte, size)
	_, _ = rand.New(rand.NewSource(1)).Read(payload)
	return payload
}

func runWithTimeout(t *testing.T, run func() error, wantErr error) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- run() }()
	select {
	case err := <-done:
		if !errors.Is(err, wantErr) {
			t.Fatalf("got error %v, want %v", err, wantErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("operation remained blocked")
	}
}
