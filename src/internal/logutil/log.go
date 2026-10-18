package logutil

import (
	"log"
	"os"
	"time"
)

type RunLogger struct {
	logger *log.Logger
}

func New() *RunLogger {
	return &RunLogger{logger: log.New(os.Stderr, "", log.LstdFlags)}
}

func (l *RunLogger) Info(msg string) {
	l.logger.Println("INFO", msg)
}

func (l *RunLogger) Error(msg string) {
	l.logger.Println("ERROR", Sanitize(msg))
}

type Result struct {
	StartedAt time.Time
	EndedAt   time.Time
	Status    string
	Location  string
	Err       error
}

func (l *RunLogger) Finish(res Result) {
	elapsed := res.EndedAt.Sub(res.StartedAt).Round(time.Millisecond)
	if res.Err != nil {
		l.logger.Printf("DONE status=%s elapsed=%s error=%s", res.Status, elapsed, Sanitize(res.Err.Error()))
		return
	}
	l.logger.Printf("DONE status=%s elapsed=%s location=%s", res.Status, elapsed, res.Location)
}
