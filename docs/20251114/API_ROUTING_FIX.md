# API Routing Fix - Frontend to Gateway

## Issue

The Admin CMS was calling the booking service directly on port 8083, which caused authentication errors:

```
{
    "error": "missing auth headers: X-User-ID = '', X-User-Roles = ''"
}
```

### Root Cause

The booking service's authentication middleware expects two headers:
- `X-User-ID` - The authenticated user's ID
- `X-User-Roles` - The user's roles (comma-separated)

These headers are **not sent by the frontend** - they are added by the **API Gateway** after validating the JWT token.

### Original (Incorrect) Flow

```
Frontend → http://localhost:8083/v1/facilities (Booking Service)
           ❌ Missing X-User-ID and X-User-Roles headers
```

The frontend sent:
- ✅ `Authorization: Bearer <JWT>`
- ❌ No `X-User-ID`
- ❌ No `X-User-Roles`

## Solution

Route all facility and booking requests through the API Gateway (port 8080), which:
1. Validates the JWT token
2. Extracts user ID and roles from the token
3. Adds `X-User-ID` and `X-User-Roles` headers
4. Forwards the request to the booking service

### Correct Flow

```
Frontend → http://localhost:8080/v1/facilities (API Gateway)
           ✅ Sends: Authorization: Bearer <JWT>
                    ↓
           Gateway validates JWT and adds headers
                    ↓
           → http://booking-service:8080/v1/facilities (Booking Service)
             ✅ Receives: Authorization: Bearer <JWT>
                          X-User-ID: 53fd4d40-e98d-4b69-9d51-862af4dfde40
                          X-User-Roles: MEMBER
```

## Changes Made

### 1. Facility Service ([src/services/facility.service.js](frontend_codes/admin_cms/src/services/facility.service.js))

**Before:**
```javascript
async getFacilities(params = {}) {
  const response = await axios.get(`${API_ENDPOINTS.BOOKING_URL}/v1/facilities`, {
    // Directly called port 8083
  })
}
```

**After:**
```javascript
async getFacilities(params = {}) {
  const response = await axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/facilities`, {
    // Now goes through Gateway (port 8080)
  })
}
```

**All facility endpoints updated:**
- `GET /v1/facilities` - List facilities
- `GET /v1/facilities/:id` - Get facility by ID
- `POST /v1/facilities` - Create facility
- `PUT /v1/facilities/:id` - Update facility
- `DELETE /v1/facilities/:id` - Delete facility
- `GET /v1/facilities/:id/schedule` - Get facility schedule

### 2. Booking Service ([src/services/booking.service.js](frontend_codes/admin_cms/src/services/booking.service.js))

**Before:**
```javascript
async getBookings(params = {}) {
  const response = await axios.get(`${API_ENDPOINTS.BOOKING_URL}/v1/bookings`, {
    // Directly called port 8083
  })
}
```

**After:**
```javascript
async getBookings(params = {}) {
  const response = await axios.get(`${API_ENDPOINTS.GATEWAY_URL}/v1/bookings`, {
    // Now goes through Gateway (port 8080)
  })
}
```

**All booking endpoints updated:**
- `GET /v1/bookings` - List bookings
- `GET /v1/bookings/:id` - Get booking by ID
- `POST /v1/bookings` - Create booking
- `PATCH /v1/bookings/:id/status` - Update booking status
- `PATCH /v1/bookings/:id/cancel` - Cancel booking
- `POST /v1/bookings/:id/confirm` - Confirm booking
- `GET /v1/bookings/stats` - Get booking statistics

## API Gateway's Role

The API Gateway ([codes/services/api-gateway/cmd/gateway/main.go](codes/services/api-gateway/cmd/gateway/main.go)) performs these critical functions:

1. **JWT Validation**: Validates the Bearer token
2. **User Extraction**: Extracts user ID and roles from JWT claims
3. **Header Injection**: Adds `X-User-ID` and `X-User-Roles` headers
4. **Request Forwarding**: Proxies the request to the appropriate backend service

```go
// Gateway extracts from JWT
claims := validateJWT(token)

