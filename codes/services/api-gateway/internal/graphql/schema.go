package graphqlhandler

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/graphql-go/graphql"

	"github.com/venue-master/platform/services/api-gateway/internal/services"
)

type schemaBuilder struct {
	clients       *services.ServiceClients
	user          *graphql.Object
	facility      *graphql.Object
	booking       *graphql.Object
	override      *graphql.Object
	slot          *graphql.Object
	schedule      *graphql.Object
	overrideInput *graphql.InputObject
}

func buildSchema(clients *services.ServiceClients) (graphql.Schema, error) {
	builder := &schemaBuilder{clients: clients}
	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    builder.queryType(),
		Mutation: builder.mutationType(),
	})
}

func (b *schemaBuilder) queryType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"me": {
				Type:    b.userType(),
				Resolve: b.resolveMe,
			},
			"facilities": {
				Type: graphql.NewList(b.facilityType()),
				Args: graphql.FieldConfigArgument{
					"venueId":   &graphql.ArgumentConfig{Type: graphql.ID},
					"available": &graphql.ArgumentConfig{Type: graphql.Boolean},
					"limit":     &graphql.ArgumentConfig{Type: graphql.Int},
					"offset":    &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: b.resolveFacilities,
			},
			"bookings": {
				Type: graphql.NewList(b.bookingType()),
				Args: graphql.FieldConfigArgument{
					"userId": &graphql.ArgumentConfig{Type: graphql.ID},
					"limit":  &graphql.ArgumentConfig{Type: graphql.Int},
					"offset": &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: b.resolveBookings,
			},
			"booking": {
				Type: b.bookingType(),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: b.resolveBooking,
			},
			"facilitySchedule": {
				Type: graphql.NewList(b.scheduleDayType()),
				Args: graphql.FieldConfigArgument{
					"facilityId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"from":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"to":         &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: b.resolveFacilitySchedule,
			},
		},
	})
}

func (b *schemaBuilder) mutationType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createBooking": {
				Type: b.bookingType(),
				Args: graphql.FieldConfigArgument{
					"facilityId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"startsAt":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"endsAt":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: b.resolveCreateBooking,
			},
			"cancelBooking": {
				Type: b.bookingType(),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: b.resolveCancelBooking,
			},
			"updateFacilityAvailability": {
				Type: b.facilityType(),
				Args: graphql.FieldConfigArgument{
					"id":        &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"available": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Boolean)},
				},
				Resolve: b.resolveUpdateFacilityAvailability,
			},
			"createFacilityOverride": {
				Type: b.facilityOverrideType(),
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(b.facilityOverrideInput())},
				},
				Resolve: b.resolveCreateFacilityOverride,
			},
			"removeFacilityOverride": {
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"facilityId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"id":         &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: b.resolveRemoveFacilityOverride,
			},
		},
	})
}

func (b *schemaBuilder) resolveMe(p graphql.ResolveParams) (any, error) {
	claims := ClaimsFromContext(p.Context)
	if claims == nil {
		return nil, errors.New("unauthorized")
	}
	return b.clients.Users.Me(p.Context, claims.UserID)
}

func (b *schemaBuilder) resolveFacilities(p graphql.ResolveParams) (any, error) {
	venueID, _ := p.Args["venueId"].(string)
	limit, offset, err := paginationArgs(p)
	if err != nil {
		return nil, err
	}
	var availablePtr *bool
	if val, ok := p.Args["available"]; ok {
		parsed, err := boolFromArg(val)
		if err != nil {
			return nil, err
		}
		availablePtr = &parsed
	}
	query := services.FacilityQuery{
		VenueID:   venueID,
		Available: availablePtr,
		Limit:     limit,
		Offset:    offset,
	}
	return b.clients.Bookings.ListFacilities(p.Context, query)
}

func (b *schemaBuilder) resolveBookings(p graphql.ResolveParams) (any, error) {
	userID, _ := p.Args["userId"].(string)
	if userID == "" {
		claims := ClaimsFromContext(p.Context)
		if claims == nil {
			return nil, errors.New("unauthorized")
		}
		userID = claims.UserID
	}
	limit, offset, err := paginationArgs(p)
	if err != nil {
		return nil, err
	}
	query := services.BookingQuery{UserID: userID, Limit: limit, Offset: offset}
	return b.clients.Bookings.ListBookings(p.Context, query)
}

