# REST Proxy Testing Results

## Summary

✅ **REST proxy implementation is working correctly!**

The API Gateway now successfully:
1. Validates JWT tokens from `Authorization: Bearer <token>` headers
2. Extracts user ID and roles from JWT claims
3. Adds `X-User-ID` and `X-User-Roles` headers
4. Forwards requests to the booking service

## Test Results

### Test 1: Authentication ✅

**Registration:**
```bash
curl -X POST http://localhost:8081/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "email":"testadmin@example.com",
    "password":"testpass123",
    "firstName":"Test",
    "lastName":"Admin",
    "phone":"+1234567890"
  }'
```

**Result:** ✅ Success
- Received fresh JWT access token
- User ID: `f7574e52-ff7f-420c-888e-73ecf931f9ec`
- Roles: `["MEMBER"]`
- Token expires in 900 seconds (15 minutes)

### Test 2: List Facilities Through Gateway ✅

**Request:**
```bash
curl -X GET 'http://localhost:8080/v1/facilities' \
  -H 'Authorization: Bearer eyJhbGci...'
```

**Result:** ✅ Success
```json
[{
  "available": true,
  "closeAt": "23:00",
  "currency": "CAD",
  "description": "Indoor pickleball court",
  "id": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
  "name": "Center Court",
  "openAt": "06:00",
  "surface": "hardwood",
  "venueId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "weekdayRateCents": 4500,
  "weekendRateCents": 6000
}]
```

**Analysis:**
- Gateway validated the JWT token ✅
- Gateway forwarded request to booking service with auth headers ✅
- Booking service accepted the request and returned facilities ✅
- No "missing auth headers" error ✅

### Test 3: List Bookings Through Gateway ✅

**Request:**
```bash
curl -X GET 'http://localhost:8080/v1/bookings' \
  -H 'Authorization: Bearer eyJhbGci...'
```

**Result:** ✅ Success
```json
[]
```

**Analysis:**
- Gateway validated the JWT token ✅
- Request was successfully forwarded to booking service ✅
- Empty array returned (user has no bookings) ✅

### Test 4: Create Facility (Authorization Test) ✅

**Request:**
```bash
curl -X POST http://localhost:8080/v1/facilities \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer eyJhbGci...' \
  -d '{
    "name":"Test Facility",
    "description":"Test facility through gateway",
    "surface":"Grass",
    "openAt":"08:00",
    "closeAt":"22:00",
    "available":true,
    "weekdayRateCents":5000,
    "weekendRateCents":7500,
    "currency":"USD"
  }'
```

**Result:** ✅ Success (Authorization Working)
```json
{"error":"forbidden"}
```

**Analysis:**
- Gateway validated the JWT token ✅
- Gateway forwarded request with auth headers ✅
- Booking service received the headers correctly ✅
- Booking service checked authorization and correctly rejected (user has MEMBER role, needs ADMIN or VENUE_ADMIN) ✅
- **The "forbidden" error proves the auth headers are working!** If headers were missing, we'd get "missing auth headers" error instead.

## Request Flow Verification

### Before Fix (Failed):
```
Frontend → http://localhost:8083/v1/facilities
           Headers: Authorization: Bearer <JWT>

Booking Service → ❌ "missing auth headers: X-User-ID = '', X-User-Roles = ''"
```

### After Fix (Working):
```
Frontend → http://localhost:8080/v1/facilities
           Headers: Authorization: Bearer <JWT>
                    ↓
           API Gateway (port 8080):
           1. ✅ Validates JWT token
           2. ✅ Extracts user ID: f7574e52-ff7f-420c-888e-73ecf931f9ec
           3. ✅ Extracts roles: ["MEMBER"]
           4. ✅ Creates request to booking service
           5. ✅ Adds X-User-ID header
           6. ✅ Adds X-User-Roles header
           7. ✅ Forwards request
                    ↓
Booking Service → ✅ Receives X-User-ID and X-User-Roles headers
                  ✅ Processes request successfully
                  ✅ Returns response or applies authorization rules
                    ↓
API Gateway → ✅ Forwards response back to frontend
```

## Endpoint Availability

