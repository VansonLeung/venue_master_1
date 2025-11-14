# API Configuration

The Admin CMS is configured to call different backend services on their specific ports, matching the structure in `scripts/test-api.sh`.

## Service Endpoints

Following the microservices architecture, each service runs on its own port:

| Service | Port | Purpose | Admin CMS Usage |
|---------|------|---------|-----------------|
| **Gateway** | 8080 | API Gateway, GraphQL, routing | Venues, Users |
| **Auth** | 8081 | Authentication, registration, token refresh | Login, Register, Token Refresh |
| **Booking** | 8083 | Facilities, bookings management | Facilities CRUD, Bookings CRUD |

## Configuration Files

### Environment Variables (`.env`)

```env
# Base URL for all services (without port)
VITE_BASE_URL=http://localhost

# Individual service ports
VITE_GATEWAY_PORT=8080
VITE_AUTH_PORT=8081
VITE_BOOKING_PORT=8083
```

### API Endpoints ([src/services/api.js](src/services/api.js))

```javascript
export const API_ENDPOINTS = {
  GATEWAY_URL: 'http://localhost:8080',
  AUTH_URL: 'http://localhost:8081',
  BOOKING_URL: 'http://localhost:8083',
}
```

## Service Routing

### Auth Service (Port 8081)
**File**: [src/services/auth.service.js](src/services/auth.service.js)

Routes directly to Auth service:
- `POST http://localhost:8081/v1/auth/login` - Login
- `POST http://localhost:8081/v1/auth/register` - Register
- `POST http://localhost:8081/v1/auth/refresh` - Refresh token

### Booking Service (Port 8083)
**Files**:
- [src/services/facility.service.js](src/services/facility.service.js)
- [src/services/booking.service.js](src/services/booking.service.js)

Routes directly to Booking service:
- `GET http://localhost:8083/v1/facilities` - List facilities
- `POST http://localhost:8083/v1/facilities` - Create facility
- `PUT http://localhost:8083/v1/facilities/:id` - Update facility
- `DELETE http://localhost:8083/v1/facilities/:id` - Delete facility
- `GET http://localhost:8083/v1/bookings` - List bookings
- `POST http://localhost:8083/v1/bookings` - Create booking
- `PATCH http://localhost:8083/v1/bookings/:id/cancel` - Cancel booking
- `POST http://localhost:8083/v1/bookings/:id/confirm` - Confirm booking

### Gateway (Port 8080)
**Files**:
- [src/services/venue.service.js](src/services/venue.service.js)
- [src/services/user.service.js](src/services/user.service.js)

Routes through Gateway:
- `GET http://localhost:8080/v1/venues` - List venues
- `POST http://localhost:8080/v1/venues` - Create venue
- `PUT http://localhost:8080/v1/venues/:id` - Update venue
- `DELETE http://localhost:8080/v1/venues/:id` - Delete venue
- `GET http://localhost:8080/v1/users` - List users
- `PATCH http://localhost:8080/v1/users/:id/activate` - Activate user
- `PATCH http://localhost:8080/v1/users/:id/deactivate` - Deactivate user

## Request Flow

### Authentication Flow

```
Login Request Flow:
1. User enters credentials
2. Frontend calls: http://localhost:8081/v1/auth/login
3. Auth service validates credentials
4. Returns: { accessToken, refreshToken, user }
5. Tokens stored in localStorage

Token Refresh Flow (on 401):
1. Request fails with 401 Unauthorized
2. Frontend calls: http://localhost:8081/v1/auth/refresh
3. Auth service validates refresh token
4. Returns new accessToken
5. Original request retried with new token
```

### Facilities Management Flow

```
List Facilities:
Frontend → http://localhost:8083/v1/facilities → Booking Service
           ↑ Authorization: Bearer <token>

Create Facility:
Frontend → http://localhost:8083/v1/facilities → Booking Service
           ↑ Authorization: Bearer <token>
           ↑ Body: { venueId, name, description, ... }
```

### Venues Management Flow

```
List Venues:
Frontend → http://localhost:8080/v1/venues → Gateway → Venue Service
           ↑ Authorization: Bearer <token>
```

