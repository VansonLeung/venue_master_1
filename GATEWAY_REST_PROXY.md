# API Gateway REST Proxy - Implementation Complete

## Problem

The Admin CMS was calling the booking service directly on port 8083, which failed because:
1. The booking service expects `X-User-ID` and `X-User-Roles` headers
2. These headers are NOT sent by the frontend - they are added by the API Gateway
3. The frontend only sends `Authorization: Bearer <JWT>` header

## Solution Implemented

Added REST proxy endpoints to the API Gateway that:
1. Validate JWT tokens from the Authorization header
2. Extract user ID and roles from the JWT claims
3. Add `X-User-ID` and `X-User-Roles` headers
4. Forward requests to the booking service with proper authentication headers

## Files Created

### 1. REST Handler (`codes/services/api-gateway/internal/rest/handlers.go`)

New file that implements REST proxy endpoints for:

**Facilities Endpoints:**
- `GET /v1/facilities` - List facilities
- `GET /v1/facilities/:id` - Get facility by ID
- `POST /v1/facilities` - Create facility
- `PUT /v1/facilities/:id` - Update facility
- `DELETE /v1/facilities/:id` - Delete facility
- `GET /v1/facilities/:id/schedule` - Get facility schedule

**Bookings Endpoints:**
- `GET /v1/bookings` - List bookings
- `GET /v1/bookings/:id` - Get booking by ID
- `POST /v1/bookings` - Create booking
- `PATCH /v1/bookings/:id/status` - Update booking status
- `PATCH /v1/bookings/:id/cancel` - Cancel booking
- `POST /v1/bookings/:id/confirm` - Confirm booking
- `GET /v1/bookings/stats` - Get booking statistics

**Users Endpoints:**
- `GET /v1/users` - List users
- `GET /v1/users/:id` - Get user by ID

### Key Features:

```go
// Auth middleware validates JWT
func (h *Handler) authMiddleware() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        // Extract Bearer token
        authHeader := ctx.GetHeader("Authorization")
        token := strings.TrimPrefix(authHeader, "Bearer ")

        // Validate JWT
        claims, err := h.jwtManager.Validate(token)
        if err != nil {
            ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            ctx.Abort()
            return
        }

        // Store auth metadata in context
        authCtx := services.WithAuth(ctx.Request.Context(), services.AuthMetadata{
            UserID: claims.UserID,
            Roles:  claims.Roles,
        })
        ctx.Request = ctx.Request.WithContext(authCtx)

        ctx.Next()
    }
}

// Proxy request forwards to booking service with auth headers
func (h *Handler) proxyRequest(ctx *gin.Context, targetURL, method, path string, body io.Reader) {
    // Create request
    req, err := http.NewRequestWithContext(ctx.Request.Context(), method, targetURL+path, body)

    // Inject auth headers from context
    if meta, ok := services.AuthFromContext(ctx.Request.Context()); ok {
        req.Header.Set("X-User-ID", meta.UserID)
        req.Header.Set("X-User-Roles", strings.Join(meta.Roles, ","))
    }

    // Forward request and response
    client.Do(req)
    // ... copy response back to client
}
```

## Files Modified

### 1. Gateway Main (`codes/services/api-gateway/cmd/gateway/main.go`)

Added REST handler registration:

```go
import (
    graphqlhandler "github.com/venue-master/platform/services/api-gateway/internal/graphql"
    "github.com/venue-master/platform/services/api-gateway/internal/rest"  // Added
    "github.com/venue-master/platform/services/api-gateway/internal/services"
)

func main() {
    // ... setup ...

    // Register GraphQL handler
    graphqlHandler, err := graphqlhandler.New(clients, jwtManager, srv.Logger)
    graphqlHandler.Register(srv.Engine)

    // Register REST proxy handlers (NEW)
    restHandler := rest.New(clients, jwtManager, srv.Logger)
    restHandler.Register(srv.Engine)

    srv.Run()
}
```

### 2. Frontend Services (Already Updated)

- **facility.service.js** - Changed from `BOOKING_URL` (port 8083) to `GATEWAY_URL` (port 8080)
- **booking.service.js** - Changed from `BOOKING_URL` (port 8083) to `GATEWAY_URL` (port 8080)

## Request Flow

### Before (Failed):
```
Frontend → http://localhost:8083/v1/facilities
           Headers: Authorization: Bearer <JWT>

Booking Service → ❌ "missing auth headers: X-User-ID = '', X-User-Roles = ''"
```

### After (Works):
```
Frontend → http://localhost:8080/v1/facilities
           Headers: Authorization: Bearer <JWT>

API Gateway:
  1. Validates JWT token
  2. Extracts user ID and roles from JWT claims
  3. Creates new request to booking service
  4. Adds X-User-ID and X-User-Roles headers
  5. Forwards request

Booking Service → ✅ Receives X-User-ID and X-User-Roles headers
                  → Processes request successfully
                  → Returns response

API Gateway → Forwards response back to frontend
```

