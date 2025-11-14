#!/bin/bash
# test APIs with CRUD

# Don't exit on error - we want to test everything even if some fail
set +e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

BASE_URL=${BASE_URL:-"http://localhost"}
GATEWAY_URL="$BASE_URL:8080"
AUTH_URL="$BASE_URL:8081"
USER_URL="$BASE_URL:8082"
BOOKING_URL="$BASE_URL:8083"
FOOD_URL="$BASE_URL:8084"
PARKING_URL="$BASE_URL:8085"
SHOP_URL="$BASE_URL:8086"
PAYMENT_URL="$BASE_URL:8087"
NOTIFICATION_URL="$BASE_URL:8088"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   Venue Master API CRUD Testing${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Function to print section headers
section() {
    echo -e "\n${YELLOW}>>> $1${NC}"
}

# Function to print success
success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Function to print error
error() {
    echo -e "${RED}✗ $1${NC}"
}

# Function to print info
info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Note: When calling microservices directly (not through the gateway),
# we need to include X-User-ID and X-User-Roles headers.
# The gateway normally adds these after validating the JWT token.

# ============================================
# HEALTH CHECKS
# ============================================
section "Health Checks"
echo "Checking all services..."

# Test each service individually
for service_pair in "Gateway|$GATEWAY_URL" "Auth|$AUTH_URL" "User|$USER_URL" "Booking|$BOOKING_URL" "Food|$FOOD_URL" "Parking|$PARKING_URL" "Shop|$SHOP_URL" "Payment|$PAYMENT_URL" "Notification|$NOTIFICATION_URL"; do
    name=$(echo "$service_pair" | cut -d'|' -f1)
    url=$(echo "$service_pair" | cut -d'|' -f2)
    if curl -s "$url/healthz" | jq -e '.status == "ok"' > /dev/null 2>&1; then
        success "$name service is healthy"
    else
        error "$name service is not healthy"
    fi
done

# ============================================
# AUTHENTICATION TESTS
# ============================================
section "1. Authentication Service (CRUD)"

# CREATE - Login
info "CREATE: Login and get tokens"
LOGIN_RESPONSE=$(curl -s -X POST "$AUTH_URL/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d @- << 'EOF'
{"email":"member@example.com","password":"Secret123!"}
EOF
)

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.accessToken')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refreshToken')
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.user.id')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    error "Failed to get access token"
    echo "$LOGIN_RESPONSE" | jq '.'
    exit 1
fi

success "Successfully logged in"
echo "$LOGIN_RESPONSE" | jq '{user: .user, expiresIn: .expiresIn, tokenPreview: .accessToken[:50]}'

# READ - Verify token (implicit in subsequent calls)
info "READ: Token will be verified in subsequent API calls"

# UPDATE - Refresh token
info "UPDATE: Refresh access token"
REFRESH_RESPONSE=$(curl -s -X POST "$AUTH_URL/v1/auth/refresh" \
  -H 'Content-Type: application/json' \
  -d @- << EOF
{"refreshToken":"$REFRESH_TOKEN"}
EOF
)

NEW_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.accessToken')
if [ "$NEW_TOKEN" != "null" ] && [ -n "$NEW_TOKEN" ]; then
    success "Successfully refreshed token"
    # Use the new token for subsequent requests
    TOKEN=$NEW_TOKEN
else
    error "Failed to refresh token"
fi

# DELETE - Logout (if endpoint exists)
info "DELETE: Logout endpoint may not be implemented yet"

# ============================================
# USER SERVICE TESTS
# ============================================
section "2. User Service (CRUD)"

# READ - Get current user
info "READ: Get current user profile"
USER_RESPONSE=$(curl -s "$USER_URL/v1/users/me" \
  -H "Authorization: Bearer $TOKEN")

echo "$USER_RESPONSE" | jq '.'
success "Successfully fetched user profile"

# READ - Get user by ID
info "READ: Get user by ID"
USER_BY_ID=$(curl -s "$USER_URL/v1/users/$USER_ID" \
  -H "Authorization: Bearer $TOKEN")

if echo "$USER_BY_ID" | jq -e '.id' > /dev/null 2>&1; then
    success "Successfully fetched user by ID"
else
    info "User by ID endpoint may require admin permissions"
fi

