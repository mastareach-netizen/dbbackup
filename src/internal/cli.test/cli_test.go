package cli_test

import (
	"bytes"
	"testing"

	"dbbackup/src/internal/cli"
)

func TestHelpDocumentsSubcommands(t *testing.T) {
	root := cli.NewRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, word := range []string{"backup", "restore", "check", "storage", "DATABASE_URL"} {
		if !bytes.Contains([]byte(out), []byte(word)) {
			t.Fatalf("help missing %q", word)
		}
	}
}

func TestBackupRejectsIncrementalMode(t *testing.T) {
	root := cli.NewRoot()
	root.SetArgs([]string{"backup", "--mode", "incremental", "--output-dir", t.TempDir()})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for incremental mode")
	}
}
