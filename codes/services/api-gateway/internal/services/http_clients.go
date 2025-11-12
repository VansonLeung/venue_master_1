package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NewHTTPClients constructs ServiceClients backed by real HTTP requests.
func NewHTTPClients(httpClient *http.Client, userBaseURL, bookingBaseURL string) *ServiceClients {
	return &ServiceClients{
		Users: &userHTTPClient{
			client:  httpClient,
			baseURL: strings.TrimRight(userBaseURL, "/"),
		},
		Bookings: &bookingHTTPClient{
			client:  httpClient,
			baseURL: strings.TrimRight(bookingBaseURL, "/"),
		},
	}
}

type userHTTPClient struct {
	client  *http.Client
	baseURL string
}

func (c *userHTTPClient) Me(ctx context.Context, userID string) (*User, error) {
	return c.fetchUser(ctx, userID)
}

func (c *userHTTPClient) fetchUser(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/users/%s", c.baseURL, userID), nil)
	if err != nil {
		return nil, err
	}
	var dto userDTO
	if err := doJSONRequest(c.client, req, &dto); err != nil {
		return nil, err
	}
	return dto.asDomain(), nil
}

type bookingHTTPClient struct {
	client  *http.Client
	baseURL string
}

func (c *bookingHTTPClient) ListFacilities(ctx context.Context, venueID string) ([]*Facility, error) {
	endpoint := fmt.Sprintf("%s/v1/facilities?venueId=%s", c.baseURL, url.QueryEscape(venueID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	var dto []facilityDTO
	if err := doJSONRequest(c.client, req, &dto); err != nil {
		return nil, err
	}
	facilities := make([]*Facility, 0, len(dto))
	for _, f := range dto {
		facility := f.asDomain()
		facilities = append(facilities, &facility)
	}
	return facilities, nil
}

func (c *bookingHTTPClient) ListBookings(ctx context.Context, userID string) ([]*Booking, error) {
	params := url.Values{}
	if userID != "" {
		params.Set("userId", userID)
	}
	endpoint := fmt.Sprintf("%s/v1/bookings?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	var dto []bookingDTO
	if err := doJSONRequest(c.client, req, &dto); err != nil {
		return nil, err
	}
	bookings := make([]*Booking, 0, len(dto))
	for _, b := range dto {
		booking, err := b.asDomain()
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, nil
}

func (c *bookingHTTPClient) CreateBooking(ctx context.Context, input BookingInput) (*Booking, error) {
	payload := bookingCreateRequest{
		FacilityID: input.FacilityID,
		UserID:     input.UserID,
		StartsAt:   input.StartsAt.Format(time.RFC3339),
		EndsAt:     input.EndsAt.Format(time.RFC3339),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/v1/bookings", c.baseURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	var dto bookingDTO
	if err := doJSONRequest(c.client, req, &dto); err != nil {
		return nil, err
	}
	return dto.asDomain()
}

func (c *bookingHTTPClient) CancelBooking(ctx context.Context, bookingID string) (*Booking, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/v1/bookings/%s", c.baseURL, bookingID), nil)
	if err != nil {
		return nil, err
	}
	var dto bookingDTO
	if err := doJSONRequest(c.client, req, &dto); err != nil {
		return nil, err
	}
	return dto.asDomain()
}

func doJSONRequest[T any](client *http.Client, req *http.Request, dest *T) error {
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("http %s %s failed: status=%d body=%s", req.Method, req.URL.Path, resp.StatusCode, string(body))
	}

	if dest == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

type userDTO struct {
	ID        string   `json:"id"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
}

func (u userDTO) asDomain() *User {
	return &User{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Roles:     u.Roles,
	}
}

type facilityDTO struct {
	ID          string `json:"id"`
	VenueID     string `json:"venueId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Surface     string `json:"surface"`
	OpenAt      string `json:"openAt"`
	CloseAt     string `json:"closeAt"`
}

func (f facilityDTO) asDomain() Facility {
	openAt, _ := time.Parse(time.RFC3339, f.OpenAt)
	closeAt, _ := time.Parse(time.RFC3339, f.CloseAt)
	return Facility{
		ID:          f.ID,
		VenueID:     f.VenueID,
		Name:        f.Name,
		Description: f.Description,
		Surface:     f.Surface,
		OpenAt:      openAt,
		CloseAt:     closeAt,
	}
}

type bookingDTO struct {
	ID          string `json:"id"`
	FacilityID  string `json:"facilityId"`
	UserID      string `json:"userId"`
	StartsAt    string `json:"startsAt"`
	EndsAt      string `json:"endsAt"`
	Status      string `json:"status"`
	AmountCents int64  `json:"amountCents"`
	Currency    string `json:"currency"`
}

func (b bookingDTO) asDomain() (*Booking, error) {
	start, err := time.Parse(time.RFC3339, b.StartsAt)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse(time.RFC3339, b.EndsAt)
	if err != nil {
		return nil, err
	}
	return &Booking{
		ID:          b.ID,
		FacilityID:  b.FacilityID,
		UserID:      b.UserID,
		StartsAt:    start,
		EndsAt:      end,
		Status:      b.Status,
		AmountCents: b.AmountCents,
		Currency:    b.Currency,
	}, nil
}

type bookingCreateRequest struct {
	FacilityID string `json:"facilityId"`
	UserID     string `json:"userId"`
	StartsAt   string `json:"startsAt"`
	EndsAt     string `json:"endsAt"`
}