## Authentication Headers

All authenticated requests include:
```javascript
{
  'Authorization': 'Bearer <token>',
  'Content-Type': 'application/json'
}
```

The token is automatically added by the `getAuthHeaders()` helper in each service file.

## Comparison with test-api.sh

The Admin CMS configuration exactly mirrors the test script:

**test-api.sh**:
```bash
GATEWAY_URL="$BASE_URL:8080"
AUTH_URL="$BASE_URL:8081"
BOOKING_URL="$BASE_URL:8083"

# Auth calls
curl "$AUTH_URL/v1/auth/login"

# Booking calls
curl "$BOOKING_URL/v1/facilities"

# Gateway calls
curl "$GATEWAY_URL/v1/venues"
```

**Admin CMS**:
```javascript
API_ENDPOINTS = {
  GATEWAY_URL: 'http://localhost:8080',
  AUTH_URL: 'http://localhost:8081',
  BOOKING_URL: 'http://localhost:8083',
}

// Auth calls
axios.post(`${API_ENDPOINTS.AUTH_URL}/v1/auth/login`)

// Booking calls
axios.get(`${API_ENDPOINTS.BOOKING_URL}/v1/facilities`)

// Gateway calls
axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/venues`)
```

## Why Direct Service Calls?

The Admin CMS calls services directly (not through the gateway) for:

1. **Performance**: Eliminates gateway hop for auth and booking operations
2. **Clarity**: Explicit routing makes debugging easier
3. **Flexibility**: Can route to different services as needed
4. **Consistency**: Matches the test-api.sh script structure

## CORS Configuration

For this to work, each service must allow CORS requests from `http://localhost:3001`:

```go
// Example CORS config (backend)
router.Use(cors.New(cors.Config{
    AllowOrigins: []string{
        "http://localhost:3001", // Admin CMS
    },
    AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
    AllowHeaders: []string{"Authorization", "Content-Type"},
}))
```

## Troubleshooting

### Cannot connect to Auth service

**Check**: Is auth service running on port 8081?
```bash
curl http://localhost:8081/healthz
```

**Solution**: Ensure docker-compose has started all services
```bash
docker-compose ps
docker-compose up -d
```

### Cannot connect to Booking service

**Check**: Is booking service running on port 8083?
```bash
curl http://localhost:8083/healthz
```

### CORS errors

**Check**: Browser console for CORS errors

**Solution**: Update backend CORS configuration to allow `http://localhost:3001`

### Token refresh fails

**Check**: Is refresh endpoint using correct port?
- Should be: `http://localhost:8081/v1/auth/refresh`
- Not: `http://localhost:8080/v1/auth/refresh`

**Verify in**: [src/services/api.js](src/services/api.js) line 51

## Testing API Configuration

### Test Auth Service
```bash
curl -X POST http://localhost:8081/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"Secret123!"}'
```

### Test Booking Service
```bash
# First get a token
TOKEN="your-token-here"

curl http://localhost:8083/v1/facilities \
  -H "Authorization: Bearer $TOKEN"
```

### Test Gateway
```bash
curl http://localhost:8080/v1/venues \
  -H "Authorization: Bearer $TOKEN"
```

## Production Configuration

For production, update `.env`:

```env
VITE_BASE_URL=https://api.yourdomain.com

# Production uses standard ports (443/HTTPS)
# Services behind load balancer/reverse proxy
VITE_GATEWAY_PORT=443
VITE_AUTH_PORT=443
VITE_BOOKING_PORT=443
```

Or use different subdomains:
```env
VITE_BASE_URL=https://yourdomain.com
VITE_GATEWAY_PORT=   # gateway.yourdomain.com
VITE_AUTH_PORT=      # auth.yourdomain.com
VITE_BOOKING_PORT=   # booking.yourdomain.com
```

## Summary

✅ **Auth Service (8081)**: Login, Register, Token Refresh
✅ **Booking Service (8083)**: Facilities & Bookings CRUD
✅ **Gateway (8080)**: Venues & Users management

All services are called directly on their specific ports, matching the `test-api.sh` script structure for consistency and clarity.
