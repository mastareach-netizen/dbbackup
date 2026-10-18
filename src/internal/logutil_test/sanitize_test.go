package logutil_test

import (
	"strings"
	"testing"

	"dbbackup/src/internal/logutil"
)

func TestSanitizeRedactsPasswordsAndURLs(t *testing.T) {
	in := "connect postgres://user:secret@db.example.com:5432/app password=hunter2 token=abc"
	out := logutil.Sanitize(in)
	if out == in {
		t.Fatal("expected sanitization to change output")
	}
	for _, secret := range []string{"secret", "hunter2", "abc"} {
		if strings.Contains(out, secret) {
			t.Fatalf("sanitized output still contains %q: %s", secret, out)
		}
	}
}