func (b *schemaBuilder) resolveBooking(p graphql.ResolveParams) (any, error) {
	id, _ := p.Args["id"].(string)
	if id == "" {
		return nil, errors.New("id is required")
	}
	return b.clients.Bookings.GetBooking(p.Context, id)
}

func (b *schemaBuilder) resolveCreateBooking(p graphql.ResolveParams) (any, error) {
	claims := ClaimsFromContext(p.Context)
	if claims == nil || claims.UserID == "" {
		return nil, errors.New("unauthorized")
	}
	facilityID, _ := p.Args["facilityId"].(string)
	if facilityID == "" {
		return nil, errors.New("facilityId is required")
	}
	startsAt, err := parseTimeArg(p.Args["startsAt"])
	if err != nil {
		return nil, err
	}
	endsAt, err := parseTimeArg(p.Args["endsAt"])
	if err != nil {
		return nil, err
	}

	input := services.BookingInput{FacilityID: facilityID, UserID: claims.UserID, StartsAt: startsAt, EndsAt: endsAt}
	return b.clients.Bookings.CreateBooking(p.Context, input)
}

func (b *schemaBuilder) resolveCancelBooking(p graphql.ResolveParams) (any, error) {
	bookingID, _ := p.Args["id"].(string)
	if bookingID == "" {
		return nil, errors.New("booking id is required")
	}
	return b.clients.Bookings.CancelBooking(p.Context, bookingID)
}

func (b *schemaBuilder) resolveUpdateFacilityAvailability(p graphql.ResolveParams) (any, error) {
	if err := ensureRoles(p, adminRoles...); err != nil {
		return nil, err
	}
	id, _ := p.Args["id"].(string)
	available, _ := p.Args["available"].(bool)
	if id == "" {
		return nil, errors.New("facility id is required")
	}
	return b.clients.Bookings.UpdateFacilityAvailability(p.Context, id, available)
}

func (b *schemaBuilder) resolveFacilitySchedule(p graphql.ResolveParams) (any, error) {
	facilityID, _ := p.Args["facilityId"].(string)
	fromStr, _ := p.Args["from"].(string)
	toStr, _ := p.Args["to"].(string)
	if facilityID == "" || fromStr == "" || toStr == "" {
		return nil, errors.New("facilityId, from, and to are required")
	}
	fromDate, err := time.Parse(dateOnlyFormat, fromStr)
	if err != nil {
		return nil, err
	}
	toDate, err := time.Parse(dateOnlyFormat, toStr)
	if err != nil {
		return nil, err
	}
	if toDate.Before(fromDate) {
		return nil, errors.New("to must be on or after from")
	}
	days, err := b.clients.Bookings.GetFacilitySchedule(p.Context, facilityID, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	return days, nil
}

func (b *schemaBuilder) resolveCreateFacilityOverride(p graphql.ResolveParams) (any, error) {
	if err := ensureRoles(p, adminRoles...); err != nil {
		return nil, err
	}
	inputRaw, _ := p.Args["input"].(map[string]any)
	parsed, err := parseOverrideInput(inputRaw)
	if err != nil {
		return nil, err
	}
	override, err := b.clients.Bookings.CreateFacilityOverride(p.Context, parsed)
	if err != nil {
		return nil, err
	}
	return override, nil
}

func (b *schemaBuilder) resolveRemoveFacilityOverride(p graphql.ResolveParams) (any, error) {
	if err := ensureRoles(p, adminRoles...); err != nil {
		return nil, err
	}
	facilityID, _ := p.Args["facilityId"].(string)
	id, _ := p.Args["id"].(string)
	if facilityID == "" || id == "" {
		return nil, errors.New("facilityId and id required")
	}
	if err := b.clients.Bookings.DeleteFacilityOverride(p.Context, facilityID, id); err != nil {
		return nil, err
	}
	return true, nil
}

func (b *schemaBuilder) userType() *graphql.Object {
	if b.user != nil {
		return b.user
	}
	b.user = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id":        {Type: graphql.NewNonNull(graphql.ID)},
			"firstName": {Type: graphql.String},
			"lastName":  {Type: graphql.String},
			"email":     {Type: graphql.String},
			"roles":     {Type: graphql.NewList(graphql.String)},
		},
	})
	return b.user
}

