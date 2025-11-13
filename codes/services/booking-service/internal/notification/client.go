package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client sends notifications to the notification-service API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a notification client.
func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// NotifyPayload describes the payload sent to notification-service.
type NotifyPayload struct {
	UserID  string `json:"userId"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Channel string `json:"channel"`
}

// Send dispatches a notification.
func (c *Client) Send(ctx context.Context, payload NotifyPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/v1/notifications", c.baseURL), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("notification service responded with %d", resp.StatusCode)
	}
	return nil
}
