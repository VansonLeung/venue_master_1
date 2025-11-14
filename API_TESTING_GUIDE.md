# Venue Master API Testing Guide

## Quick Start

### 1. Run the Comprehensive Test Script

```bash
cd /Users/van/Downloads/venue_master
./scripts/test-api.sh
```

This script will test all services with full CRUD operations and provide a detailed summary.

### 2. Manual Testing with curl

#### Get an Access Token

```bash
# Login and save the token
TOKEN=$(curl -s -X POST http://localhost:8081/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d @- << 'EOF'
{"email":"member@example.com","password":"Secret123!"}
EOF
| jq -r '.accessToken')

echo "Token: ${TOKEN:0:50}..."
```

## Service Endpoints

| Service | Port | Health Check | Base URL |
|---------|------|--------------|----------|
| **API Gateway** | 8080 | http://localhost:8080/healthz | http://localhost:8080 |
| **Auth Service** | 8081 | http://localhost:8081/healthz | http://localhost:8081 |
| **User Service** | 8082 | http://localhost:8082/healthz | http://localhost:8082 |
| **Booking Service** | 8083 | http://localhost:8083/healthz | http://localhost:8083 |
| **Food Service** | 8084 | http://localhost:8084/healthz | http://localhost:8084 |
| **Parking Service** | 8085 | http://localhost:8085/healthz | http://localhost:8085 |
| **Shop Service** | 8086 | http://localhost:8086/healthz | http://localhost:8086 |
| **Payment Service** | 8087 | http://localhost:8087/healthz | http://localhost:8087 |
| **Notification Service** | 8088 | http://localhost:8088/healthz | http://localhost:8088 |

## GraphQL API (via Gateway)

**Endpoint:** `http://localhost:8080/graphql`

### Available Queries

#### 1. Get Current User (`me`)

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ me { id email firstName lastName roles } }"}' | jq
```

#### 2. List Facilities

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ facilities(limit: 10) { id name description surface available weekdayRateCents weekendRateCents currency } }"}' | jq
```

#### 3. Get Facility Schedule

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ facilitySchedule(facilityId: \"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb\", from: \"2025-11-13\", to: \"2025-11-20\") { date closed reason slots { openAt closeAt } } }"}' | jq
```

#### 4. List User Bookings

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ bookings(limit: 10) { id status startsAt endsAt amountCents currency facility { name } } }"}' | jq
```

#### 5. Get Specific Booking

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ booking(id: \"<BOOKING_ID>\") { id status startsAt endsAt amountCents currency facility { name } } }"}' | jq
```

### Available Mutations

#### 1. Create Booking

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"query":"mutation { createBooking(facilityId: \"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb\", startsAt: \"2025-11-14T10:00:00Z\", endsAt: \"2025-11-14T11:00:00Z\") { id status amountCents currency paymentIntent facility { name } } }"}' | jq
```

#### 2. Cancel Booking

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"query":"mutation { cancelBooking(id: \"<BOOKING_ID>\") { id status } }"}' | jq
```

#### 3. Update Facility Availability (Admin Only)

```bash
# Requires ADMIN or VENUE_ADMIN role
curl -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"query":"mutation { updateFacilityAvailability(id: \"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb\", available: false) { id available } }"}' | jq
```

#### 4. Create Facility Override (Admin Only)

```bash
# Requires ADMIN or VENUE_ADMIN role
curl -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"query":"mutation { createFacilityOverride(input: { facilityId: \"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb\", startDate: \"2025-11-15\", endDate: \"2025-11-15\", allDay: false, openAt: \"12:00\", closeAt: \"18:00\", reason: \"Maintenance\", appliesWeekdays: [1,2,3] }) { id facilityId startDate endDate openAt closeAt reason } }"}' | jq
```

#### 5. Remove Facility Override (Admin Only)

```bash
# Requires ADMIN or VENUE_ADMIN role
curl -X POST http://localhost:8080/graphql \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"query":"mutation { removeFacilityOverride(facilityId: \"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb\", id: \"<OVERRIDE_ID>\") }"}' | jq
```

## REST API Examples

### Authentication Service (Port 8081)

#### Login
```bash
curl -X POST http://localhost:8081/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d @- << 'EOF'
{"email":"member@example.com","password":"Secret123!"}
EOF
```

**Response:**
```json
{
  "accessToken": "eyJhbG...",
  "expiresIn": 900,
  "refreshToken": "eyJhbG...",
  "user": {
    "id": "2b5b960d-88a7-4020-8cac-84bebbfaa15e",
    "email": "member@example.com",
    "firstName": "Venue",
    "lastName": "Member",
    "roles": ["MEMBER"]
  }
}
```

#### Refresh Token
```bash
curl -X POST http://localhost:8081/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refreshToken":"<REFRESH_TOKEN>"}'
```

### User Service (Port 8082)

#### Get Current User
```bash
curl http://localhost:8082/v1/users/me \
  -H "Authorization: Bearer $TOKEN"