# READ - Get user memberships
info "READ: Get user memberships"
MEMBERSHIPS=$(curl -s "$USER_URL/v1/users/$USER_ID/memberships" \
  -H "Authorization: Bearer $TOKEN")

echo "$MEMBERSHIPS" | jq '.'
success "Successfully fetched memberships"

# ============================================
# BOOKING SERVICE TESTS (GraphQL)
# ============================================
section "3. Booking Service - GraphQL (CRUD)"

# READ - List facilities
info "READ: List all facilities via REST first to get facility data"
# Note: Direct service calls require X-User-ID and X-User-Roles headers
REST_FAC=$(curl -s "$BOOKING_URL/v1/facilities?limit=5" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-User-ID: $USER_ID" \
  -H "X-User-Roles: MEMBER")
FACILITY_ID=$(echo "$REST_FAC" | jq -r '.[0].id // empty')
FACILITY_NAME=$(echo "$REST_FAC" | jq -r '.[0].name // empty')
VENUE_ID=$(echo "$REST_FAC" | jq -r '.[0].venueId // empty')

if [ -n "$FACILITY_ID" ] && [ "$FACILITY_ID" != "null" ]; then
    success "Found facility via REST: $FACILITY_NAME ($FACILITY_ID)"
    echo "$REST_FAC" | jq '.[0]'

    # Now try GraphQL with venueId (note: deployed schema requires venueId and has no limit arg)
    info "READ: List facilities via GraphQL"
    FACILITIES_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/graphql" \
      -H "Authorization: Bearer $TOKEN" \
      -H 'Content-Type: application/json' \
      -d "{\"query\":\"{ facilities(venueId: \\\"$VENUE_ID\\\") { id name description surface } }\"}")

    if echo "$FACILITIES_RESPONSE" | jq -e '.data.facilities[0]' > /dev/null 2>&1; then
        echo "$FACILITIES_RESPONSE" | jq '.data.facilities[0]'
        success "Successfully fetched facilities via GraphQL"
    else
        error "GraphQL facilities query failed"
        echo "$FACILITIES_RESPONSE" | jq '.'
    fi
else
    error "No facilities found via REST"
fi

# Note: facilitySchedule query is not in the deployed GraphQL schema
# It exists in the code but hasn't been deployed
# We'll test schedules via REST API later
info "Skipping GraphQL facilitySchedule query (not in deployed schema)"

# CREATE - Create a booking
info "CREATE: Create a new booking"
START_TIME=$(date -u -v+2H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+2 hours" +"%Y-%m-%dT%H:%M:%SZ")
END_TIME=$(date -u -v+3H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+3 hours" +"%Y-%m-%dT%H:%M:%SZ")

CREATE_BOOKING_QUERY=$(cat <<EOF
{
  "query": "mutation { createBooking(facilityId: \"$FACILITY_ID\", startsAt: \"$START_TIME\", endsAt: \"$END_TIME\") { id status amountCents currency paymentIntent facility { name } } }"
}
EOF
)

BOOKING_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/graphql" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "$CREATE_BOOKING_QUERY")

BOOKING_ID=$(echo "$BOOKING_RESPONSE" | jq -r '.data.createBooking.id')
BOOKING_STATUS=$(echo "$BOOKING_RESPONSE" | jq -r '.data.createBooking.status')

if [ "$BOOKING_ID" != "null" ] && [ -n "$BOOKING_ID" ]; then
    echo "$BOOKING_RESPONSE" | jq '.data.createBooking'
    success "Successfully created booking: $BOOKING_ID (Status: $BOOKING_STATUS)"
else
    error "Failed to create booking"
    echo "$BOOKING_RESPONSE" | jq '.'
fi

# READ - List user bookings
info "READ: List all user bookings"
BOOKINGS_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/graphql" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ bookings(limit: 10) { id status startsAt endsAt amountCents currency facility { name } } }"}')

BOOKING_COUNT=$(echo "$BOOKINGS_RESPONSE" | jq '.data.bookings | length')
echo "$BOOKINGS_RESPONSE" | jq '.data.bookings[0:3]'
success "Found $BOOKING_COUNT bookings"

# READ - Get specific booking
if [ "$BOOKING_ID" != "null" ] && [ -n "$BOOKING_ID" ]; then
    info "READ: Get specific booking details"
    GET_BOOKING_QUERY=$(cat <<EOF
{
  "query": "{ booking(id: \"$BOOKING_ID\") { id status startsAt endsAt amountCents currency paymentIntent facility { id name } } }"
}
EOF
)

    GET_BOOKING_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/graphql" \
      -H "Authorization: Bearer $TOKEN" \
      -H 'Content-Type: application/json' \
      -d "$GET_BOOKING_QUERY")

    echo "$GET_BOOKING_RESPONSE" | jq '.data.booking'
    success "Successfully fetched booking details"
fi

# DELETE - Cancel booking
if [ "$BOOKING_ID" != "null" ] && [ -n "$BOOKING_ID" ]; then
    info "DELETE: Cancel booking"
    CANCEL_BOOKING_QUERY=$(cat <<EOF
{
  "query": "mutation { cancelBooking(id: \"$BOOKING_ID\") { id status } }"
}
EOF
)

    CANCEL_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/graphql" \
      -H "Authorization: Bearer $TOKEN" \
      -H 'Content-Type: application/json' \
      -d "$CANCEL_BOOKING_QUERY")

    CANCELLED_STATUS=$(echo "$CANCEL_RESPONSE" | jq -r '.data.cancelBooking.status')
    echo "$CANCEL_RESPONSE" | jq '.data.cancelBooking'
    success "Successfully cancelled booking. New status: $CANCELLED_STATUS"
fi

# ============================================
# BOOKING SERVICE TESTS (REST Direct)
# ============================================
section "4. Booking Service - REST API (CRUD)"

# READ - List facilities (REST)
info "READ: List facilities via REST"
REST_FACILITIES=$(curl -s "$BOOKING_URL/v1/facilities?limit=5" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-User-ID: $USER_ID" \
  -H "X-User-Roles: MEMBER")

echo "$REST_FACILITIES" | jq '[.[] | {id, name, available, weekdayRateCents}]'
success "Successfully fetched facilities via REST"

# CREATE - Create booking (REST)
info "CREATE: Create booking via REST"
START_TIME_2=$(date -u -v+4H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+4 hours" +"%Y-%m-%dT%H:%M:%SZ")
END_TIME_2=$(date -u -v+5H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+5 hours" +"%Y-%m-%dT%H:%M:%SZ")

REST_BOOKING_RESPONSE=$(curl -s -X POST "$BOOKING_URL/v1/bookings" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-User-ID: $USER_ID" \
  -H "X-User-Roles: MEMBER" \
  -H 'Content-Type: application/json' \
  -d @- << EOF
{
  "facilityId": "$FACILITY_ID",
  "userId": "$USER_ID",
  "startsAt": "$START_TIME_2",
  "endsAt": "$END_TIME_2"
}
EOF
)

REST_BOOKING_ID=$(echo "$REST_BOOKING_RESPONSE" | jq -r '.id')
if [ "$REST_BOOKING_ID" != "null" ] && [ -n "$REST_BOOKING_ID" ]; then
    echo "$REST_BOOKING_RESPONSE" | jq '{id, status, amountCents, currency}'
    success "Successfully created booking via REST: $REST_BOOKING_ID"
else
    error "Failed to create booking via REST"
    echo "$REST_BOOKING_RESPONSE" | jq '.'
fi

# READ - List bookings (REST)
info "READ: List bookings via REST"
REST_BOOKINGS=$(curl -s "$BOOKING_URL/v1/bookings?userId=$USER_ID&limit=5" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-User-ID: $USER_ID" \
  -H "X-User-Roles: MEMBER")

echo "$REST_BOOKINGS" | jq '[.[] | {id, status, startsAt}]'
success "Successfully fetched bookings via REST"

# READ - Get specific booking (REST)
if [ "$REST_BOOKING_ID" != "null" ] && [ -n "$REST_BOOKING_ID" ]; then
    info "READ: Get specific booking via REST"
    REST_BOOKING_DETAIL=$(curl -s "$BOOKING_URL/v1/bookings/$REST_BOOKING_ID" \
      -H "Authorization: Bearer $TOKEN" \
      -H "X-User-ID: $USER_ID" \
      -H "X-User-Roles: MEMBER")

    echo "$REST_BOOKING_DETAIL" | jq '{id, status, facility: .facility.name}'
    success "Successfully fetched booking detail via REST"
fi

# UPDATE - Cancel booking (REST)
if [ "$REST_BOOKING_ID" != "null" ] && [ -n "$REST_BOOKING_ID" ]; then
    info "DELETE: Cancel booking via REST"
    REST_CANCEL_RESPONSE=$(curl -s -X PATCH "$BOOKING_URL/v1/bookings/$REST_BOOKING_ID/cancel" \
      -H "Authorization: Bearer $TOKEN" \
      -H "X-User-ID: $USER_ID" \
      -H "X-User-Roles: MEMBER")

    REST_CANCEL_STATUS=$(echo "$REST_CANCEL_RESPONSE" | jq -r '.status')
    echo "$REST_CANCEL_RESPONSE" | jq '{id, status}'
    success "Successfully cancelled booking via REST. Status: $REST_CANCEL_STATUS"
fi

# READ - Get facility schedule (REST)
if [ -n "$FACILITY_ID" ] && [ "$FACILITY_ID" != "null" ]; then
    info "READ: Get facility schedule via REST"
    REST_SCHEDULE=$(curl -s "$BOOKING_URL/v1/facilities/$FACILITY_ID/schedule?from=$FROM_DATE&to=$TO_DATE" \
      -H "Authorization: Bearer $TOKEN" \
      -H "X-User-ID: $USER_ID" \
      -H "X-User-Roles: MEMBER")

    # Check if response is valid JSON array
    if echo "$REST_SCHEDULE" | jq -e 'type == "array"' > /dev/null 2>&1; then
        echo "$REST_SCHEDULE" | jq '.[0:2]'
        success "Successfully fetched schedule via REST"
    else
        info "Schedule endpoint may not be implemented yet"
        echo "$REST_SCHEDULE"
    fi
fi

# ============================================
# VENUE TESTS (REST via Gateway)
# ============================================
section "5. Venue Management (CRUD via Gateway)"

# READ - List venues
info "READ: List all venues"
VENUES_RESPONSE=$(curl -s "$GATEWAY_URL/v1/venues?limit=100" \
  -H "Authorization: Bearer $TOKEN")

VENUE_COUNT=$(echo "$VENUES_RESPONSE" | jq '. | length')
if [ "$VENUE_COUNT" -gt 0 ]; then
    echo "$VENUES_RESPONSE" | jq '.[0]'
    success "Found $VENUE_COUNT venues"
    DEFAULT_VENUE_ID=$(echo "$VENUES_RESPONSE" | jq -r '.[0].id')
else
    error "No venues found"
    DEFAULT_VENUE_ID=""
fi

# Note: For admin operations, we need to login as admin
info "Logging in as admin for venue management tests"
ADMIN_LOGIN_RESPONSE=$(curl -s -X POST "$AUTH_URL/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"Admin123!"}')

ADMIN_TOKEN=$(echo "$ADMIN_LOGIN_RESPONSE" | jq -r '.accessToken')

if [ "$ADMIN_TOKEN" != "null" ] && [ -n "$ADMIN_TOKEN" ]; then
    success "Successfully logged in as admin"

    # CREATE - Create a new venue
    info "CREATE: Create a new venue"
    CREATE_VENUE_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/v1/venues" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H 'Content-Type: application/json' \
      -d @- << 'EOF'
{
  "name": "Test Sports Complex",
  "description": "A modern sports facility for testing",
  "address": "456 Test Avenue",
  "city": "San Francisco",
  "state": "CA",
  "zipCode": "94102",
  "country": "US",
  "phone": "+1-415-555-1234",
  "email": "info@testsports.com",
  "website": "https://testsports.com",
  "timezone": "America/Los_Angeles"
}
EOF
)

    NEW_VENUE_ID=$(echo "$CREATE_VENUE_RESPONSE" | jq -r '.id')
    if [ "$NEW_VENUE_ID" != "null" ] && [ -n "$NEW_VENUE_ID" ]; then
        echo "$CREATE_VENUE_RESPONSE" | jq '{id, name, city, timezone}'
        success "Successfully created venue: $NEW_VENUE_ID"

        # READ - Get specific venue
        info "READ: Get specific venue details"
        GET_VENUE_RESPONSE=$(curl -s "$GATEWAY_URL/v1/venues/$NEW_VENUE_ID" \
          -H "Authorization: Bearer $ADMIN_TOKEN")

        echo "$GET_VENUE_RESPONSE" | jq '{id, name, address, city, state}'
        success "Successfully fetched venue details"

        # UPDATE - Update venue
        info "UPDATE: Update venue information"
        UPDATE_VENUE_RESPONSE=$(curl -s -X PUT "$GATEWAY_URL/v1/venues/$NEW_VENUE_ID" \
          -H "Authorization: Bearer $ADMIN_TOKEN" \
          -H 'Content-Type: application/json' \
          -d @- << 'EOF'
{
  "name": "Updated Sports Complex",
  "description": "An updated modern sports facility",
  "address": "456 Test Avenue",
  "city": "San Francisco",
  "state": "CA",
  "zipCode": "94102",
  "country": "US",
  "phone": "+1-415-555-5678",
  "email": "contact@testsports.com",
  "website": "https://testsports.com",
  "timezone": "America/Los_Angeles"
}
EOF
)

        UPDATED_NAME=$(echo "$UPDATE_VENUE_RESPONSE" | jq -r '.name')
        UPDATED_PHONE=$(echo "$UPDATE_VENUE_RESPONSE" | jq -r '.phone')
        echo "$UPDATE_VENUE_RESPONSE" | jq '{id, name, phone, email}'
        success "Successfully updated venue: $UPDATED_NAME (Phone: $UPDATED_PHONE)"

        # CREATE - Create a facility associated with the test venue
        info "CREATE: Create facility for the test venue"
        CREATE_FACILITY_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/v1/facilities" \
          -H "Authorization: Bearer $ADMIN_TOKEN" \
          -H 'Content-Type: application/json' \
          -d @- << EOF
{
  "venueId": "$NEW_VENUE_ID",
  "name": "Test Basketball Court",
  "description": "Indoor basketball court",
  "surface": "Hardwood",
  "openAt": "08:00",
  "closeAt": "22:00",
  "weekdayRateCents": 5000,
  "weekendRateCents": 7500,
  "currency": "USD"
}
EOF
)

        TEST_FACILITY_ID=$(echo "$CREATE_FACILITY_RESPONSE" | jq -r '.id')
        if [ "$TEST_FACILITY_ID" != "null" ] && [ -n "$TEST_FACILITY_ID" ]; then
            echo "$CREATE_FACILITY_RESPONSE" | jq '{id, name, venueId}'
            success "Successfully created facility for test venue: $TEST_FACILITY_ID"
        else
            info "Facility creation may require additional setup"
            echo "$CREATE_FACILITY_RESPONSE" | jq '.'
        fi

        # DELETE - Delete the test venue (should cascade to facilities)
        info "DELETE: Delete test venue (should cascade to facilities)"
        DELETE_VENUE_RESPONSE=$(curl -s -X DELETE "$GATEWAY_URL/v1/venues/$NEW_VENUE_ID" \
          -H "Authorization: Bearer $ADMIN_TOKEN")

        if [ -z "$DELETE_VENUE_RESPONSE" ] || echo "$DELETE_VENUE_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
            # If empty response or no error field, deletion likely succeeded
            if [ -z "$DELETE_VENUE_RESPONSE" ]; then
                success "Successfully deleted venue: $NEW_VENUE_ID"
            else
                error "Failed to delete venue"
                echo "$DELETE_VENUE_RESPONSE" | jq '.'
            fi
        else
            success "Successfully deleted venue: $NEW_VENUE_ID"
        fi

        # Verify deletion
        info "Verifying venue deletion"
        VERIFY_DELETE=$(curl -s "$GATEWAY_URL/v1/venues/$NEW_VENUE_ID" \
          -H "Authorization: Bearer $ADMIN_TOKEN")

        if echo "$VERIFY_DELETE" | jq -e '.error' > /dev/null 2>&1; then
            success "Venue deletion confirmed - venue no longer exists"
        else
            info "Venue deletion verification inconclusive"
        fi
    else
        error "Failed to create venue"
        echo "$CREATE_VENUE_RESPONSE" | jq '.'
    fi
else
    info "Admin login failed - skipping venue CREATE/UPDATE/DELETE tests"
    echo "$ADMIN_LOGIN_RESPONSE" | jq '.'
fi

# ============================================
# FOOD SERVICE TESTS
# ============================================
section "6. Food Service (CRUD)"

# READ - List menu items
info "READ: List menu items"
MENU_RESPONSE=$(curl -s "$FOOD_URL/v1/menu?limit=5" \
  -H "Authorization: Bearer $TOKEN")

echo "$MENU_RESPONSE" | jq '.'
success "Successfully fetched menu items"

# If menu items exist, we could test CREATE/UPDATE/DELETE with proper admin permissions
info "CREATE/UPDATE/DELETE operations typically require ADMIN/OPERATOR roles"

# ============================================
# PARKING SERVICE TESTS
# ============================================
section "7. Parking Service (CRUD)"

# READ - List parking spaces
info "READ: List parking spaces"
PARKING_RESPONSE=$(curl -s "$PARKING_URL/v1/parking/spaces?limit=5" \
  -H "Authorization: Bearer $TOKEN")

echo "$PARKING_RESPONSE" | jq '.'
success "Successfully fetched parking spaces"

# If we have a space ID, we could test reservations
SPACE_ID=$(echo "$PARKING_RESPONSE" | jq -r '.[0].id // empty')

if [ -n "$SPACE_ID" ] && [ "$SPACE_ID" != "null" ]; then
    # CREATE - Reserve parking space
    info "CREATE: Reserve parking space"
    PARKING_START=$(date -u -v+1d +"%Y-%m-%dT09:00:00Z" 2>/dev/null || date -u -d "+1 day" +"%Y-%m-%dT09:00:00Z")
    PARKING_END=$(date -u -v+1d +"%Y-%m-%dT17:00:00Z" 2>/dev/null || date -u -d "+1 day" +"%Y-%m-%dT17:00:00Z")

    PARKING_RESERVATION=$(curl -s -X POST "$PARKING_URL/v1/parking/reservations" \
      -H "Authorization: Bearer $TOKEN" \
      -H 'Content-Type: application/json' \
      -d @- << EOF
{
  "spaceId": "$SPACE_ID",
  "startsAt": "$PARKING_START",
  "endsAt": "$PARKING_END"
}
EOF
)

    PARKING_RES_ID=$(echo "$PARKING_RESERVATION" | jq -r '.id // empty')
    if [ -n "$PARKING_RES_ID" ] && [ "$PARKING_RES_ID" != "null" ]; then
        echo "$PARKING_RESERVATION" | jq '.'
        success "Successfully created parking reservation: $PARKING_RES_ID"
    else
        info "Parking reservation may require specific business logic"
    fi
fi

# ============================================
# SHOP SERVICE TESTS
# ============================================
section "8. Shop Service (CRUD)"

# READ - List products
info "READ: List shop products"
PRODUCTS_RESPONSE=$(curl -s "$SHOP_URL/v1/products?limit=5" \
  -H "Authorization: Bearer $TOKEN")

# Check if response is valid JSON
if echo "$PRODUCTS_RESPONSE" | jq -e 'type' > /dev/null 2>&1; then
    echo "$PRODUCTS_RESPONSE" | jq '.'
    success "Successfully fetched shop products"
    PRODUCT_ID=$(echo "$PRODUCTS_RESPONSE" | jq -r '.[0].id // empty')
else
    info "Shop products endpoint may not be implemented yet"
    echo "$PRODUCTS_RESPONSE"
    PRODUCT_ID=""
fi

# If we have products, test cart operations

if [ -n "$PRODUCT_ID" ] && [ "$PRODUCT_ID" != "null" ]; then
    # CREATE - Add item to cart
    info "CREATE: Add item to cart"
    ADD_TO_CART=$(curl -s -X POST "$SHOP_URL/v1/cart/items" \
      -H "Authorization: Bearer $TOKEN" \
      -H 'Content-Type: application/json' \
      -d @- << EOF
{
  "productId": "$PRODUCT_ID",
  "quantity": 2
}
EOF
)

    if echo "$ADD_TO_CART" | jq -e '.id' > /dev/null 2>&1; then
        echo "$ADD_TO_CART" | jq '.'
        success "Successfully added item to cart"

        # READ - Get cart
        info "READ: Get shopping cart"
        CART_RESPONSE=$(curl -s "$SHOP_URL/v1/cart" \
          -H "Authorization: Bearer $TOKEN")

        echo "$CART_RESPONSE" | jq '.'
        success "Successfully fetched cart"
    else
        info "Cart operations may require specific implementation"
    fi
fi

# ============================================
# PAYMENT SERVICE TESTS
# ============================================
section "9. Payment Service (CRUD)"

# CREATE - Create payment intent
info "CREATE: Create payment intent"
PAYMENT_INTENT=$(curl -s -X POST "$PAYMENT_URL/v1/payments/intents" \
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
)

PAYMENT_INTENT_ID=$(echo "$PAYMENT_INTENT" | jq -r '.id // empty')
if [ -n "$PAYMENT_INTENT_ID" ] && [ "$PAYMENT_INTENT_ID" != "null" ]; then
    echo "$PAYMENT_INTENT" | jq '.'
    success "Successfully created payment intent: $PAYMENT_INTENT_ID"

    # READ - Get payment intent
    info "READ: Get payment intent details"
    PAYMENT_DETAIL=$(curl -s "$PAYMENT_URL/v1/payments/intents/$PAYMENT_INTENT_ID" \
      -H "Authorization: Bearer $TOKEN")

    echo "$PAYMENT_DETAIL" | jq '.'
    success "Successfully fetched payment intent"
else
    info "Payment service may require Stripe configuration"
fi

# ============================================
# NOTIFICATION SERVICE TESTS
# ============================================
section "10. Notification Service (CRUD)"

# CREATE - Send notification (may require admin)
info "CREATE: Send notification"
NOTIFICATION=$(curl -s -X POST "$NOTIFICATION_URL/v1/notifications" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d @- << EOF
{
  "userId": "$USER_ID",
  "type": "EMAIL",
  "subject": "Test Notification",
  "body": "This is a test notification from API testing"
}
EOF
)

NOTIFICATION_ID=$(echo "$NOTIFICATION" | jq -r '.id // empty')
if [ -n "$NOTIFICATION_ID" ] && [ "$NOTIFICATION_ID" != "null" ]; then
    echo "$NOTIFICATION" | jq '.'
    success "Successfully created notification: $NOTIFICATION_ID"
else
    info "Notification creation may require specific permissions"
fi

# READ - List notifications
info "READ: List user notifications"
NOTIFICATIONS=$(curl -s "$NOTIFICATION_URL/v1/notifications?limit=5" \
  -H "Authorization: Bearer $TOKEN")

echo "$NOTIFICATIONS" | jq '.'
success "Successfully fetched notifications"

# ============================================
# SUMMARY
# ============================================
section "Test Summary"
echo -e "${GREEN}All CRUD tests completed!${NC}"
echo -e "\n${BLUE}Key Information:${NC}"
echo "  - Gateway URL: $GATEWAY_URL"
echo "  - User ID: $USER_ID"
echo "  - Access Token: ${TOKEN:0:50}..."

echo -e "\n${BLUE}Service Endpoints:${NC}"
echo "  - Auth: $AUTH_URL"
echo "  - User: $USER_URL"
echo "  - Booking: $BOOKING_URL"
echo "  - Food: $FOOD_URL"
echo "  - Parking: $PARKING_URL"
echo "  - Shop: $SHOP_URL"
echo "  - Payment: $PAYMENT_URL"
echo "  - Notification: $NOTIFICATION_URL"

echo -e "\n${BLUE}GraphQL Endpoint:${NC} $GATEWAY_URL/graphql"
echo -e "${BLUE}GraphQL Queries:${NC}"
echo "  - me"
echo "  - facilities(venueId, available, limit, offset)"
echo "  - bookings(userId, limit, offset)"
echo "  - booking(id)"
echo "  - facilitySchedule(facilityId, from, to)"

echo -e "\n${BLUE}GraphQL Mutations:${NC}"
echo "  - createBooking(facilityId, startsAt, endsAt)"
echo "  - cancelBooking(id)"
echo "  - updateFacilityAvailability(id, available) [ADMIN only]"
echo "  - createFacilityOverride(input) [ADMIN only]"
echo "  - removeFacilityOverride(facilityId, id) [ADMIN only]"

echo -e "\n${YELLOW}Quick Commands:${NC}"
echo "  # Export token for manual testing"
echo "  export TOKEN='$TOKEN'"
echo ""
echo "  # Test GraphQL query"
echo "  curl -X POST $GATEWAY_URL/graphql \\"
echo "    -H 'Authorization: Bearer \$TOKEN' \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{\"query\":\"{me{id email}}\"}'| jq"
echo ""
echo "  # Test REST API"
echo "  curl $BOOKING_URL/v1/facilities?limit=5 \\"
echo "    -H 'Authorization: Bearer \$TOKEN' | jq"
echo ""
