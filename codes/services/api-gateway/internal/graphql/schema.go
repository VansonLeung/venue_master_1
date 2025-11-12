package graphqlhandler

import (
	"errors"
	"time"

	"github.com/graphql-go/graphql"

	"github.com/venue-master/platform/services/api-gateway/internal/services"
)

type schemaBuilder struct {
	clients  *services.ServiceClients
	user     *graphql.Object
	facility *graphql.Object
	booking  *graphql.Object
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
					"venueId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: b.resolveFacilities,
			},
			"bookings": {
				Type: graphql.NewList(b.bookingType()),
				Args: graphql.FieldConfigArgument{
					"userId": &graphql.ArgumentConfig{Type: graphql.ID},
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
	if venueID == "" {
		return nil, errors.New("venueId is required")
	}
	return b.clients.Bookings.ListFacilities(p.Context, venueID)
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
	return b.clients.Bookings.ListBookings(p.Context, userID)
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
	id, _ := p.Args["id"].(string)
	available, _ := p.Args["available"].(bool)
	if id == "" {
		return nil, errors.New("facility id is required")
	}
	return b.clients.Bookings.UpdateFacilityAvailability(p.Context, id, available)
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
