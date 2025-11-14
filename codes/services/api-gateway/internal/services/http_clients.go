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

func (c *bookingHTTPClient) ListFacilities(ctx context.Context, query FacilityQuery) ([]*Facility, error) {
	params := url.Values{}
	if query.VenueID != "" {
		params.Set("venueId", query.VenueID)
	}
	if query.Available != nil {
		params.Set("available", fmt.Sprintf("%t", *query.Available))
	}
	if query.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", query.Limit))
	}
	if query.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", query.Offset))
	}
	endpoint := fmt.Sprintf("%s/v1/facilities", c.baseURL)
	if enc := params.Encode(); enc != "" {
		endpoint = fmt.Sprintf("%s?%s", endpoint, enc)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	injectAuthHeaders(ctx, req)
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

func (c *bookingHTTPClient) ListBookings(ctx context.Context, query BookingQuery) ([]*Booking, error) {
	params := url.Values{}
	if query.UserID != "" {
		params.Set("userId", query.UserID)
	}
	if query.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", query.Limit))
	}
	if query.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", query.Offset))
	}
	endpoint := fmt.Sprintf("%s/v1/bookings", c.baseURL)
	if enc := params.Encode(); enc != "" {
		endpoint = fmt.Sprintf("%s?%s", endpoint, enc)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	injectAuthHeaders(ctx, req)
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
	injectAuthHeaders(ctx, req)
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
	injectAuthHeaders(ctx, req)
	var dto bookingDTO
	if err := doJSONRequest(c.client, req, &dto); err != nil {
		return nil, err
	}
	return dto.asDomain()
}

func (c *bookingHTTPClient) GetBooking(ctx context.Context, bookingID string) (*Booking, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/bookings/%s", c.baseURL, bookingID), nil)
	if err != nil {
		return nil, err
	}
	injectAuthHeaders(ctx, req)
	var dto bookingDTO
	if err := doJSONRequest(c.client, req, &dto); err != nil {
		return nil, err
	}
	return dto.asDomain()
}

func (c *bookingHTTPClient) UpdateFacilityAvailability(ctx context.Context, facilityID string, available bool) (*Facility, error) {
	payload := map[string]bool{"available": available}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("%s/v1/facilities/%s", c.baseURL, facilityID), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	injectAuthHeaders(ctx, req)
	var dto facilityDTO
	if err := doJSONRequest(c.client, req, &dto); err != nil {
		return nil, err
	}
	facility := dto.asDomain()
	return &facility, nil
}