func (b *schemaBuilder) facilityType() *graphql.Object {
	if b.facility != nil {
		return b.facility
	}
	b.facility = graphql.NewObject(graphql.ObjectConfig{
		Name: "Facility",
		Fields: graphql.Fields{
			"id":          {Type: graphql.NewNonNull(graphql.ID)},
			"venueId":     {Type: graphql.NewNonNull(graphql.ID)},
			"name":        {Type: graphql.String},
			"description": {Type: graphql.String},
			"surface":     {Type: graphql.String},
			"openAt": {
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					if facility, ok := p.Source.(*services.Facility); ok {
						return facility.OpenAt.Format(time.RFC3339), nil
					}
					return nil, nil
				},
			},
			"closeAt": {
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					if facility, ok := p.Source.(*services.Facility); ok {
						return facility.CloseAt.Format(time.RFC3339), nil
					}
					return nil, nil
				},
			},
			"available":        {Type: graphql.Boolean},
			"weekdayRateCents": {Type: graphql.Int},
			"weekendRateCents": {Type: graphql.Int},
			"currency":         {Type: graphql.String},
		},
	})
	return b.facility
}

func (b *schemaBuilder) bookingType() *graphql.Object {
	if b.booking != nil {
		return b.booking
	}
	b.booking = graphql.NewObject(graphql.ObjectConfig{
		Name: "Booking",
		Fields: graphql.Fields{
			"id":         {Type: graphql.NewNonNull(graphql.ID)},
			"facilityId": {Type: graphql.NewNonNull(graphql.ID)},
			"userId":     {Type: graphql.NewNonNull(graphql.ID)},
			"startsAt": {
				Type:    graphql.String,
				Resolve: formatTimeField(func(b *services.Booking) time.Time { return b.StartsAt }),
			},
			"endsAt": {
				Type:    graphql.String,
				Resolve: formatTimeField(func(b *services.Booking) time.Time { return b.EndsAt }),
			},
			"status":        {Type: graphql.String},
			"amountCents":   {Type: graphql.Int},
			"currency":      {Type: graphql.String},
			"paymentIntent": {Type: graphql.String},
			"facility": {
				Type: b.facilityType(),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					if booking, ok := p.Source.(*services.Booking); ok {
						return booking.Facility, nil
					}
					return nil, nil
				},
			},
		},
	})
	return b.booking
}

func (b *schemaBuilder) facilityOverrideType() *graphql.Object {
	if b.override != nil {
		return b.override
	}
	b.override = graphql.NewObject(graphql.ObjectConfig{
		Name: "FacilityOverride",
		Fields: graphql.Fields{
			"id":         {Type: graphql.NewNonNull(graphql.ID)},
			"facilityId": {Type: graphql.NewNonNull(graphql.ID)},
			"startDate": {
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					if override, ok := overrideFromSource(p.Source); ok {
						return override.StartDate.Format(dateOnlyFormat), nil
					}
					return nil, nil
				},
			},
			"endDate": {
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					if override, ok := overrideFromSource(p.Source); ok {
						return override.EndDate.Format(dateOnlyFormat), nil
					}
					return nil, nil
				},
			},
			"allDay": {Type: graphql.Boolean},
			"openAt": {
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					if override, ok := overrideFromSource(p.Source); ok {
						return formatOptionalTime(override.OpenAt), nil
					}
					return nil, nil
				},
			},
			"closeAt": {
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					if override, ok := overrideFromSource(p.Source); ok {
						return formatOptionalTime(override.CloseAt), nil
					}
					return nil, nil
				},
			},
			"reason":          {Type: graphql.String},
			"appliesWeekdays": {Type: graphql.NewList(graphql.Int)},
		},
	})
	return b.override
}