```

#### Get User by ID
```bash
curl http://localhost:8082/v1/users/<USER_ID> \
  -H "Authorization: Bearer $TOKEN"
```

#### Get User Memberships
```bash
curl http://localhost:8082/v1/users/<USER_ID>/memberships \
  -H "Authorization: Bearer $TOKEN"
```

### Booking Service (Port 8083)

#### List Facilities
```bash
curl "http://localhost:8083/v1/facilities?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### Get Facility by ID
```bash
curl "http://localhost:8083/v1/facilities/bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### Get Facility Schedule
```bash
curl "http://localhost:8083/v1/facilities/bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb/schedule?from=2025-11-13&to=2025-11-20" \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### Create Booking
```bash
curl -X POST http://localhost:8083/v1/bookings \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d @- << 'EOF'
{
  "facilityId": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
  "userId": "2b5b960d-88a7-4020-8cac-84bebbfaa15e",
  "startsAt": "2025-11-14T10:00:00Z",
  "endsAt": "2025-11-14T11:00:00Z"
}
EOF
```

#### List Bookings
```bash
curl "http://localhost:8083/v1/bookings?userId=<USER_ID>&limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### Get Booking by ID
```bash
curl "http://localhost:8083/v1/bookings/<BOOKING_ID>" \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### Cancel Booking
```bash
curl -X PATCH "http://localhost:8083/v1/bookings/<BOOKING_ID>/cancel" \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### Update Facility Availability (Admin)
```bash
curl -X PATCH "http://localhost:8083/v1/facilities/<FACILITY_ID>" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"available":false}' | jq
```

#### Create Facility Override (Admin)
```bash
curl -X POST "http://localhost:8083/v1/facilities/<FACILITY_ID>/overrides" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d @- << 'EOF'
{
  "startDate": "2025-11-15",
  "endDate": "2025-11-15",
  "allDay": false,
  "openAt": "12:00",
  "closeAt": "18:00",
  "reason": "Maintenance",
  "appliesWeekdays": [1, 2, 3]
}
EOF
```

#### Delete Facility Override (Admin)
```bash
curl -X DELETE "http://localhost:8083/v1/facilities/<FACILITY_ID>/overrides/<OVERRIDE_ID>" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

### Food Service (Port 8084)

#### List Menu Items
```bash
curl "http://localhost:8084/v1/menu?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq
```

### Parking Service (Port 8085)

#### List Parking Spaces
```bash
curl "http://localhost:8085/v1/parking/spaces?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### Create Parking Reservation
```bash
curl -X POST http://localhost:8085/v1/parking/reservations \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d @- << 'EOF'
{
  "spaceId": "spot-1",
  "startsAt": "2025-11-14T09:00:00Z",
  "endsAt": "2025-11-14T17:00:00Z"
}
EOF
```

### Shop Service (Port 8086)

#### List Products
```bash
curl "http://localhost:8086/v1/products?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### Add Item to Cart
```bash
curl -X POST http://localhost:8086/v1/cart/items \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"productId":"<PRODUCT_ID>","quantity":2}' | jq
```

