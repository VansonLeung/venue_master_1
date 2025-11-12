1. RBAC expectations – Which roles should control each booking-service endpoint? (e.g., ADMIN vs VENUE_ADMIN vs OPERATOR). Do you want booking creation restricted to the JWT subject, or can admins impersonate other users? I’ll wire the enforcement both in the gateway and the service once I know the exact matrix.

2. Pricing model – Do facilities charge per hour, per slot, or variable rates (weekday/weekend)? I plan to add hourly_rate_cents + currency columns and compute the booking amount from the facility’s duration, but let me know if there’s a more nuanced rule set before I migrate the DB.

3. Payment retry semantics – How aggressive should the retry queue be (max attempts/backoff), and do you want notification-service calls immediately on the first failure or only after the retries exhaust? I’ll build a worker that hits payment-service again and emits notifications, but I need the policy.

4. Pagination defaults – For GraphQL filtering/pagination, do you have target page sizes or max limits? I’m planning limit/offset arguments plus an optional available filter for facilities.

5. Integration test target – Should the automated flow run inside CI (spinning up docker compose) or just live as a Postman collection you can run manually? I can add a Go e2e test that expects the services to be up, but if you prefer Playwright/Postman artifacts let me know.