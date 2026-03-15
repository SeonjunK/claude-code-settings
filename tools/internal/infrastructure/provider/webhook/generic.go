// Package webhook provides a generic webhook notification provider.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
)

// GenericSender implements provider.Provider for generic webhooks.
type GenericSender struct {
	url     string
	headers map[string]string
	client  *http.Client
}

// NewGenericSender creates a generic webhook sender.
func NewGenericSender(url string, headers map[string]string) *GenericSender {
	return &GenericSender{
		url:     url,
		headers: headers,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Name returns the provider name.
func (s *GenericSender) Name() string {
	return "webhook"
}

// Send delivers the event as a JSON POST to the configured URL.
func (s *GenericSender) Send(ctx context.Context, event notification.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range s.headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
