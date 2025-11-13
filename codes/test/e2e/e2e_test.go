//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func TestAuthBookingFlow(t *testing.T) {
	cfg := loadConfig()
	waitFor(t, cfg.GatewayURL+"/healthz")

	token := login(t, cfg.AuthURL)

	query := `{ me { id email } }`
	resp := callGraphQL(t, cfg.GatewayURL, token, query)
	if len(resp.Errors) > 0 {
		t.Fatalf("me query errors: %+v", resp.Errors)
	}

	facilityID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	start := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Minute)
	end := start.Add(90 * time.Minute)
	mutation := fmt.Sprintf(`mutation {
        createBooking(facilityId:"%s", startsAt:"%s", endsAt:"%s") {
            id status
        }
    }`, facilityID, start.Format(time.RFC3339), end.Format(time.RFC3339))
	resp = callGraphQL(t, cfg.GatewayURL, token, mutation)
	if len(resp.Errors) > 0 {
		t.Fatalf("createBooking errors: %+v", resp.Errors)
	}
	if _, ok := resp.Data["createBooking"]; !ok {
		t.Fatalf("createBooking missing data")
	}
}

type graphQLResponse struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func loadConfig() struct {
	GatewayURL string
	AuthURL    string
} {
	gateway := os.Getenv("E2E_GATEWAY_URL")
	if gateway == "" {
		gateway = "http://localhost:8080"
	}
	auth := os.Getenv("E2E_AUTH_URL")
	if auth == "" {
		auth = "http://localhost:8081"
	}
	return struct {
		GatewayURL string
		AuthURL    string
	}{gateway, auth}
}

func waitFor(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		resp, err := httpClient.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("service not ready: %s", url)
}

func login(t *testing.T, authURL string) string {
	t.Helper()
	payload := map[string]string{
		"email":    "member@example.com",
		"password": "Secret123!",
	}
	body, _ := json.Marshal(payload)
	resp, err := httpClient.Post(authURL+"/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("login status %d: %s", resp.StatusCode, string(data))
	}
	var parsed struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if parsed.AccessToken == "" {
		t.Fatalf("access token empty")
	}
	return parsed.AccessToken
}

func callGraphQL(t *testing.T, gateway, token, query string) graphQLResponse {
	t.Helper()
	payload := map[string]string{"query": query}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, gateway+"/graphql", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("graphql request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("graphql status %d: %s", resp.StatusCode, string(data))
	}
	var parsed graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("decode graphql: %v", err)
	}
	return parsed
}
