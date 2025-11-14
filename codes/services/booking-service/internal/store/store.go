package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/venue-master/platform/lib/config"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Store manages booking persistence.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a booking store using shared config.
func New(ctx context.Context, cfg config.DatabaseConfig) (*Store, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

// Close closes the underlying pool.
func (s *Store) Close() { s.pool.Close() }

// RunMigrations executes embedded SQL migrations.
func (s *Store) RunMigrations(ctx context.Context) error {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		sqlBytes, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}
		if _, err := s.pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migration %s failed: %w", entry.Name(), err)
		}
	}
	return nil
}

// Venue represents a location with facilities.
type Venue struct {
	ID          uuid.UUID
	Name        string
	Description string
	Address     string
	City        string
	State       string
	ZipCode     string
	Country     string
	Phone       string
	Email       string
	Website     string
	Timezone    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Facility represents a bookable resource.
type Facility struct {
	ID               uuid.UUID
	VenueID          uuid.UUID
	Name             string
	Description      string
	Surface          string
	OpenAt           time.Time
	CloseAt          time.Time
	Available        bool
	WeekdayRateCents int
	WeekendRateCents int
	Currency         string
}

// Booking aggregates booking data plus facility linkage.
type Booking struct {
	ID            uuid.UUID
	FacilityID    uuid.UUID
	UserID        uuid.UUID
	StartsAt      time.Time
	EndsAt        time.Time
	Status        string
	AmountCents   int
	Currency      string
	PaymentIntent *string // nullable in database
	Facility      *Facility
}

// FacilityOverride describes temporary overrides or blackouts.
type FacilityOverride struct {
	ID             uuid.UUID
	FacilityID     uuid.UUID
	StartDate      time.Time
	EndDate        time.Time
	OpenAt         *time.Time
	CloseAt        *time.Time
	AllDay         bool
	Reason         string
	AppliesWeekday []int
}

// FacilityScheduleDay represents merged schedule output.
type FacilityScheduleDay struct {
	Date   time.Time
	Closed bool
	Reason string
	Slots  []FacilitySlot
}

// FacilitySlot represents an available window.
type FacilitySlot struct {
	OpenAt  string
	CloseAt string
}

// PaymentRetry tracks pending payment retries.
type PaymentRetry struct {
	BookingID     uuid.UUID
	Attempt       int
	NextAttemptAt time.Time
	LastError     string
}

// SeedFacility ensures there is at least one facility to book.
func (s *Store) SeedFacility(ctx context.Context, f Facility) error {
	_, err := s.pool.Exec(ctx, `
        INSERT INTO facilities (id, venue_id, name, description, surface, open_at, close_at, available)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
        ON CONFLICT (id) DO NOTHING
    `, f.ID, f.VenueID, f.Name, f.Description, f.Surface, f.OpenAt, f.CloseAt, f.Available)
	return err
}

// GetFacility returns a facility by ID.
func (s *Store) GetFacility(ctx context.Context, id uuid.UUID) (*Facility, error) {
	row := s.pool.QueryRow(ctx, `
        SELECT id, venue_id, name, description, surface, open_at, close_at, available, weekday_rate_cents, weekend_rate_cents, currency
        FROM facilities WHERE id = $1
    `, id)
	var f Facility
	if err := row.Scan(&f.ID, &f.VenueID, &f.Name, &f.Description, &f.Surface, &f.OpenAt, &f.CloseAt, &f.Available, &f.WeekdayRateCents, &f.WeekendRateCents, &f.Currency); err != nil {
		return nil, err
	}
	return &f, nil
}

// ListFacilities fetches facilities with optional availability filter.
func (s *Store) ListFacilities(ctx context.Context, venueID uuid.UUID, onlyAvailable *bool, limit, offset int) ([]Facility, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	query := `SELECT id, venue_id, name, description, surface, open_at, close_at, available, weekday_rate_cents, weekend_rate_cents, currency FROM facilities WHERE 1=1`
	args := []any{}
	idx := 1
	if venueID != uuid.Nil {
		query += fmt.Sprintf(" AND venue_id = $%d", idx)
		args = append(args, venueID)
		idx++
	}
	if onlyAvailable != nil {
		query += fmt.Sprintf(" AND available = $%d", idx)
		args = append(args, *onlyAvailable)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY name ASC LIMIT %d OFFSET %d", limit, offset)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facilities []Facility
	for rows.Next() {
		var f Facility
		if err := rows.Scan(&f.ID, &f.VenueID, &f.Name, &f.Description, &f.Surface, &f.OpenAt, &f.CloseAt, &f.Available, &f.WeekdayRateCents, &f.WeekendRateCents, &f.Currency); err != nil {
			return nil, err
		}
		facilities = append(facilities, f)
	}
	return facilities, rows.Err()
}

// UpdateFacilityAvailability toggles facility availability.
func (s *Store) UpdateFacilityAvailability(ctx context.Context, id uuid.UUID, available bool) (*Facility, error) {
	row := s.pool.QueryRow(ctx, `
        UPDATE facilities SET available=$2, updated_at=NOW()
        WHERE id=$1 RETURNING id, venue_id, name, description, surface, open_at, close_at, available, weekday_rate_cents, weekend_rate_cents, currency
    `, id, available)
	var f Facility
	if err := row.Scan(&f.ID, &f.VenueID, &f.Name, &f.Description, &f.Surface, &f.OpenAt, &f.CloseAt, &f.Available, &f.WeekdayRateCents, &f.WeekendRateCents, &f.Currency); err != nil {
		return nil, err
	}
	return &f, nil
}

// CreateFacility inserts a new facility row.
func (s *Store) CreateFacility(ctx context.Context, f Facility) (*Facility, error) {
	row := s.pool.QueryRow(ctx, `
        INSERT INTO facilities (id, venue_id, name, description, surface, open_at, close_at, available, weekday_rate_cents, weekend_rate_cents, currency)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
        RETURNING id, venue_id, name, description, surface, open_at, close_at, available, weekday_rate_cents, weekend_rate_cents, currency
    `, f.ID, f.VenueID, f.Name, f.Description, f.Surface, f.OpenAt, f.CloseAt, f.Available, f.WeekdayRateCents, f.WeekendRateCents, f.Currency)
	if err := row.Scan(&f.ID, &f.VenueID, &f.Name, &f.Description, &f.Surface, &f.OpenAt, &f.CloseAt, &f.Available, &f.WeekdayRateCents, &f.WeekendRateCents, &f.Currency); err != nil {
		return nil, err
	}
	return &f, nil
}

// ListBookings returns bookings for a user (optional) with facility data.
func (s *Store) ListBookings(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Booking, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	query := `
        SELECT b.id, b.facility_id, b.user_id, b.starts_at, b.ends_at, b.status, b.amount_cents, b.currency, b.payment_intent,
               f.id, f.venue_id, f.name, f.description, f.surface, f.open_at, f.close_at, f.available, f.weekday_rate_cents, f.weekend_rate_cents, f.currency
        FROM bookings b
        JOIN facilities f ON f.id = b.facility_id
    `
	args := []any{}
	if userID != uuid.Nil {
		query += " WHERE b.user_id = $1"
		args = append(args, userID)
	}
	query += fmt.Sprintf(" ORDER BY b.starts_at DESC LIMIT %d OFFSET %d", limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		var facility Facility
		if err := rows.Scan(&b.ID, &b.FacilityID, &b.UserID, &b.StartsAt, &b.EndsAt, &b.Status, &b.AmountCents, &b.Currency, &b.PaymentIntent,
			&facility.ID, &facility.VenueID, &facility.Name, &facility.Description, &facility.Surface, &facility.OpenAt, &facility.CloseAt, &facility.Available, &facility.WeekdayRateCents, &facility.WeekendRateCents, &facility.Currency); err != nil {
			return nil, err
		}
		b.Facility = &facility
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}

// CreateBookingInput carries booking parameters.
type CreateBookingInput struct {
	FacilityID  uuid.UUID
	UserID      uuid.UUID
	StartsAt    time.Time
	EndsAt      time.Time
	AmountCents int
	Currency    string
}

// CreateBooking inserts a booking row if no conflict exists.
func (s *Store) CreateBooking(ctx context.Context, input CreateBookingInput) (*Booking, error) {
	conflict, err := s.hasConflict(ctx, input.FacilityID, input.StartsAt, input.EndsAt)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, errors.New("facility already booked for that time range")
	}

	bookingID := uuid.New()
	row := s.pool.QueryRow(ctx, `
        INSERT INTO bookings (id, facility_id, user_id, starts_at, ends_at, status, amount_cents, currency)
        VALUES ($1,$2,$3,$4,$5,'PENDING_PAYMENT',$6,$7)
        RETURNING id, facility_id, user_id, starts_at, ends_at, status, amount_cents, currency, payment_intent
    `, bookingID, input.FacilityID, input.UserID, input.StartsAt, input.EndsAt, input.AmountCents, input.Currency)

	var b Booking
	if err := row.Scan(&b.ID, &b.FacilityID, &b.UserID, &b.StartsAt, &b.EndsAt, &b.Status, &b.AmountCents, &b.Currency, &b.PaymentIntent); err != nil {
		return nil, err
	}
	return &b, nil
}

// AttachFacility hydrates booking with facility details.
func (s *Store) AttachFacility(ctx context.Context, booking *Booking) error {
	if booking == nil {
		return nil
	}
	row := s.pool.QueryRow(ctx, `SELECT id, venue_id, name, description, surface, open_at, close_at, available, weekday_rate_cents, weekend_rate_cents, currency FROM facilities WHERE id=$1`, booking.FacilityID)
	var facility Facility
	if err := row.Scan(&facility.ID, &facility.VenueID, &facility.Name, &facility.Description, &facility.Surface, &facility.OpenAt, &facility.CloseAt, &facility.Available, &facility.WeekdayRateCents, &facility.WeekendRateCents, &facility.Currency); err != nil {
		return err
	}
	booking.Facility = &facility
	return nil
}

// UpdateBookingStatus sets status/payment info.
func (s *Store) UpdateBookingStatus(ctx context.Context, id uuid.UUID, status, paymentIntent string) (*Booking, error) {
	row := s.pool.QueryRow(ctx, `
        UPDATE bookings SET status=$2, payment_intent=$3, updated_at=NOW()
        WHERE id=$1
        RETURNING id, facility_id, user_id, starts_at, ends_at, status, amount_cents, currency, payment_intent
    `, id, status, paymentIntent)
	var b Booking
	if err := row.Scan(&b.ID, &b.FacilityID, &b.UserID, &b.StartsAt, &b.EndsAt, &b.Status, &b.AmountCents, &b.Currency, &b.PaymentIntent); err != nil {
		return nil, err
	}
	return &b, nil
}

// CancelBooking marks a booking cancelled.
func (s *Store) CancelBooking(ctx context.Context, id uuid.UUID) (*Booking, error) {
	row := s.pool.QueryRow(ctx, `
        UPDATE bookings SET status='CANCELLED', updated_at=NOW()
        WHERE id=$1
        RETURNING id, facility_id, user_id, starts_at, ends_at, status, amount_cents, currency, payment_intent
    `, id)
	var b Booking
	if err := row.Scan(&b.ID, &b.FacilityID, &b.UserID, &b.StartsAt, &b.EndsAt, &b.Status, &b.AmountCents, &b.Currency, &b.PaymentIntent); err != nil {
		return nil, err
	}
	return &b, nil
}

// GetBooking fetches a single booking by id.
func (s *Store) GetBooking(ctx context.Context, id uuid.UUID) (*Booking, error) {
	row := s.pool.QueryRow(ctx, `
        SELECT b.id, b.facility_id, b.user_id, b.starts_at, b.ends_at, b.status, b.amount_cents, b.currency, b.payment_intent,
               f.id, f.venue_id, f.name, f.description, f.surface, f.open_at, f.close_at, f.available, f.weekday_rate_cents, f.weekend_rate_cents, f.currency
        FROM bookings b
        JOIN facilities f ON f.id = b.facility_id
        WHERE b.id = $1
    `, id)
	var b Booking
	var facility Facility
	if err := row.Scan(&b.ID, &b.FacilityID, &b.UserID, &b.StartsAt, &b.EndsAt, &b.Status, &b.AmountCents, &b.Currency, &b.PaymentIntent,
		&facility.ID, &facility.VenueID, &facility.Name, &facility.Description, &facility.Surface, &facility.OpenAt, &facility.CloseAt, &facility.Available, &facility.WeekdayRateCents, &facility.WeekendRateCents, &facility.Currency); err != nil {
		return nil, err
	}
	b.Facility = &facility
	return &b, nil
}

func (s *Store) hasConflict(ctx context.Context, facilityID uuid.UUID, start, end time.Time) (bool, error) {
	row := s.pool.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM bookings
            WHERE facility_id=$1 AND status IN ('PENDING_PAYMENT','CONFIRMED')
              AND starts_at < $3 AND ends_at > $2
        )
    `, facilityID, start, end)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// CreateFacilityOverride inserts a new override entry.
func (s *Store) CreateFacilityOverride(ctx context.Context, override *FacilityOverride) (*FacilityOverride, error) {
	if override.ID == uuid.Nil {
		override.ID = uuid.New()
	}
	row := s.pool.QueryRow(ctx, `
	    INSERT INTO facility_overrides (id, facility_id, start_date, end_date, open_at, close_at, all_day, reason, applies_weekdays)
	    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	    RETURNING id, facility_id, start_date, end_date, open_at, close_at, all_day, reason, applies_weekdays
	`, override.ID, override.FacilityID, override.StartDate, override.EndDate, nullableTime(override.OpenAt), nullableTime(override.CloseAt), override.AllDay, override.Reason, intSliceToArray(override.AppliesWeekday))
	return scanOverride(row)
}

// DeleteFacilityOverride removes an override by ID.
func (s *Store) DeleteFacilityOverride(ctx context.Context, id uuid.UUID) error {
	res, err := s.pool.Exec(ctx, `DELETE FROM facility_overrides WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ListFacilityOverrides returns overrides for a facility.
func (s *Store) ListFacilityOverrides(ctx context.Context, facilityID uuid.UUID) ([]FacilityOverride, error) {
	rows, err := s.pool.Query(ctx, `
	    SELECT id, facility_id, start_date, end_date, open_at, close_at, all_day, reason, applies_weekdays
	    FROM facility_overrides
	    WHERE facility_id = $1
	    ORDER BY start_date ASC
	`, facilityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FacilityOverride
	for rows.Next() {
		over, err := scanOverride(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *over)
	}
	return items, rows.Err()
}

// GetFacilitySchedule merges base hours with overrides for the range.
func (s *Store) GetFacilitySchedule(ctx context.Context, facilityID uuid.UUID, fromDate, toDate time.Time) ([]FacilityScheduleDay, error) {
	if toDate.Before(fromDate) {
		return nil, errors.New("invalid date range")
	}
	facility, err := s.GetFacility(ctx, facilityID)
	if err != nil {
		return nil, err
	}
	overrides, err := s.fetchOverrides(ctx, facilityID, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	var days []FacilityScheduleDay
	for d := truncateDate(fromDate); !d.After(truncateDate(toDate)); d = d.AddDate(0, 0, 1) {
		day := FacilityScheduleDay{Date: d}
		override := matchOverride(overrides, d)
		if override != nil {
			day.Reason = override.Reason
			if override.AllDay || override.OpenAt == nil || override.CloseAt == nil {
				day.Closed = true
			} else {
				day.Slots = append(day.Slots, FacilitySlot{
					OpenAt:  override.OpenAt.Format("15:04"),
					CloseAt: override.CloseAt.Format("15:04"),
				})
			}
		} else {
			day.Slots = append(day.Slots, FacilitySlot{
				OpenAt:  facility.OpenAt.Format("15:04"),
				CloseAt: facility.CloseAt.Format("15:04"),
			})
		}
		if len(day.Slots) == 0 && !day.Closed {
			day.Closed = true
		}
		days = append(days, day)
	}
	return days, nil
}

func (s *Store) fetchOverrides(ctx context.Context, facilityID uuid.UUID, fromDate, toDate time.Time) ([]FacilityOverride, error) {
	rows, err := s.pool.Query(ctx, `
	    SELECT id, facility_id, start_date, end_date, open_at, close_at, all_day, reason, applies_weekdays
	    FROM facility_overrides
	    WHERE facility_id = $1 AND start_date <= $3 AND end_date >= $2
	    ORDER BY start_date ASC
	`, facilityID, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var overrides []FacilityOverride
	for rows.Next() {
		over, err := scanOverride(rows)
		if err != nil {
			return nil, err
		}
		overrides = append(overrides, *over)
	}
	return overrides, rows.Err()
}

func scanOverride(row interface{ Scan(dest ...any) error }) (*FacilityOverride, error) {
	var o FacilityOverride
	var openAt, closeAt sql.NullTime
	var weekdays []int32
	if err := row.Scan(&o.ID, &o.FacilityID, &o.StartDate, &o.EndDate, &openAt, &closeAt, &o.AllDay, &o.Reason, &weekdays); err != nil {
		return nil, err
	}
	if openAt.Valid {
		val := openAt.Time
		o.OpenAt = &val
	}
	if closeAt.Valid {
		val := closeAt.Time
		o.CloseAt = &val
	}
	for _, w := range weekdays {
		o.AppliesWeekday = append(o.AppliesWeekday, int(w))
	}
	return &o, nil
}

func matchOverride(overrides []FacilityOverride, day time.Time) *FacilityOverride {
	weekday := int(day.Weekday())
	for _, override := range overrides {
		if day.Before(truncateDate(override.StartDate)) || day.After(truncateDate(override.EndDate)) {
			continue
		}
		if len(override.AppliesWeekday) > 0 && !containsInt(override.AppliesWeekday, weekday) {
			continue
		}
		o := override
		return &o
	}
	return nil
}

func truncateDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func containsInt(values []int, target int) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t
}

func intSliceToArray(values []int) interface{} {
	if len(values) == 0 {
		return []int{0, 1, 2, 3, 4, 5, 6}
	}
	return values
}

// SchedulePaymentRetry inserts/updates a pending retry.
func (s *Store) SchedulePaymentRetry(ctx context.Context, bookingID uuid.UUID, next time.Time, attempt int, lastErr string) error {
	_, err := s.pool.Exec(ctx, `
        INSERT INTO payment_retries (booking_id, attempt, next_attempt_at, last_error)
        VALUES ($1,$2,$3,$4)
        ON CONFLICT (booking_id) DO UPDATE SET attempt=$2, next_attempt_at=$3, last_error=$4, updated_at=NOW()
    `, bookingID, attempt, next, lastErr)
	return err
}

// FetchDuePaymentRetries returns retries due for processing.
func (s *Store) FetchDuePaymentRetries(ctx context.Context, limit int) ([]PaymentRetry, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.pool.Query(ctx, `
        SELECT booking_id, attempt, next_attempt_at, last_error
        FROM payment_retries
        WHERE next_attempt_at <= NOW()
        ORDER BY next_attempt_at ASC
        LIMIT $1
    `, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var retries []PaymentRetry
	for rows.Next() {
		var pr PaymentRetry
		if err := rows.Scan(&pr.BookingID, &pr.Attempt, &pr.NextAttemptAt, &pr.LastError); err != nil {
			return nil, err
		}
		retries = append(retries, pr)
	}
	return retries, rows.Err()
}

// DeletePaymentRetry removes a retry row.
func (s *Store) DeletePaymentRetry(ctx context.Context, bookingID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM payment_retries WHERE booking_id = $1`, bookingID)
	return err
}

// === VENUE CRUD OPERATIONS ===

// ListVenues fetches all venues with pagination.
func (s *Store) ListVenues(ctx context.Context, limit, offset int) ([]Venue, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, name, description, address, city, state, zip_code, country, phone, email, website, timezone, created_at, updated_at
		FROM venues
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var venues []Venue
	for rows.Next() {
		v, err := scanVenue(rows)
		if err != nil {
			return nil, err
		}
		venues = append(venues, *v)
	}
	return venues, rows.Err()
}

// GetVenue returns a venue by ID.
func (s *Store) GetVenue(ctx context.Context, id uuid.UUID) (*Venue, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, description, address, city, state, zip_code, country, phone, email, website, timezone, created_at, updated_at
		FROM venues
		WHERE id = $1
	`, id)

	return scanVenue(row)
}

// CreateVenue inserts a new venue.
func (s *Store) CreateVenue(ctx context.Context, v Venue) (*Venue, error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if v.Country == "" {
		v.Country = "US"
	}
	if v.Timezone == "" {
		v.Timezone = "America/New_York"
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO venues (id, name, description, address, city, state, zip_code, country, phone, email, website, timezone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, name, description, address, city, state, zip_code, country, phone, email, website, timezone, created_at, updated_at
	`, v.ID, v.Name, v.Description, v.Address, v.City, v.State, v.ZipCode, v.Country, v.Phone, v.Email, v.Website, v.Timezone)

	return scanVenue(row)
}

// UpdateVenue updates venue information.
func (s *Store) UpdateVenue(ctx context.Context, id uuid.UUID, v Venue) (*Venue, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE venues
		SET name = $2, description = $3, address = $4, city = $5, state = $6, zip_code = $7, country = $8, phone = $9, email = $10, website = $11, timezone = $12, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, address, city, state, zip_code, country, phone, email, website, timezone, created_at, updated_at
	`, id, v.Name, v.Description, v.Address, v.City, v.State, v.ZipCode, v.Country, v.Phone, v.Email, v.Website, v.Timezone)

	return scanVenue(row)
}

// DeleteVenue removes a venue by ID.
func (s *Store) DeleteVenue(ctx context.Context, id uuid.UUID) error {
	res, err := s.pool.Exec(ctx, `DELETE FROM venues WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// scanVenue scans a venue row with proper NULL handling.
func scanVenue(row interface{ Scan(dest ...any) error }) (*Venue, error) {
	var v Venue
	var description, address, city, state, zipCode, phone, email, website sql.NullString

	if err := row.Scan(
		&v.ID,
		&v.Name,
		&description,
		&address,
		&city,
		&state,
		&zipCode,
		&v.Country,
		&phone,
		&email,
		&website,
		&v.Timezone,
		&v.CreatedAt,
		&v.UpdatedAt,
	); err != nil {
		return nil, err
	}

	// Convert NULL values to empty strings
	if description.Valid {
		v.Description = description.String
	}
	if address.Valid {
		v.Address = address.String
	}
	if city.Valid {
		v.City = city.String
	}
	if state.Valid {
		v.State = state.String
	}
	if zipCode.Valid {
		v.ZipCode = zipCode.String
	}
	if phone.Valid {
		v.Phone = phone.String
	}
	if email.Valid {
		v.Email = email.String
	}
	if website.Valid {
		v.Website = website.String
	}

	return &v, nil
}