func (b *schemaBuilder) slotType() *graphql.Object {
	if b.slot != nil {
		return b.slot
	}
	b.slot = graphql.NewObject(graphql.ObjectConfig{
		Name: "FacilitySlot",
		Fields: graphql.Fields{
			"openAt": {
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					if slot, ok := slotFromSource(p.Source); ok {
						return slot.OpenAt, nil
					}
					return nil, nil
				},
			},
			"closeAt": {
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					if slot, ok := slotFromSource(p.Source); ok {
						return slot.CloseAt, nil
					}
					return nil, nil
				},
			},
		},
	})
	return b.slot
}

func (b *schemaBuilder) scheduleDayType() *graphql.Object {
	if b.schedule != nil {
		return b.schedule
	}
	b.schedule = graphql.NewObject(graphql.ObjectConfig{
		Name: "FacilityScheduleDay",
		Fields: graphql.Fields{
			"date": {
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					if day, ok := scheduleDayFromSource(p.Source); ok {
						return day.Date.Format(dateOnlyFormat), nil
					}
					return nil, nil
				},
			},
			"closed": {Type: graphql.Boolean},
			"reason": {Type: graphql.String},
			"slots": {
				Type: graphql.NewList(b.slotType()),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					if day, ok := scheduleDayFromSource(p.Source); ok {
						return day.Slots, nil
					}
					return nil, nil
				},
			},
		},
	})
	return b.schedule
}

func (b *schemaBuilder) facilityOverrideInput() *graphql.InputObject {
	if b.overrideInput != nil {
		return b.overrideInput
	}
	b.overrideInput = graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "FacilityOverrideInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"facilityId":      {Type: graphql.NewNonNull(graphql.ID)},
			"startDate":       {Type: graphql.NewNonNull(graphql.String)},
			"endDate":         {Type: graphql.NewNonNull(graphql.String)},
			"allDay":          {Type: graphql.Boolean},
			"openAt":          {Type: graphql.String},
			"closeAt":         {Type: graphql.String},
			"reason":          {Type: graphql.String},
			"appliesWeekdays": {Type: graphql.NewList(graphql.Int)},
		},
	})
	return b.overrideInput
}

func formatTimeField(extractor func(*services.Booking) time.Time) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (any, error) {
		booking, ok := p.Source.(*services.Booking)
		if !ok {
			return nil, nil
		}
		return extractor(booking).Format(time.RFC3339), nil
	}
}

func parseTimeArg(value interface{}) (time.Time, error) {
	val, ok := value.(string)
	if !ok {
		return time.Time{}, errors.New("time value must be a string")
	}
	parsed, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}

func paginationArgs(p graphql.ResolveParams) (int, int, error) {
	limit := defaultPageSize
	offset := 0
	if raw, ok := p.Args["limit"]; ok {
		val, err := intFromArg(raw)
		if err != nil || val <= 0 {
			return 0, 0, fmt.Errorf("invalid limit")
		}
		limit = val
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}
	if raw, ok := p.Args["offset"]; ok {
		val, err := intFromArg(raw)
		if err != nil || val < 0 {
			return 0, 0, fmt.Errorf("invalid offset")
		}
		offset = val
	}
	return limit, offset, nil
}

func intFromArg(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("invalid number type")
	}
}

func boolFromArg(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1":
			return true, nil
		case "false", "0":
			return false, nil
		default:
			return false, fmt.Errorf("invalid boolean value")
		}
	default:
		return false, fmt.Errorf("invalid boolean type")
	}
}
func ensureRoles(p graphql.ResolveParams, allowed ...string) error {
	claims := ClaimsFromContext(p.Context)
	if claims == nil {
		return errors.New("unauthorized")
	}
	if len(allowed) == 0 {
		return nil
	}
	if !hasAnyRole(claims.Roles, allowed...) {
		return errors.New("forbidden")
	}
	return nil
}

func hasAnyRole(have []string, allowed ...string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, role := range have {
		for _, target := range allowed {
			if strings.EqualFold(strings.TrimSpace(role), strings.TrimSpace(target)) {
				return true
			}
		}
	}
	return false
}

