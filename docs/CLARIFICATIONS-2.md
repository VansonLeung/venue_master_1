1. RBAC expectations – Which roles should control each booking-service endpoint? (e.g., ADMIN vs VENUE_ADMIN vs OPERATOR). >> ADMIN, VENUE_ADMIN for all endpoints; OPERATOR only for read endpoints.

Do you want booking creation restricted to the JWT subject, or can admins impersonate other users?  >> Admins can impersonate other users for booking creation.


2. Pricing model – Do facilities charge per hour, per slot, or variable rates (weekday/weekend)?  >> Facilities charge per hour, with different rates for weekday and weekend.

3. Payment retry semantics – How aggressive should the retry queue be (max attempts/backoff), and do you want notification-service calls immediately on the first failure or only after the retries exhaust? >> Max 5 attempts with exponential backoff starting at 1 minute. Notify notification-service only after all retries are exhausted.

4. Pagination defaults – For GraphQL filtering/pagination, do you have target page sizes or max limits? >> Default page size of 20, max limit of 100.

5. Integration test target – Should the automated flow run inside CI (spinning up docker compose) or just live as a Postman collection you can run manually?  >> It should run inside CI using docker compose.