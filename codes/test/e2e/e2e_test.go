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
	"strings"
	"testing"
	"time"

	"github.com/venue-master/platform/lib/config"
	"github.com/venue-master/platform/lib/jwtutil"
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
	var me struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	decodeData(t, resp.Data, "me", &me)
	if me.ID == "" || me.Email == "" {
		t.Fatalf("me payload missing fields: %+v", me)
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
	var createdBooking struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	decodeData(t, resp.Data, "createBooking", &createdBooking)
	if createdBooking.ID == "" {
		t.Fatalf("createBooking missing id")
	}

	adminToken := generateAdminToken(t)
	overrideDate := time.Now().Add(48 * time.Hour).UTC()
	overrideStart := overrideDate.Format("2006-01-02")
	overrideWeekday := int(overrideDate.Weekday())
	overrideMutation := fmt.Sprintf(`mutation {
      createFacilityOverride(input:{
        facilityId:"%s",
        startDate:"%s",
        endDate:"%s",
        allDay:false,
        openAt:"12:00",
        closeAt:"18:00",
        appliesWeekdays:[%d],
        reason:"e2e"
      }) { id facilityId startDate }
    }`, facilityID, overrideStart, overrideStart, overrideWeekday)
	overrideResp := callGraphQL(t, cfg.GatewayURL, adminToken, overrideMutation)
	if len(overrideResp.Errors) > 0 {
		t.Fatalf("createFacilityOverride errors: %+v", overrideResp.Errors)
	}
	var override struct {
		ID         string `json:"id"`
		FacilityID string `json:"facilityId"`
	}
	decodeData(t, overrideResp.Data, "createFacilityOverride", &override)
	if override.ID == "" || override.FacilityID != facilityID {
		t.Fatalf("override payload invalid: %+v", override)
	}

	t.Cleanup(func() {
		if override.ID == "" {
			return
		}
		removeMutation := fmt.Sprintf(`mutation { removeFacilityOverride(facilityId:"%s", id:"%s") }`, facilityID, override.ID)
		resp := callGraphQL(t, cfg.GatewayURL, adminToken, removeMutation)
		if len(resp.Errors) > 0 {
			t.Logf("cleanup override failed: %+v", resp.Errors)
		}
	})

	scheduleQuery := fmt.Sprintf(`{
      facilitySchedule(facilityId:"%s", from:"%s", to:"%s") {
        date
        closed
        slots { openAt closeAt }
      }
    }`, facilityID, overrideStart, overrideStart)
	scheduleResp := callGraphQL(t, cfg.GatewayURL, token, scheduleQuery)
	if len(scheduleResp.Errors) > 0 {
		t.Fatalf("facilitySchedule errors: %+v", scheduleResp.Errors)
	}
	var schedule []struct {
		Date  string `json:"date"`
		Slots []struct {
			OpenAt  string `json:"openAt"`
			CloseAt string `json:"closeAt"`
		} `json:"slots"`
	}
	decodeData(t, scheduleResp.Data, "facilitySchedule", &schedule)
	if len(schedule) == 0 || len(schedule[0].Slots) == 0 {
		t.Fatalf("schedule missing slots: %+v", schedule)
	}
	slot := schedule[0].Slots[0]
	if slot.OpenAt != "12:00" || slot.CloseAt != "18:00" {
		t.Fatalf("schedule slot mismatch: %+v", slot)
	}

	removeMutation := fmt.Sprintf(`mutation { removeFacilityOverride(facilityId:"%s", id:"%s") }`, facilityID, override.ID)
	removeResp := callGraphQL(t, cfg.GatewayURL, adminToken, removeMutation)
	if len(removeResp.Errors) > 0 {
		t.Fatalf("removeFacilityOverride errors: %+v", removeResp.Errors)
	}
	var removed bool
	decodeData(t, removeResp.Data, "removeFacilityOverride", &removed)
	if !removed {
		t.Fatalf("removeFacilityOverride returned false")
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
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
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

func decodeData[T any](t *testing.T, data map[string]json.RawMessage, key string, dest *T) {
	t.Helper()
	raw, ok := data[key]
	if !ok {
		t.Fatalf("graphql field %s missing", key)
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		t.Fatalf("decode %s: %v", key, err)
	}
}

func generateAdminToken(t *testing.T) string {
	t.Helper()
	cfg := jwtConfigFromEnv()
	manager := jwtutil.NewManager(cfg)
	access, _, err := manager.Generate("11111111-2222-3333-4444-555555555555", []string{"ADMIN", "VENUE_ADMIN"}, nil)
	if err != nil {
		t.Fatalf("generate admin token: %v", err)
	}
	return access
}

func jwtConfigFromEnv() config.JWTConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "change-this-in-production"
	}
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "venue-master"
	}
	audience := os.Getenv("JWT_AUDIENCE")
	if audience == "" {
		audience = "venue-master-clients"
	}
	return config.JWTConfig{
		Secret:        secret,
		Issuer:        issuer,
		Audience:      audience,
		AccessExpiry:  time.Hour,
		RefreshExpiry: 24 * time.Hour,
	}
}