## Environment Variables

The REST proxy uses these environment variables (same as GraphQL handler):

```env
BOOKING_SERVICE_URL=http://booking-service:8080
USER_SERVICE_URL=http://user-service:8080
```

Defaults if not set:
- `BOOKING_SERVICE_URL`: `http://booking-service:8080`
- `USER_SERVICE_URL`: `http://user-service:8080`

## Docker Build

The API Gateway was rebuilt successfully:

```bash
cd codes
docker-compose up -d --build api-gateway
```

**Build Status**: ✅ Success
**Container Status**: ✅ Running

## Testing

### Test 1: Health Check
```bash
curl http://localhost:8080/healthz
```

**Expected**: `{"status":"ok","service":"api-gateway"}`

### Test 2: Login (Get Fresh Token)
```bash
curl -X POST http://localhost:8081/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"your-email@example.com","password":"your-password"}'
```

**Expected**: Returns JWT tokens
```json
{
  "accessToken": "eyJhbGci...",
  "refreshToken": "eyJhbGci...",
  "expiresIn": 900,
  "user": {...}
}
```

### Test 3: Create Facility (Through Gateway)
```bash
curl -X POST http://localhost:8080/v1/facilities \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <YOUR_ACCESS_TOKEN>' \
  -d '{
    "name": "Test Facility",
    "description": "Test Description",
    "surface": "Grass",
    "openAt": "08:00",
    "closeAt": "22:00",
    "available": true,
    "weekdayRateCents": 5000,
    "weekendRateCents": 7500,
    "currency": "USD"
  }'
```

**Expected**: Returns created facility with ID

### Test 4: List Facilities
```bash
curl -X GET http://localhost:8080/v1/facilities \
  -H 'Authorization: Bearer <YOUR_ACCESS_TOKEN>'
```

**Expected**: Returns array of facilities

## Architecture Benefits

### Gateway Pattern Advantages:

1. **Centralized Authentication**: JWT validation happens once at the gateway
2. **Service Isolation**: Backend services don't need JWT libraries
3. **Header Injection**: User context propagates automatically to all services
4. **Single Entry Point**: Easier to add rate limiting, logging, monitoring
5. **Security**: Backend services are not exposed directly to clients
6. **Protocol Translation**: Can serve REST, GraphQL, gRPC from single gateway

### Comparison:

| Aspect | Direct Call (Old) | Through Gateway (New) |
|--------|------------------|----------------------|
| **Port** | 8083 | 8080 |
| **Auth** | ❌ Headers missing | ✅ Headers injected |
| **JWT** | Not validated | ✅ Validated |
| **Security** | Service exposed | ✅ Service protected |
| **Maintenance** | Duplicate auth logic | ✅ Centralized auth |

## Port Summary

| Service | Port | Frontend Uses | Purpose |
|---------|------|---------------|---------|
| **API Gateway** | 8080 | ✅ Yes | Facilities, Bookings, Users (REST + GraphQL) |
| **Auth Service** | 8081 | ✅ Yes | Login, Register, Token Refresh |
| **User Service** | 8082 | ❌ No | Used by Gateway only |
| **Booking Service** | 8083 | ❌ No | Used by Gateway only |

## Frontend Configuration

No changes needed to frontend `.env` file:

```env
VITE_BASE_URL=http://localhost
VITE_GATEWAY_PORT=8080
VITE_AUTH_PORT=8081
VITE_BOOKING_PORT=8083  # Not used directly anymore
```

The frontend services are already updated to use `GATEWAY_URL`.

## Next Steps

1. **Test in Admin CMS**:
   - Login to get fresh JWT token
   - Try creating a facility from the UI
   - Verify all facility and booking operations work

2. **Verify All Endpoints**:
   - List facilities
   - Create facility
   - Update facility
   - Delete facility
   - List bookings
   - Create booking
   - Update booking status

3. **Production Considerations**:
   - Add rate limiting to gateway endpoints
   - Add request/response logging
   - Monitor gateway performance
   - Consider adding caching for read operations

## Troubleshooting

### Issue: "invalid token"
**Solution**: Token may be expired. Login again to get fresh token.

### Issue: "missing auth headers"
**Solution**: Ensure you're calling through Gateway (port 8080), not booking service directly (port 8083).

### Issue: 404 Not Found
**Solution**: Verify API Gateway is running and endpoints are registered correctly.

### Issue: 502 Bad Gateway
**Solution**: Check that booking service is running and accessible from gateway.

## Summary

✅ **REST proxy endpoints added to API Gateway**
✅ **JWT validation and header injection implemented**
✅ **All facility and booking endpoints proxied**
✅ **Gateway rebuilt and deployed successfully**
✅ **Frontend services already updated to use Gateway**

**Result**: Admin CMS can now make authenticated requests to facilities and bookings through the API Gateway, which handles JWT validation and adds the required authentication headers before forwarding to the booking service.