func overrideFromSource(source any) (*services.FacilityOverride, bool) {
	switch v := source.(type) {
	case *services.FacilityOverride:
		return v, true
	case services.FacilityOverride:
		o := v
		return &o, true
	default:
		return nil, false
	}
}

func scheduleDayFromSource(source any) (*services.FacilityScheduleDay, bool) {
	switch v := source.(type) {
	case *services.FacilityScheduleDay:
		return v, true
	case services.FacilityScheduleDay:
		day := v
		return &day, true
	default:
		return nil, false
	}
}

func slotFromSource(source any) (*services.FacilitySlot, bool) {
	switch v := source.(type) {
	case *services.FacilitySlot:
		return v, true
	case services.FacilitySlot:
		slot := v
		return &slot, true
	default:
		return nil, false
	}
}

func formatOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(timeOnlyFormat)
}

func parseOverrideInput(raw map[string]any) (services.FacilityOverrideInput, error) {
	var input services.FacilityOverrideInput
	if raw == nil {
		return input, errors.New("input is required")
	}
	facilityID, _ := raw["facilityId"].(string)
	if facilityID == "" {
		return input, errors.New("facilityId is required")
	}
	startStr, _ := raw["startDate"].(string)
	endStr, _ := raw["endDate"].(string)
	if startStr == "" || endStr == "" {
		return input, errors.New("startDate and endDate are required")
	}
	startDate, err := time.Parse(dateOnlyFormat, startStr)
	if err != nil {
		return input, fmt.Errorf("invalid startDate: %w", err)
	}
	endDate, err := time.Parse(dateOnlyFormat, endStr)
	if err != nil {
		return input, fmt.Errorf("invalid endDate: %w", err)
	}
	if endDate.Before(startDate) {
		return input, errors.New("endDate must be on or after startDate")
	}
	allDay, _ := raw["allDay"].(bool)
	var openPtr, closePtr *time.Time
	if !allDay {
		openStr, _ := raw["openAt"].(string)
		closeStr, _ := raw["closeAt"].(string)
		if openStr == "" || closeStr == "" {
			return input, errors.New("openAt and closeAt required unless allDay is true")
		}
		openAt, err := time.Parse(timeOnlyFormat, openStr)
		if err != nil {
			return input, fmt.Errorf("invalid openAt: %w", err)
		}
		closeAt, err := time.Parse(timeOnlyFormat, closeStr)
		if err != nil {
			return input, fmt.Errorf("invalid closeAt: %w", err)
		}
		openPtr = &openAt
		closePtr = &closeAt
	}
	weekdays, err := parseWeekdays(raw["appliesWeekdays"])
	if err != nil {
		return input, err
	}
	if len(weekdays) == 0 {
		weekdays = append([]int(nil), allWeekdays...)
	}

	input = services.FacilityOverrideInput{
		FacilityID: facilityID,
		StartDate:  startDate,
		EndDate:    endDate,
		AllDay:     allDay,
		OpenAt:     openPtr,
		CloseAt:    closePtr,
		Reason:     stringValue(raw["reason"]),
		Weekdays:   weekdays,
	}
	return input, nil
}

func parseWeekdays(value any) ([]int, error) {
	if value == nil {
		return nil, nil
	}
	var rawList []interface{}
	switch v := value.(type) {
	case []interface{}:
		rawList = v
	case []int:
		out := make([]int, len(v))
		copy(out, v)
		return out, nil
	default:
		return nil, fmt.Errorf("appliesWeekdays must be a list of integers")
	}
	result := make([]int, 0, len(rawList))
	for _, item := range rawList {
		val, err := intFromArg(item)
		if err != nil {
			return nil, fmt.Errorf("appliesWeekdays must contain integers")
		}
		if val < 0 || val > 6 {
			return nil, fmt.Errorf("weekday out of range: %d", val)
		}
		result = append(result, val)
	}
	return result, nil
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return fmt.Sprint(value)
}

var adminRoles = []string{"ADMIN", "VENUE_ADMIN"}

var allWeekdays = []int{0, 1, 2, 3, 4, 5, 6}

const (
	defaultPageSize = 20
	maxPageSize     = 100
	dateOnlyFormat  = "2006-01-02"
	timeOnlyFormat  = "15:04"
)
