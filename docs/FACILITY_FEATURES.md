# Facility Availability Enhancements (Phase A)

## Goals
- Allow admins to define overrides/blackout periods per facility
- Expose a merged schedule feed (base hours + overrides) via REST/GraphQL
- Keep RBAC consistent: ADMIN & VENUE_ADMIN manage overrides; OPERATOR read-only

## Schema Additions
```
CREATE TABLE facility_overrides (
    id UUID PRIMARY KEY,
    facility_id UUID REFERENCES facilities(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    open_at TIME,
    close_at TIME,
    all_day BOOLEAN NOT NULL DEFAULT FALSE,
    reason TEXT,
    applies_weekdays INTEGER[] NOT NULL DEFAULT ARRAY[0,1,2,3,4,5,6]
);
```
- `all_day=true` means facility closed (blackout) for the date range unless `open_at/close_at` provided.
- Weekday array uses Go `time.Weekday` indexing (0=Sunday).

## API Surface
- **REST** (booking-service)
  - `GET /v1/facilities/:id/schedule?from=2026-01-01&to=2026-01-07`
  - `POST /v1/facilities/:id/overrides`
  - `DELETE /v1/facilities/:id/overrides/:overrideId`
- **GraphQL (gateway)**
  - `facilitySchedule(facilityId: ID!, from: String!, to: String!): [FacilityScheduleDay!]!`
  - `createFacilityOverride(input: FacilityOverrideInput!): FacilityOverride!`
  - `removeFacilityOverride(facilityId: ID!, id: ID!): Boolean!`
  - `FacilityScheduleDay` exposes `{ date, closed, reason, slots { openAt, closeAt } }`
  - `FacilityOverride` mirrors REST payloads and automatically formats `startDate/endDate (YYYY-MM-DD)` plus `openAt/closeAt` strings (HH:mm)
  - `FacilityOverrideInput` expects:
    ```graphql
    input FacilityOverrideInput {
      facilityId: ID!
      startDate: String!
      endDate: String!
      allDay: Boolean
      openAt: String
      closeAt: String
      reason: String
      appliesWeekdays: [Int!]
    }
    ```
    `appliesWeekdays` defaults to all days `[0-6]` when omitted.

## Schedule Response
```
{
  date: "2026-01-02",
  slots: [
    { openAt:"06:00", closeAt:"12:00" },
    { openAt:"14:00", closeAt:"20:00" }
  ],
  closed: false,
  reason: null
}
```
- If blackout, `closed=true` and slots empty.

## Testing
- `scripts/test-e2e.sh` now spins up docker-compose, runs `go test -tags e2e ./test/e2e`, and exercises auth → booking → admin override (create + remove) via GraphQL to guard the full path.

## Implementation Outline
1. Migration + store helpers (CRUD + schedule merge)
2. REST handlers with RBAC (reuse middleware)
3. Service interfaces/HTTP clients + GraphQL schema updates
