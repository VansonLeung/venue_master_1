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

### Admin / operator helpers

- Toggle facility availability (REST):  
  `curl -X PATCH http://localhost:8083/v1/facilities/<facility-id> -H 'Content-Type: application/json' -d '{"available":false}'`
- Same via GraphQL:  
  `mutation { updateFacilityAvailability(id:"...", available:false) { id available } }`

### Integration Tests (CI-ready)

```bash
./scripts/test-e2e.sh
```

The script brings up the full docker-compose stack, waits for health, and runs the `test/e2e` Go tests (tag `e2e`) that perform auth → booking creation → GraphQL verification. Use this in CI to guard the happy path end-to-end.

## Default Credentials

The user-service now provisions a development member when `DEFAULT_MEMBER_EMAIL` and `DEFAULT_MEMBER_PASSWORD` are set (see `.env.example`). Pair it with the auth-service to fetch real JWTs:

```bash
curl -X POST http://localhost:8081/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"member@example.com","password":"Secret123!"}'
```

Set `USE_MOCK_SERVICES=true` if you need the gateway to fall back to in-memory mocks while iterating locally.
