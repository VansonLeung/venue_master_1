# Venue Master Platform

This directory hosts the Venue Master monorepo. It follows the architecture defined in `docs/PRD.md` and `docs/PLAN.md`:

- **Language:** Go for all backend services
- **API Front Door:** GraphQL gateway (gqlgen + gin) that fans out to domain services via REST
- **Microservices:** Eight domain services (auth, user, booking, food, parking, shop, payment, notification)
- **Shared Libraries:** Common JWT, logging, error, and storage helpers live in `lib/`
- **Infrastructure:** Docker Compose for local orchestration with PostgreSQL + Redis + supporting services

## Repository Layout

```
/codes
  ├── docker-compose.yml     # Local orchestration
  ├── go.mod                 # Root Go module for the monorepo
  ├── lib/                   # Shared Go packages referenced by every service
  ├── services/              # All microservices (each has its own entrypoint)
  ├── tools/                 # Helper scripts (lint, gen, etc.)
  └── README.md
```

Each service exposes a consistent HTTP contract (health, metrics, APIs) and consumes shared middleware from `lib/`. The gateway is the only component exposed publicly.

## Service Grid

| Service | Port | Responsibility |
| --- | --- | --- |
| api-gateway | 8080 | GraphQL front door faning out to domain REST services |
| auth-service | 8081 | Login, refresh tokens, JWT issuance |
| user-service | 8082 | Profiles, memberships, RBAC metadata |
| booking-service | 8083 | Facilities, reservations, booking lifecycle |
| food-service | 8084 | Digital menu, availability toggles |
| parking-service | 8085 | Parking slot inventory + reservations |
| shop-service | 8086 | Pro shop catalog + cart operations |
| payment-service | 8087 | Stripe abstractions: intents + refunds |
| notification-service | 8088 | Email/push/in-app notification fan-out |

## Development Workflow

1. Install Go 1.23+
2. Copy `.env.example` to `.env` and adjust secrets
3. Run `make dev` to boot the full stack via Docker Compose
4. Use `make test` to execute unit tests for all packages (target 80% coverage per PRD)

Detailed service-by-service instructions live inside each service directory. Phase 0 focuses on scaffolding; future phases will flesh out business logic, persistence, and observability.

## GraphQL Smoke Test

Once the stack is running, hit the gateway:

```bash
curl -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -d '{ "query": "{ me { id firstName } facilities(venueId: \"venue-1\") { id name } }" }'
```

By default the gateway calls the HTTP versions of the user + booking services; toggle `USE_MOCK_SERVICES=true` to revert to the in-process mocks if you are iterating purely on schema work.

### Booking flow (end-to-end)

```bash
# 1. Log in to fetch a JWT
TOKEN=$(curl -s -X POST http://localhost:8081/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"member@example.com","password":"Secret123!"}' | jq -r '.accessToken')

# 2. Query bookings + facility context
curl -s -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{ "query": "{ bookings(limit:20, offset:0) { id status facility { name available weekdayRateCents } } }" }'

# 3. Create a booking (GraphQL mutation)
curl -s -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{ "query": "mutation { createBooking(facilityId:\"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb\", startsAt:\"2026-01-01T14:00:00Z\", endsAt:\"2026-01-01T15:30:00Z\") { id status paymentIntent } }" }'
```

### Venue Management (REST via Gateway)

The platform supports full CRUD operations for venue management. All requests go through the API Gateway which proxies to the booking service.

**List Venues** (all authenticated users):
```bash
curl -X GET http://localhost:8080/v1/venues?limit=100 \
  -H "Authorization: Bearer $TOKEN"
```

**Get Venue by ID** (all authenticated users):
```bash
curl -X GET http://localhost:8080/v1/venues/<venue-id> \
  -H "Authorization: Bearer $TOKEN"
```

**Create Venue** (ADMIN/VENUE_ADMIN only):
```bash
curl -X POST http://localhost:8080/v1/venues \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Sports Complex",
    "description": "Modern sports facility",
    "address": "123 Main St",
    "city": "New York",
    "state": "NY",
    "zipCode": "10001",
    "country": "US",
    "phone": "+1-555-123-4567",
    "email": "info@venue.com",
    "website": "https://venue.com",
    "timezone": "America/New_York"
  }'
```

**Update Venue** (ADMIN/VENUE_ADMIN only):
```bash
curl -X PUT http://localhost:8080/v1/venues/<venue-id> \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Updated Sports Complex",
    "city": "San Francisco",
    "timezone": "America/Los_Angeles"
  }'
```

**Delete Venue** (ADMIN/VENUE_ADMIN only):
```bash
curl -X DELETE http://localhost:8080/v1/venues/<venue-id> \
  -H "Authorization: Bearer $TOKEN"
```

