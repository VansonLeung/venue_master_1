package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client wraps calls to the payment-service.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// New returns a payment client.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Intent represents a simplified payment intent response.
type Intent struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// Charge attempts to create a payment intent.
func (c *Client) Charge(ctx context.Context, amountCents int, currency string, metadata map[string]string) (*Intent, error) {
	payload := map[string]any{
		"amountCents": amountCents,
		"currency":    currency,
		"metadata":    metadata,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/v1/payments/intents", c.baseURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("payment service returned %d", resp.StatusCode)
	}
	var intent Intent
	if err := json.NewDecoder(resp.Body).Decode(&intent); err != nil {
		return nil, err
	}
	return &intent, nil
}
