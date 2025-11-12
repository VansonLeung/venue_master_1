package services

import (
	"context"
	"errors"
	"time"
)

// ServiceClients aggregates all downstream clients used by the gateway.
type ServiceClients struct {
	Users    UserService
	Bookings BookingService
}

// UserService exposes user-domain operations needed by the gateway.
type UserService interface {
	Me(ctx context.Context, userID string) (*User, error)
}

// BookingService exposes facility + booking operations.
type BookingService interface {
	ListFacilities(ctx context.Context, venueID string) ([]*Facility, error)
	ListBookings(ctx context.Context, userID string) ([]*Booking, error)
	CreateBooking(ctx context.Context, input BookingInput) (*Booking, error)
	CancelBooking(ctx context.Context, bookingID string) (*Booking, error)
}

// User mirrors a subset of the user-service DTO.
type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Roles     []string
}

// Facility captures the minimal data required by the booking UI.
type Facility struct {
	ID          string
	VenueID     string
	Name        string
	Description string
	Surface     string
	OpenAt      time.Time
	CloseAt     time.Time
}

// Booking describes a single reservation.
type Booking struct {
	ID          string
	FacilityID  string
	UserID      string
	StartsAt    time.Time
	EndsAt      time.Time
	Status      string
	AmountCents int64
	Currency    string
}

// BookingInput is used by the createBooking mutation.
type BookingInput struct {
	FacilityID string
	UserID     string
	StartsAt   time.Time
	EndsAt     time.Time
}

// NewMockClients returns deterministic in-memory implementations so the gateway can boot before real services exist.
func NewMockClients() *ServiceClients {
	return &ServiceClients{
		Users:    &mockUserService{},
		Bookings: &mockBookingService{},
	}
}

type mockUserService struct{}

type mockBookingService struct{}

func (m *mockUserService) Me(_ context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, errors.New("missing user id")
	}
	return &User{
		ID:        userID,
		FirstName: "Venue",
		LastName:  "Member",
		Email:     "member@example.com",
		Roles:     []string{"MEMBER"},
	}, nil
}

func (m *mockBookingService) ListFacilities(_ context.Context, venueID string) ([]*Facility, error) {
	facilities := []*Facility{
		{
			ID:          "facility-1",
			VenueID:     venueID,
			Name:        "Center Court",
			Description: "Indoor pickleball court with premium lighting",
			Surface:     "hardwood",
			OpenAt:      time.Date(0, 1, 1, 6, 0, 0, 0, time.UTC),
			CloseAt:     time.Date(0, 1, 1, 23, 0, 0, 0, time.UTC),
		},
		{
			ID:          "facility-2",
			VenueID:     venueID,
			Name:        "Court B",
			Description: "Outdoor pickleball court",
			Surface:     "acrylic",
			OpenAt:      time.Date(0, 1, 1, 7, 0, 0, 0, time.UTC),
			CloseAt:     time.Date(0, 1, 1, 22, 0, 0, 0, time.UTC),
		},
	}
	return facilities, nil
}

func (m *mockBookingService) ListBookings(_ context.Context, userID string) ([]*Booking, error) {
	if userID == "" {
		return nil, errors.New("user id required")
	}

	start := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	booking := &Booking{
		ID:          "booking-1",
		FacilityID:  "facility-1",
		UserID:      userID,
		StartsAt:    start,
		EndsAt:      start.Add(90 * time.Minute),
		Status:      "CONFIRMED",
		AmountCents: 4500,
		Currency:    "CAD",
	}

	return []*Booking{booking}, nil
}

func (m *mockBookingService) CreateBooking(_ context.Context, input BookingInput) (*Booking, error) {
	if input.FacilityID == "" {
		return nil, errors.New("facility id required")
	}

	return &Booking{
		ID:          "booking-" + input.FacilityID,
		FacilityID:  input.FacilityID,
		UserID:      input.UserID,
		StartsAt:    input.StartsAt,
		EndsAt:      input.EndsAt,
		Status:      "PENDING_PAYMENT",
		AmountCents: 4500,
		Currency:    "CAD",
	}, nil
}

func (m *mockBookingService) CancelBooking(_ context.Context, bookingID string) (*Booking, error) {
	if bookingID == "" {
		return nil, errors.New("booking id required")
	}

	return &Booking{
		ID:          bookingID,
		FacilityID:  "facility-1",
		UserID:      "user-1",
		StartsAt:    time.Now().Add(24 * time.Hour),
		EndsAt:      time.Now().Add(25 * time.Hour),
		Status:      "CANCELLED",
		AmountCents: 4500,
		Currency:    "CAD",
	}, nil
}