#### Get Cart
```bash
curl http://localhost:8086/v1/cart \
  -H "Authorization: Bearer $TOKEN" | jq
```

### Payment Service (Port 8087)

#### Create Payment Intent
```bash
curl -X POST http://localhost:8087/v1/payments/intents \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d @- << 'EOF'
{
  "amount": 5000,
  "currency": "USD",
  "metadata": {
    "bookingId": "test-booking-123"
  }
}
EOF
```

#### Get Payment Intent
```bash
curl "http://localhost:8087/v1/payments/intents/<PAYMENT_INTENT_ID>" \
  -H "Authorization: Bearer $TOKEN" | jq
```

### Notification Service (Port 8088)

#### Send Notification
```bash
curl -X POST http://localhost:8088/v1/notifications \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d @- << 'EOF'
{
  "userId": "<USER_ID>",
  "type": "EMAIL",
  "subject": "Test Notification",
  "body": "This is a test notification"
}
EOF
```

#### List Notifications
```bash
curl "http://localhost:8088/v1/notifications?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq
```

## Default Test Credentials

| Email | Password | Role |
|-------|----------|------|
| member@example.com | Secret123! | MEMBER |

## Common Query Parameters

- `limit` - Number of results to return (default: 20, max: 100)
- `offset` - Number of results to skip for pagination (default: 0)
- `from` - Start date for date range queries (format: YYYY-MM-DD)
- `to` - End date for date range queries (format: YYYY-MM-DD)

## Date/Time Formats

- **Date only:** `YYYY-MM-DD` (e.g., "2025-11-13")
- **Date and time:** RFC3339 format `YYYY-MM-DDTHH:MM:SSZ` (e.g., "2025-11-14T10:00:00Z")
- **Time only:** `HH:MM` (e.g., "12:00")

## Role-Based Access Control

| Role | Description | Permissions |
|------|-------------|-------------|
| **MEMBER** | Regular users | View facilities, create/cancel own bookings, view own profile |
| **OPERATOR** | Staff members | All MEMBER permissions + manage facilities |
| **ADMIN** | Administrators | All OPERATOR permissions + manage users, create overrides |
| **VENUE_ADMIN** | Venue administrators | Same as ADMIN |

## Testing Tips

1. **Export Token for Reuse:**
   ```bash
   export TOKEN='<your-token-here>'
   ```

2. **Pretty Print JSON with jq:**
   ```bash
   curl ... | jq '.'
   ```

3. **Save Response to Variable:**
   ```bash
   RESPONSE=$(curl -s ...)
   BOOKING_ID=$(echo "$RESPONSE" | jq -r '.id')
   ```

4. **Run E2E Tests:**
   ```bash
   cd codes
   ./scripts/test-e2e.sh
   ```

5. **Check Service Logs:**
   ```bash
   docker logs codes-api-gateway-1
   docker logs codes-booking-service-1
   # etc.
   ```

## Troubleshooting

### Issue: "unauthorized" error
**Solution:** Make sure you're passing a valid JWT token in the Authorization header:
```bash
-H "Authorization: Bearer $TOKEN"
```

### Issue: Token expired
**Solution:** Login again to get a new token (tokens expire after 15 minutes):
```bash
TOKEN=$(curl -s -X POST http://localhost:8081/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"member@example.com","password":"Secret123!"}' \
  | jq -r '.accessToken')
```

### Issue: "forbidden" error
**Solution:** Your user role doesn't have permission. Some operations require ADMIN or VENUE_ADMIN roles.

### Issue: Service not responding
**Solution:** Check if Docker containers are running:
```bash
docker ps
```

## Next Steps

1. Run the comprehensive test script: `./scripts/test-api.sh`
2. Review the [codes/README.md](codes/README.md) for more details
3. Check [docs/FACILITY_FEATURES.md](docs/FACILITY_FEATURES.md) for facility scheduling features
4. Run E2E tests: `cd codes && ./scripts/test-e2e.sh`