func (c *bookingHTTPClient) CreateFacilityOverride(ctx context.Context, input FacilityOverrideInput) (*FacilityOverride, error) {
	payload := facilityOverrideRequest{
		StartDate:       input.StartDate.Format("2006-01-02"),
		EndDate:         input.EndDate.Format("2006-01-02"),
		AllDay:          input.AllDay,
		Reason:          input.Reason,
		AppliesWeekdays: input.Weekdays,
	}
	if input.OpenAt != nil {
		payload.OpenAt = input.OpenAt.Format("15:04")
	}
	if input.CloseAt != nil {
		payload.CloseAt = input.CloseAt.Format("15:04")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("%s/v1/facilities/%s/overrides", c.baseURL, input.FacilityID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	injectAuthHeaders(ctx, req)
	var dto facilityOverrideDTO
	if err := doJSONRequest(c.client, req, &dto); err != nil {
		return nil, err
	}
	result, err := dto.asDomain()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *bookingHTTPClient) DeleteFacilityOverride(ctx context.Context, facilityID, overrideID string) error {
	endpoint := fmt.Sprintf("%s/v1/facilities/%s/overrides/%s", c.baseURL, facilityID, overrideID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	injectAuthHeaders(ctx, req)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("http DELETE %s failed: %d %s", req.URL.Path, resp.StatusCode, string(body))
	}
	return nil
}

func (c *bookingHTTPClient) GetFacilitySchedule(ctx context.Context, facilityID string, from, to time.Time) ([]*FacilityScheduleDay, error) {
	endpoint := fmt.Sprintf("%s/v1/facilities/%s/schedule?from=%s&to=%s", c.baseURL, facilityID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	injectAuthHeaders(ctx, req)
	var dto []facilityScheduleDTO
	if err := doJSONRequest(c.client, req, &dto); err != nil {
		return nil, err
	}
	result := make([]*FacilityScheduleDay, 0, len(dto))
	for _, day := range dto {
		parsed, err := day.asDomain()
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}
	return result, nil
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
	Available   bool   `json:"available"`
	WeekdayRate int    `json:"weekdayRateCents"`
	WeekendRate int    `json:"weekendRateCents"`
	Currency    string `json:"currency"`
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
		Available:   f.Available,
		WeekdayRate: f.WeekdayRate,
		WeekendRate: f.WeekendRate,
		Currency:    f.Currency,
	}
}

type bookingDTO struct {
	ID            string       `json:"id"`
	FacilityID    string       `json:"facilityId"`
	UserID        string       `json:"userId"`
	StartsAt      string       `json:"startsAt"`
	EndsAt        string       `json:"endsAt"`
	Status        string       `json:"status"`
	AmountCents   int64        `json:"amountCents"`
	Currency      string       `json:"currency"`
	PaymentIntent string       `json:"paymentIntent"`
	Facility      *facilityDTO `json:"facility"`
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
		ID:            b.ID,
		FacilityID:    b.FacilityID,
		UserID:        b.UserID,
		StartsAt:      start,
		EndsAt:        end,
		Status:        b.Status,
		AmountCents:   b.AmountCents,
		Currency:      b.Currency,
		PaymentIntent: b.PaymentIntent,
		Facility:      b.facilityDomain(),
	}, nil
}

type bookingCreateRequest struct {
	FacilityID string `json:"facilityId"`
	UserID     string `json:"userId"`
	StartsAt   string `json:"startsAt"`
	EndsAt     string `json:"endsAt"`
}

type facilityOverrideRequest struct {
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	AllDay          bool   `json:"allDay"`
	OpenAt          string `json:"openAt,omitempty"`
	CloseAt         string `json:"closeAt,omitempty"`
	Reason          string `json:"reason,omitempty"`
	AppliesWeekdays []int  `json:"appliesWeekdays,omitempty"`
}

type facilityOverrideDTO struct {
	ID              string `json:"id"`
	FacilityID      string `json:"facilityId"`
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	AllDay          bool   `json:"allDay"`
	OpenAt          string `json:"openAt"`
	CloseAt         string `json:"closeAt"`
	Reason          string `json:"reason"`
	AppliesWeekdays []int  `json:"appliesWeekdays"`
}

type facilityScheduleDTO struct {
	Date   string            `json:"date"`
	Closed bool              `json:"closed"`
	Reason string            `json:"reason"`
	Slots  []facilitySlotDTO `json:"slots"`
}

type facilitySlotDTO struct {
	OpenAt  string `json:"openAt"`
	CloseAt string `json:"closeAt"`
}

func (o facilityOverrideDTO) asDomain() (*FacilityOverride, error) {
	start, err := time.Parse("2006-01-02", o.StartDate)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", o.EndDate)
	if err != nil {
		return nil, err
	}
	var openPtr, closePtr *time.Time
	if o.OpenAt != "" {
		parsed, err := time.Parse("15:04", o.OpenAt)
		if err != nil {
			return nil, err
		}
		openPtr = &parsed
	}
	if o.CloseAt != "" {
		parsed, err := time.Parse("15:04", o.CloseAt)
		if err != nil {
			return nil, err
		}
		closePtr = &parsed
	}
	return &FacilityOverride{
		ID:         o.ID,
		FacilityID: o.FacilityID,
		StartDate:  start,
		EndDate:    end,
		AllDay:     o.AllDay,
		OpenAt:     openPtr,
		CloseAt:    closePtr,
		Reason:     o.Reason,
		Weekdays:   o.AppliesWeekdays,
	}, nil
}

func (d facilityScheduleDTO) asDomain() (*FacilityScheduleDay, error) {
	date, err := time.Parse("2006-01-02", d.Date)
	if err != nil {
		return nil, err
	}
	slots := make([]FacilitySlot, 0, len(d.Slots))
	for _, slot := range d.Slots {
		slots = append(slots, FacilitySlot{OpenAt: slot.OpenAt, CloseAt: slot.CloseAt})
	}
	return &FacilityScheduleDay{Date: date, Closed: d.Closed, Reason: d.Reason, Slots: slots}, nil
}

func (b bookingDTO) facilityDomain() *Facility {
	if b.Facility == nil {
		return nil
	}
	domain := b.Facility.asDomain()
	return &domain
}

func injectAuthHeaders(ctx context.Context, req *http.Request) {
	if meta, ok := AuthFromContext(ctx); ok {
		if meta.UserID != "" {
			req.Header.Set("X-User-ID", meta.UserID)
		}
		if len(meta.Roles) > 0 {
			req.Header.Set("X-User-Roles", strings.Join(meta.Roles, ","))
		}
	}
}
