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
	ListFacilities(ctx context.Context, query FacilityQuery) ([]*Facility, error)
	ListBookings(ctx context.Context, query BookingQuery) ([]*Booking, error)
	CreateBooking(ctx context.Context, input BookingInput) (*Booking, error)
	CancelBooking(ctx context.Context, bookingID string) (*Booking, error)
	GetBooking(ctx context.Context, bookingID string) (*Booking, error)
	UpdateFacilityAvailability(ctx context.Context, facilityID string, available bool) (*Facility, error)
	CreateFacilityOverride(ctx context.Context, input FacilityOverrideInput) (*FacilityOverride, error)
	DeleteFacilityOverride(ctx context.Context, facilityID, overrideID string) error
	GetFacilitySchedule(ctx context.Context, facilityID string, from, to time.Time) ([]*FacilityScheduleDay, error)
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
	Available   bool
	WeekdayRate int
	WeekendRate int
	Currency    string
}

type FacilityOverride struct {
	ID             string
	FacilityID     string
	StartDate      time.Time
	EndDate        time.Time
	AllDay         bool
	OpenAt         *time.Time
	CloseAt        *time.Time
	Reason         string
	Weekdays       []int
}

type FacilityScheduleDay struct {
	Date   time.Time
	Closed bool
	Reason string
	Slots  []FacilitySlot
}

type FacilitySlot struct {
	OpenAt  string
	CloseAt string
}

// Booking describes a single reservation.
type Booking struct {
	ID            string
	FacilityID    string
	UserID        string
	StartsAt      time.Time
	EndsAt        time.Time
	Status        string
	AmountCents   int64
	Currency      string
	PaymentIntent string
	Facility      *Facility
}

// BookingInput is used by the createBooking mutation.
type BookingInput struct {
	FacilityID string
	UserID     string
	StartsAt   time.Time
	EndsAt     time.Time
}

type FacilityOverrideInput struct {
	FacilityID string
	StartDate  time.Time
	EndDate    time.Time
	AllDay     bool
	OpenAt     *time.Time
	CloseAt    *time.Time
	Reason     string
	Weekdays   []int
}

// FacilityQuery carries pagination/filter filters.
type FacilityQuery struct {
	VenueID   string
	Available *bool
	Limit     int
	Offset    int
}

// BookingQuery carries pagination filters for bookings.
type BookingQuery struct {
	UserID string
	Limit  int
	Offset int
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

func (m *mockBookingService) ListFacilities(_ context.Context, query FacilityQuery) ([]*Facility, error) {
	venueID := query.VenueID
	facilities := []*Facility{
		{
			ID:          "facility-1",
			VenueID:     venueID,
			Name:        "Center Court",
			Description: "Indoor pickleball court with premium lighting",
			Surface:     "hardwood",
			OpenAt:      time.Date(0, 1, 1, 6, 0, 0, 0, time.UTC),
			CloseAt:     time.Date(0, 1, 1, 23, 0, 0, 0, time.UTC),
			Available:   true,
		},
		{
			ID:          "facility-2",
			VenueID:     venueID,
			Name:        "Court B",
			Description: "Outdoor pickleball court",
			Surface:     "acrylic",
			OpenAt:      time.Date(0, 1, 1, 7, 0, 0, 0, time.UTC),
			CloseAt:     time.Date(0, 1, 1, 22, 0, 0, 0, time.UTC),
			Available:   true,
		},
	}
	return facilities, nil
}

func (m *mockBookingService) ListBookings(_ context.Context, query BookingQuery) ([]*Booking, error) {
	userID := query.UserID
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
		Facility: &Facility{
			ID:        "facility-1",
			VenueID:   "venue-1",
			Name:      "Center Court",
			Available: true,
			OpenAt:    time.Date(0, 1, 1, 6, 0, 0, 0, time.UTC),
			CloseAt:   time.Date(0, 1, 1, 23, 0, 0, 0, time.UTC),
		},
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
		Facility: &Facility{
			ID:        input.FacilityID,
			VenueID:   "venue-1",
			Name:      "Center Court",
			Available: true,
		},
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
		Facility: &Facility{
			ID:        "facility-1",
			VenueID:   "venue-1",
			Name:      "Center Court",
			Available: true,
		},
	}, nil
}

func (m *mockBookingService) CreateFacilityOverride(_ context.Context, input FacilityOverrideInput) (*FacilityOverride, error) {
	if input.FacilityID == "" {
		return nil, errors.New("facility id required")
	}
	return &FacilityOverride{
		ID:         "override-1",
		FacilityID: input.FacilityID,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
		AllDay:     input.AllDay,
		OpenAt:     input.OpenAt,
		CloseAt:    input.CloseAt,
		Reason:     input.Reason,
		Weekdays:   input.Weekdays,
	}, nil
}


func (m *mockBookingService) DeleteFacilityOverride(_ context.Context, facilityID, overrideID string) error {
	if facilityID == "" || overrideID == "" {
		return errors.New("ids required")
	}
	return nil
}

func (m *mockBookingService) GetFacilitySchedule(_ context.Context, facilityID string, from, to time.Time) ([]*FacilityScheduleDay, error) {
	if facilityID == "" {
		return nil, errors.New("facility id required")
	}
	day := &FacilityScheduleDay{
		Date:  from,
		Slots: []FacilitySlot{{OpenAt: "06:00", CloseAt: "22:00"}},
	}
	return []*FacilityScheduleDay{day}, nil
}

func (m *mockBookingService) GetBooking(ctx context.Context, bookingID string) (*Booking, error) {
	bookings, err := m.ListBookings(ctx, BookingQuery{UserID: "user-1"})
	if err != nil || len(bookings) == 0 {
		return nil, errors.New("booking not found")
	}
	bookings[0].ID = bookingID
	return bookings[0], nil
}

func (m *mockBookingService) UpdateFacilityAvailability(_ context.Context, facilityID string, available bool) (*Facility, error) {
	return &Facility{
		ID:        facilityID,
		VenueID:   "venue-1",
		Name:      "Center Court",
		Available: available,
	}, nil
}
