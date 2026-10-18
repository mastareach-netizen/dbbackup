package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"dbbackup/src/internal/logutil"
)

type Payload struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Location string `json:"location,omitempty"`
	Elapsed  string `json:"elapsed,omitempty"`
}

func SendCompletion(ctx context.Context, status, message, location, elapsed string) error {
	url := os.Getenv("WEBHOOK_URL")
	if url == "" {
		return nil
	}
	body, err := json.Marshal(Payload{
		Status:   status,
		Message:  logutil.Sanitize(message),
		Location: location,
		Elapsed:  elapsed,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
