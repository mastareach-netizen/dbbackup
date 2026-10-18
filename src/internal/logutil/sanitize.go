package logutil

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	passwordKV = regexp.MustCompile(`(?i)(password|passwd|pwd|token|secret|api[_-]?key)\s*[=:]\s*\S+`)
	pgURL      = regexp.MustCompile(`postgres(?:ql)?://[^\s]+`)
)

func Sanitize(s string) string {
	out := s
	out = pgURL.ReplaceAllString(out, "postgres://***")
	out = passwordKV.ReplaceAllString(out, "$1=***")
	if u, err := url.Parse(s); err == nil && u.User != nil {
		if _, hasPass := u.User.Password(); hasPass {
			out = strings.ReplaceAll(out, u.User.String()+"@", "***@")
		}
	}
	return out
}
