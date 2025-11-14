# Venue Master Platform

A comprehensive venue management and booking platform built with Go microservices, GraphQL API Gateway, and React Admin CMS.

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend Admin CMS (React)                │
│                     http://localhost:3001                    │
└──────────────────────────┬──────────────────────────────────┘
                           │ HTTP/REST
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              API Gateway (GraphQL + REST Proxy)              │
│                     http://localhost:8080                    │
│  • JWT Authentication                                        │
│  • GraphQL Endpoint: /graphql                                │
│  • REST Proxy: /v1/*                                         │
└──────────────────────────┬──────────────────────────────────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
         ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ Auth Service │  │User Service  │  │Booking Svc   │  ...
│   :8081      │  │   :8082      │  │   :8083      │
└──────────────┘  └──────────────┘  └──────┬───────┘
                                            │
                                            ▼
                                    ┌──────────────┐
                                    │ PostgreSQL   │
                                    │   :15432     │
                                    └──────────────┘
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.23+** for backend services
- **Node.js 18+** for frontend
- **Docker & Docker Compose** for local development
- **PostgreSQL 15+** (via Docker)
- **Redis 7+** (via Docker)

### 1. Start Backend Services

```bash
cd codes
docker-compose up -d
```

This starts all microservices:
- API Gateway (port 8080)
- Auth Service (port 8081)
- User Service (port 8082)
- Booking Service (port 8083)
- Food Service (port 8084)
- Parking Service (port 8085)
- Shop Service (port 8086)
- Payment Service (port 8087)
- Notification Service (port 8088)
- PostgreSQL (port 15432)
- Redis (port 6379)

### 2. Start Frontend Admin CMS

```bash
cd frontend_codes/admin_cms
npm install
npm run dev
```

Access the admin interface at **http://localhost:3001**

### 3. Login

**Default Credentials:**
- **Admin**: `admin@example.com` / `Admin123!`
- **Member**: `member@example.com` / `Secret123!`

## 📚 Documentation

### Core Documentation
- **[codes/README.md](codes/README.md)** - Backend services, API examples, GraphQL queries
- **[VENUE_IMPLEMENTATION.md](VENUE_IMPLEMENTATION.md)** - Technical venue implementation details
- **[VENUE_IMPLEMENTATION_SUMMARY.md](VENUE_IMPLEMENTATION_SUMMARY.md)** - User-friendly implementation guide
- **[API_TESTING_GUIDE.md](API_TESTING_GUIDE.md)** - API testing documentation
- **[docs/FACILITY_FEATURES.md](docs/FACILITY_FEATURES.md)** - Facility features and overrides

### Key Features Documentation
- Venue Management (this document - see below)
- Facility Management with venue association
- Booking lifecycle and payment flow
- Schedule overrides and blackouts
- Role-based access control (RBAC)

## 🏢 Venue Management

The platform supports comprehensive venue management with full CRUD operations.

### Features

✅ **Complete CRUD Operations**
- Create, read, update, delete venues
- Timezone-aware configuration
- Full contact information management
- Address and location data

✅ **Venue-Facility Relationship**
- One venue contains many facilities
- Cascade deletion (deleting venue removes all facilities)
- Required venue association for all facilities

✅ **Role-Based Access Control**
- **READ** (All authenticated users): List venues, get venue details
- **WRITE** (ADMIN/VENUE_ADMIN only): Create, update, delete venues

✅ **NULL Handling**
- Proper database NULL handling for optional fields
- Clean API responses with empty strings for missing data

### Database Schema

```sql
-- Venues table (parent)
CREATE TABLE venues (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,           -- nullable
    address TEXT,               -- nullable
    city TEXT,                  -- nullable
    state TEXT,                 -- nullable
    zip_code TEXT,              -- nullable
    country TEXT DEFAULT 'US',
    phone TEXT,                 -- nullable
    email TEXT,                 -- nullable
    website TEXT,               -- nullable
    timezone TEXT DEFAULT 'America/New_York',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Facilities table (child)
CREATE TABLE facilities (
    id UUID PRIMARY KEY,
    venue_id UUID NOT NULL REFERENCES venues(id) ON DELETE CASCADE,
    -- ... other facility fields
);
```

### API Examples

**List All Venues** (authenticated users):
```bash
TOKEN="your-jwt-token"

curl -X GET "http://localhost:8080/v1/venues?limit=100" \
  -H "Authorization: Bearer $TOKEN"
```

**Create Venue** (admin only):
```bash
curl -X POST "http://localhost:8080/v1/venues" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Downtown Sports Complex",
    "description": "Modern multi-sport facility",
    "address": "123 Main Street",
    "city": "New York",
    "state": "NY",
    "zipCode": "10001",
    "country": "US",
    "phone": "+1-212-555-1234",
    "email": "info@downtown-sports.com",
    "website": "https://downtown-sports.com",
    "timezone": "America/New_York"
  }'
```

**Update Venue** (admin only):
```bash
curl -X PUT "http://localhost:8080/v1/venues/{venue-id}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Venue Name",
    "city": "Brooklyn",
    "timezone": "America/New_York"
  }'
```

**Delete Venue** (admin only - cascades to facilities):
```bash
curl -X DELETE "http://localhost:8080/v1/venues/{venue-id}" \
  -H "Authorization: Bearer $TOKEN"
```

### Admin CMS Usage

1. **Login** as admin (`admin@example.com` / `Admin123!`)
2. **Navigate** to "Venues" in the sidebar
3. **Create** new venue with the "Add Venue" button
4. **Edit** existing venues by clicking the edit icon
5. **Delete** venues with the delete icon (confirms before deletion)

When creating facilities, you'll now see a **venue selector dropdown** - select the venue before creating the facility.

## 🧪 Testing

### Run All Tests

```bash
# Backend unit tests
cd codes
make test

# E2E integration tests
./scripts/test-e2e.sh

# Comprehensive API testing (all services)
./scripts/test-api.sh
```

### API Test Coverage

The `test-api.sh` script validates:
- ✅ Health checks for all services
- ✅ Authentication (login, refresh, logout)
- ✅ User management
- ✅ **Venue CRUD operations**
- ✅ **Facility creation with venue association**
- ✅ Booking lifecycle (GraphQL + REST)
- ✅ Schedule and override management
- ✅ Food, parking, shop, payment, notification services
- ✅ Role-based access control
- ✅ Cascade deletion behavior

## 📁 Repository Structure

```
venue_master/
├── codes/                          # Backend microservices (Go)
│   ├── docker-compose.yml          # Local orchestration
│   ├── services/
│   │   ├── api-gateway/           # GraphQL + REST proxy
│   │   ├── auth-service/          # JWT authentication
│   │   ├── user-service/          # User profiles & RBAC
│   │   ├── booking-service/       # Venues, facilities, bookings
│   │   │   └── internal/store/
│   │   │       └── migrations/
│   │   │           └── 0004_venues.sql  # Venue schema migration
│   │   ├── food-service/          # Menu management
│   │   ├── parking-service/       # Parking reservations
│   │   ├── shop-service/          # Pro shop catalog
│   │   ├── payment-service/       # Stripe integration
│   │   └── notification-service/  # Email/push notifications
│   └── lib/                       # Shared Go packages
│
├── frontend_codes/
│   └── admin_cms/                 # React Admin CMS
│       ├── src/
│       │   ├── pages/
│       │   │   ├── VenuesPage.jsx        # Venue management UI
│       │   │   └── FacilitiesPage.jsx    # Facility management UI
│       │   ├── services/
│       │   │   ├── venue.service.js      # Venue API client
│       │   │   └── facility.service.js   # Facility API client
│       │   └── components/
│       │       └── Layout.jsx            # Navigation with Venues link
│       └── package.json
│
├── scripts/
│   ├── test-api.sh                # Comprehensive API testing
│   └── test-e2e.sh                # E2E integration tests
│
├── docs/                          # Architecture & planning docs
│   ├── PRD.md                     # Product requirements
│   ├── PLAN.md                    # Implementation plan
│   └── FACILITY_FEATURES.md       # Facility features guide
│
├── README.md                      # This file
├── VENUE_IMPLEMENTATION.md        # Technical implementation details
└── VENUE_IMPLEMENTATION_SUMMARY.md # Implementation summary
```

## 🔧 Development

### Backend Development

```bash
cd codes

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f booking-service

# Rebuild a service
docker-compose up -d --build booking-service

# Run migrations
# Migrations run automatically on service startup
```

### Frontend Development

```bash
cd frontend_codes/admin_cms

# Install dependencies
npm install

# Start dev server
npm run dev

# Build for production
npm run build
```

### Database Access

```bash
# Connect to PostgreSQL
psql -h localhost -p 15432 -U postgres -d venue_master

# View venues
SELECT id, name, city, state FROM venues;

# View facilities with venue info
SELECT f.id, f.name, v.name as venue_name
FROM facilities f
JOIN venues v ON f.venue_id = v.id;
```

## 🏗️ Technology Stack

### Backend
- **Language**: Go 1.23
- **API**: GraphQL (gqlgen) + REST
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Auth**: JWT tokens
- **Infrastructure**: Docker Compose

### Frontend
- **Framework**: React 18 + Vite
- **Routing**: React Router v6
- **UI Components**: Shadcn/ui + Tailwind CSS
- **State**: React Context API
- **HTTP Client**: Axios
- **Icons**: Lucide React

## 🔐 Security & RBAC

### User Roles
- **MEMBER**: Read-only access to venues and facilities
- **OPERATOR**: Facility management
- **VENUE_ADMIN**: Full venue and facility management
- **ADMIN**: Full system access

### Authentication Flow
1. User logs in via `/v1/auth/login`
2. Receives JWT access token (expires in 15 min)
3. Includes `Authorization: Bearer <token>` in requests
4. Gateway validates JWT and extracts user ID + roles
5. Gateway adds `X-User-ID` and `X-User-Roles` headers
6. Services enforce role-based access control

## 📊 Service Endpoints

| Service | Port | Health Check | Primary Function |
|---------|------|-------------|------------------|
| API Gateway | 8080 | `/healthz` | GraphQL + REST proxy |
| Auth Service | 8081 | `/healthz` | JWT authentication |
| User Service | 8082 | `/healthz` | User profiles & RBAC |
| Booking Service | 8083 | `/healthz` | Venues & facilities |
| Food Service | 8084 | `/healthz` | Menu management |
| Parking Service | 8085 | `/healthz` | Parking reservations |
| Shop Service | 8086 | `/healthz` | Pro shop catalog |
| Payment Service | 8087 | `/healthz` | Stripe integration |
| Notification Service | 8088 | `/healthz` | Notifications |

## 🐛 Troubleshooting

### Services won't start
```bash
# Check Docker status
docker-compose ps

# View logs
docker-compose logs

# Restart all services
docker-compose restart
```

### Database connection errors
```bash
# Verify PostgreSQL is running
docker-compose ps postgres

# Check connection
psql -h localhost -p 15432 -U postgres -d venue_master
```

### Frontend can't connect to API
```bash
# Verify API Gateway is running
curl http://localhost:8080/healthz

# Check frontend env variables
cat frontend_codes/admin_cms/.env
```

### Venue API returns NULL errors
This has been **fixed** in the latest version. The booking service now properly handles NULL values in venue fields.

## 📄 License

This project is proprietary software for Venue Master Platform.

## 🤝 Contributing

1. Create a feature branch
2. Make your changes
3. Run tests: `./scripts/test-api.sh`
4. Submit a pull request

## 📞 Support

For issues and questions:
- Check [codes/README.md](codes/README.md) for API documentation
- Review [VENUE_IMPLEMENTATION.md](VENUE_IMPLEMENTATION.md) for technical details
- Run `./scripts/test-api.sh` to validate your setup

---

**Version**: 1.0.0
**Last Updated**: 2025-11-14
**Status**: ✅ Production Ready