// Gateway adds these headers before forwarding
headers := map[string]string{
    "X-User-ID":    claims.UserID,      // e.g., "53fd4d40-..."
    "X-User-Roles": claims.Roles.Join() // e.g., "MEMBER"
}

// Forward to booking service with added headers
proxyToBookingService(request, headers)
```

## Booking Service Middleware

The booking service middleware ([codes/services/booking-service/internal/middleware/auth.go](codes/services/booking-service/internal/middleware/auth.go)) validates these headers:

```go
func RequireAuth() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        userID := ctx.GetHeader("X-User-ID")
        rolesHeader := ctx.GetHeader("X-User-Roles")

        if userID == "" || rolesHeader == "" {
            ctx.AbortWithStatusJSON(http.StatusUnauthorized,
                gin.H{"error": "missing auth headers"})
            return
        }

        // Parse roles and set context
        roles := strings.Split(rolesHeader, ",")
        ctx.Set("authUser", ContextUser{UserID: userID, Roles: roles})
        ctx.Next()
    }
}
```

## Port Usage Summary

Updated port routing for Admin CMS:

| Service | Port | Used For | Route Through |
|---------|------|----------|---------------|
| **API Gateway** | 8080 | ✅ Facilities, Bookings, Users | Direct |
| **Auth Service** | 8081 | ✅ Login, Register, Refresh | Direct |
| **User Service** | 8082 | ❌ Not used directly | - |
| **Booking Service** | 8083 | ❌ Not used directly (Gateway forwards) | - |

## Environment Variables

No changes needed to `.env` file. The configuration remains:

```env
VITE_BASE_URL=http://localhost
VITE_GATEWAY_PORT=8080
VITE_AUTH_PORT=8081
VITE_BOOKING_PORT=8083  # Not used directly anymore
```

## Testing

### Before Fix

```bash
# Direct call to booking service - FAILS
curl -X POST http://localhost:8083/v1/facilities \
  -H 'Authorization: Bearer eyJhbGci...' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Test Facility",...}'

# Response: ❌
{
  "error": "missing auth headers: X-User-ID = '', X-User-Roles = ''"
}
```

### After Fix

```bash
# Call through API Gateway - WORKS
curl -X POST http://localhost:8080/v1/facilities \
  -H 'Authorization: Bearer eyJhbGci...' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Test Facility",...}'

# Response: ✅
{
  "id": "...",
  "name": "Test Facility",
  ...
}
```

## Why This Architecture?

### Benefits of Gateway Pattern

1. **Centralized Authentication**: JWT validation happens once at the gateway
2. **Service Isolation**: Backend services don't need JWT libraries
3. **Header Injection**: User context propagates automatically
4. **Single Entry Point**: Easier to add rate limiting, logging, etc.
5. **Security**: Backend services are not exposed directly

### Alternative Approach (Not Recommended)

You could modify the booking service to accept JWT tokens directly, but this would:
- Duplicate JWT validation logic across services
- Require all services to have JWT secret keys
- Make services tightly coupled to auth mechanism
- Lose the benefits of the gateway pattern

## Summary

**Problem**: Frontend called booking service directly, which expects headers that only the gateway provides.

**Solution**: Route all facility and booking requests through the API Gateway (port 8080).

**Files Modified**:
1. [frontend_codes/admin_cms/src/services/facility.service.js](frontend_codes/admin_cms/src/services/facility.service.js) - Changed `BOOKING_URL` to `GATEWAY_URL`
2. [frontend_codes/admin_cms/src/services/booking.service.js](frontend_codes/admin_cms/src/services/booking.service.js) - Changed `BOOKING_URL` to `GATEWAY_URL`

**Result**: All API calls now work correctly with proper authentication and authorization.