**Note:** Deleting a venue will cascade delete all associated facilities due to the foreign key constraint.

### Admin / operator helpers

- Toggle facility availability (REST):
  `curl -X PATCH http://localhost:8083/v1/facilities/<facility-id> -H 'Content-Type: application/json' -d '{"available":false}'`
- Same via GraphQL:
  `mutation { updateFacilityAvailability(id:"...", available:false) { id available } }`

### Facility schedule & overrides

- REST (booking-service):
  - `GET /v1/facilities/:id/schedule?from=2026-01-01&to=2026-01-07`
  - `POST /v1/facilities/:id/overrides`
  - `DELETE /v1/facilities/:id/overrides/:overrideId`
- GraphQL (gateway):
  ```graphql
  {
    facilitySchedule(facilityId:"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", from:"2026-01-01", to:"2026-01-07") {
      date
      closed
      slots { openAt closeAt }
    }
  }
  mutation {
    createFacilityOverride(input:{
      facilityId:"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
      startDate:"2026-01-05",
      endDate:"2026-01-05",
      allDay:false,
      openAt:"12:00",
      closeAt:"18:00",
      appliesWeekdays:[1,2,3]
    }) { id facilityId startDate endDate openAt closeAt }
  }
  ```
  Only `ADMIN`/`VENUE_ADMIN` callers can mutate overrides; `MEMBER`/`OPERATOR` have read-only access.

### Integration Tests (CI-ready)

```bash
./scripts/test-e2e.sh
```

The script sources `.env`, brings up the full docker-compose stack from `codes/`, waits for health, and runs the `test/e2e` Go tests (tag `e2e`). The suite now drives auth → booking → admin override (GraphQL) end-to-end, which protects the new schedule/override plumbing before merging.

**Comprehensive API Testing:**
```bash
./scripts/test-api.sh
```

This script tests all CRUD operations across all services including:
- Authentication (login, refresh, logout)
- User management
- Venue management (list, create, update, delete)
- Facility management with venue association
- Booking lifecycle (GraphQL and REST)
- Schedule and override management
- Food, parking, shop, payment, and notification services

The test script validates both GraphQL and REST endpoints, role-based access control, and cascade deletion behavior.

## Database Schema

### Booking Service Schema

The booking service manages venues, facilities, bookings, and schedule overrides with the following relationships:

```
venues (parent)
  ↓ (1:N, ON DELETE CASCADE)
facilities
  ↓ (1:N)
bookings

facilities
  ↓ (1:N)
facility_overrides
```

**Key Tables:**

- **venues**: Physical locations that contain facilities
  - Fields: id, name, description, address, city, state, zip_code, country, phone, email, website, timezone
  - All fields except id, name, country, and timezone are nullable

- **facilities**: Bookable resources within a venue
  - Required field: venue_id (FK → venues.id with CASCADE DELETE)
  - Deleting a venue automatically deletes all its facilities

- **bookings**: Reservations for facilities
  - Links to facilities via facility_id

- **facility_overrides**: Temporary schedule changes or blackouts
  - Defines special hours, closures, or availability rules for specific date ranges

**NULL Handling:** The venue store implementation properly handles NULL values in optional fields, converting them to empty strings in the API response.

## Default Credentials

The user-service now provisions a development member when `DEFAULT_MEMBER_EMAIL` and `DEFAULT_MEMBER_PASSWORD` are set (see `.env.example`). Pair it with the auth-service to fetch real JWTs:

```bash
curl -X POST http://localhost:8081/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"member@example.com","password":"Secret123!"}'
```

**Default Users:**
- **Member**: `member@example.com` / `Secret123!` (READ access to venues/facilities)
- **Admin**: `admin@example.com` / `Admin123!` (FULL CRUD access)

Set `USE_MOCK_SERVICES=true` if you need the gateway to fall back to in-memory mocks while iterating locally.

## Frontend Admin CMS

The platform includes a React-based Admin CMS for managing venues and facilities through a user-friendly interface.

**Location:** `../frontend_codes/admin_cms/`

**Features:**
- Full venue CRUD operations with form validation
- Facility management with venue association
- User authentication and role-based UI
- Responsive design with mobile support

**Running the Admin CMS:**
```bash
cd ../frontend_codes/admin_cms
npm install
npm run dev
```

The admin interface will be available at `http://localhost:3001` (default Vite port).

**Navigation:**
- Dashboard: Overview and metrics
- Venues: Create, edit, delete venue locations
- Facilities: Manage facilities and associate with venues
- Bookings: View and manage reservations
- Users: User management

**Note:** All admin operations (CREATE/UPDATE/DELETE for venues) require ADMIN or VENUE_ADMIN role.
