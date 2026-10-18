package backup

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"dbbackup/src/internal/adapters"
	"dbbackup/src/internal/logutil"
	"dbbackup/src/internal/notify"
	"dbbackup/src/internal/storage"
)

type Runner struct {
	Log *logutil.RunLogger
}

func (r *Runner) Backup(ctx context.Context, adapter adapters.Adapter, mode adapters.BackupMode, store storage.Provider, conn adapters.ConnectionConfig, objectKey string) error {
	start := time.Now()
	res := logutil.Result{StartedAt: start, Status: "failed"}
	defer func() {
		res.EndedAt = time.Now()
		r.Log.Finish(res)
		_ = notify.SendCompletion(ctx, res.Status, backupMessage(res), res.Location, res.EndedAt.Sub(res.StartedAt).String())
	}()

	if err := adapters.ValidateMode(adapter, mode); err != nil {
		res.Err = err
		return err
	}
	if err := adapter.ValidateConnection(ctx, conn); err != nil {
		res.Err = err
		r.Log.Error(err.Error())
		return err
	}

	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		gz := gzip.NewWriter(pw)
		dumpErr := adapter.DumpFull(streamCtx, conn, gz)
		closeErr := gz.Close()
		closeErr2 := pw.CloseWithError(dumpErr)
		if dumpErr != nil {
			errCh <- dumpErr
			return
		}
		if closeErr != nil {
			errCh <- closeErr
			return
		}
		if closeErr2 != nil {
			errCh <- closeErr2
			return
		}
		errCh <- nil
	}()

	key := objectKey
	if key == "" {
		key = defaultObjectKey(adapter.Name(), start)
	}
	if !strings.HasSuffix(key, ".gz") {
		key += ".gz"
	}

	loc, upErr := store.Upload(streamCtx, key, pr)
	if upErr != nil {
		cancel()
		_ = pr.CloseWithError(upErr)
	}
	pipeErr := <-errCh
	if upErr != nil {
		res.Err = upErr
		r.Log.Error(upErr.Error())
		return upErr
	}
	if pipeErr != nil {
		res.Err = pipeErr
		r.Log.Error(pipeErr.Error())
		return pipeErr
	}

	res.Status = "success"
	res.Location = loc
	r.Log.Info(fmt.Sprintf("backup stored at %s", loc))
	return nil
}

func (r *Runner) Restore(ctx context.Context, adapter adapters.Adapter, store storage.Provider, conn adapters.ConnectionConfig, objectKey string, opts adapters.RestoreOptions) error {
	start := time.Now()
	res := logutil.Result{StartedAt: start, Status: "failed"}
	defer func() {
		res.EndedAt = time.Now()
		r.Log.Finish(res)
		_ = notify.SendCompletion(ctx, res.Status, "restore "+res.Status, res.Location, res.EndedAt.Sub(res.StartedAt).String())
	}()

	if objectKey == "" {
		err := fmt.Errorf("restore requires a backup key or path (--key)")
		res.Err = err
		return err
	}
	if err := adapter.ValidateConnection(ctx, conn); err != nil {
		res.Err = err
		r.Log.Error(err.Error())
		return err
	}

	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		err := store.Download(streamCtx, objectKey, pw)
		_ = pw.CloseWithError(err)
		errCh <- err
	}()

	gz, err := gzip.NewReader(pr)
	if err != nil {
		cancel()
		_ = pr.CloseWithError(err)
		<-errCh
		res.Err = err
		r.Log.Error(err.Error())
		return err
	}
	defer gz.Close()

	if err := adapter.RestoreFull(streamCtx, conn, gz, opts); err != nil {
		cancel()
		_ = pr.CloseWithError(err)
		<-errCh
		res.Err = err
		r.Log.Error(err.Error())
		return err
	}
	if dlErr := <-errCh; dlErr != nil {
		res.Err = dlErr
		r.Log.Error(dlErr.Error())
		return dlErr
	}

	res.Status = "success"
	res.Location = objectKey
	r.Log.Info("restore completed")
	return nil
}

func defaultObjectKey(dbName string, t time.Time) string {
	return path.Join(dbName, t.UTC().Format("20060102-150405")+".dump")
}

func backupMessage(res logutil.Result) string {
	if res.Err != nil {
		return "backup failed"
	}
	return "backup succeeded"
}