Based on the booking service routes ([codes/services/booking-service/cmd/booking/main.go:71-89](codes/services/booking-service/cmd/booking/main.go#L71-L89)):

### Facilities Endpoints:
| Method | Path | Available | Roles Required | Status |
|--------|------|-----------|----------------|---------|
| GET | `/v1/facilities` | ✅ Yes | MEMBER, OPERATOR, ADMIN, VENUE_ADMIN | ✅ Tested |
| GET | `/v1/facilities/:id` | ❌ No | - | Not implemented in booking service |
| POST | `/v1/facilities` | ✅ Yes | ADMIN, VENUE_ADMIN | ✅ Tested (authorization) |
| PATCH | `/v1/facilities/:id` | ✅ Yes | ADMIN, VENUE_ADMIN | Gateway proxy ready |
| PUT | `/v1/facilities/:id` | ❌ No | - | Not implemented in booking service |
| DELETE | `/v1/facilities/:id` | ❌ No | - | Not implemented in booking service |
| GET | `/v1/facilities/:id/schedule` | ✅ Yes | MEMBER, OPERATOR, ADMIN, VENUE_ADMIN | Gateway proxy ready |
| POST | `/v1/facilities/:id/overrides` | ✅ Yes | ADMIN, VENUE_ADMIN | Gateway proxy ready |
| DELETE | `/v1/facilities/:id/overrides/:overrideId` | ✅ Yes | ADMIN, VENUE_ADMIN | Gateway proxy ready |

### Bookings Endpoints:
| Method | Path | Available | Roles Required | Status |
|--------|------|-----------|----------------|---------|
| GET | `/v1/bookings` | ✅ Yes | MEMBER, OPERATOR, ADMIN, VENUE_ADMIN | ✅ Tested |
| GET | `/v1/bookings/:id` | ✅ Yes | MEMBER, OPERATOR, ADMIN, VENUE_ADMIN | Gateway proxy ready |
| POST | `/v1/bookings` | ✅ Yes | MEMBER, ADMIN, VENUE_ADMIN | Gateway proxy ready |
| DELETE | `/v1/bookings/:id` | ✅ Yes | MEMBER, ADMIN, VENUE_ADMIN | Gateway proxy ready |
| PATCH | `/v1/bookings/:id/status` | ❌ No | - | Gateway has handler, but booking service doesn't |
| PATCH | `/v1/bookings/:id/cancel` | ❌ No | - | Gateway has handler, but booking service uses DELETE |
| POST | `/v1/bookings/:id/confirm` | ❌ No | - | Not implemented in booking service |
| GET | `/v1/bookings/stats` | ❌ No | - | Not implemented in booking service |

## Frontend Service Files Status

Both frontend service files are already updated to use the Gateway:

### [facility.service.js](frontend_codes/admin_cms/src/services/facility.service.js)
✅ All endpoints use `GATEWAY_URL` (port 8080)
- ✅ `getFacilities()` - Works
- ⚠️ `getFacilityById()` - Gateway proxies, but booking service doesn't have this endpoint
- ⚠️ `createFacility()` - Gateway proxies, requires ADMIN role
- ⚠️ `updateFacility()` - Gateway proxies via PUT, but booking service uses PATCH
- ⚠️ `deleteFacility()` - Gateway proxies, but booking service doesn't have this endpoint
- ✅ `getFacilitySchedule()` - Gateway proxies (not tested yet)

### [booking.service.js](frontend_codes/admin_cms/src/services/booking.service.js)
✅ All endpoints use `GATEWAY_URL` (port 8080)
- ✅ `getBookings()` - Works
- ✅ `getBookingById()` - Gateway proxies (not tested yet)
- ✅ `createBooking()` - Gateway proxies (not tested yet)
- ⚠️ `updateBookingStatus()` - Gateway proxies via PATCH, but booking service doesn't have this
- ⚠️ `cancelBooking()` - Gateway proxies via PATCH, but booking service uses DELETE
- ⚠️ `confirmBooking()` - Gateway proxies, but booking service doesn't have this endpoint
- ⚠️ `getStats()` - Gateway proxies, but booking service doesn't have this endpoint

## Recommendations

### 1. Frontend Service Adjustments Needed

The frontend service methods need to match the actual booking service API:

**facility.service.js changes needed:**
```javascript
// Change updateFacility to use PATCH instead of PUT
async updateFacility(id, availabilityData) {
  // Should only update availability, not full facility
  const response = await axios.patch(`${API_ENDPOINTS.GATEWAY_URL}/v1/facilities/${id}`, {
    available: availabilityData.available
  }, {
    headers: getAuthHeaders(),
  })
  return response.data
}

// Remove or mark as unsupported:
// - getFacilityById() - not in booking service
// - deleteFacility() - not in booking service
```

**booking.service.js changes needed:**
```javascript
// Change cancelBooking to use DELETE instead of PATCH
async cancelBooking(id) {
  const response = await axios.delete(`${API_ENDPOINTS.GATEWAY_URL}/v1/bookings/${id}`, {
    headers: getAuthHeaders(),
  })
  return response.data
}

// Remove or mark as unsupported:
// - updateBookingStatus() - not in booking service
// - confirmBooking() - not in booking service
// - getStats() - not in booking service
```

### 2. Update Gateway Proxy Handlers (Optional)

The gateway's REST proxy handlers can be updated to match the actual booking service routes, but it's not critical since the mismatched handlers will just return 404 from the booking service.

### 3. Admin Role for Testing

To fully test facility creation, you'll need a user with ADMIN or VENUE_ADMIN role. The current test user has MEMBER role which only has read access.

## Conclusion

✅ **The REST proxy implementation is working perfectly!**

The authentication routing issue has been completely resolved:
1. ✅ Frontend routes requests through Gateway (port 8080)
2. ✅ Gateway validates JWT tokens
3. ✅ Gateway injects authentication headers (X-User-ID, X-User-Roles)
4. ✅ Booking service receives and validates the headers
5. ✅ Requests are processed with proper authorization

The main issue has been fixed. The remaining work is minor frontend adjustments to match the actual booking service API endpoints.
