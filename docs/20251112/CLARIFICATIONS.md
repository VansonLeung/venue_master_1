Critical Clarifications Needed
1. Timeline and Date Inconsistencies
Issue: Current date is Nov 2025, but Milestone 0 shows "Dec - Jan 2025" (past dates)
Milestone 1A-1C: Jan-May 2026 (only 6 months from now)
Final Delivery: Aug 1, 2026 (9 months total)
Question: Are these dates still realistic? Should the timeline be adjusted to start from current date (Nov 2025)?  >> Start from Dec 2025

2. Database Selection
Issue: Section 1.3 and 6.3 state "Open for suggestions"
Questions:
What are the database requirements (ACID compliance, scalability, etc.)?   >>> ACID compliance, scalability
Preference for SQL (PostgreSQL, MySQL) vs NoSQL (MongoDB, DynamoDB)?  >> SQL (PostgreSQL)
Will different modules use different databases as suggested in Section 3?   >> Yes
What about database migration tools and versioning strategy?   >> use Sequelize and the Sequelize CLI for migrations

3. Cloud Storage Provider
Issue: Section 1.4 says "TBA" with potential candidates (AWS S3, Google Cloud, Azure)  
Section 6.4: Later specifies AWS S3 definitively
Question: Is AWS S3 confirmed, or still open for discussion? >> AWS S3 confirmed

4. Architecture Contradictions
Section 1.2: Uses GraphQL (gqlgen) with gin-gonic
Section 3: States "RESTful APIs" for inter-module communication
Questions:
Will the system use GraphQL, REST, or both? >> Both
If both, what's the division? (GraphQL for client-facing, REST for inter-service?) >> GraphQL for client-facing, REST for inter-service
How will GraphQL directives like @hasRole work across microservices? >> Implement a gateway service that handles authentication and authorization, applying @hasRole directives before routing requests to respective microservices.

5. Microservices vs Monolith
Section 3: Mentions "microservices for each module" and "per-module databases"
Section 1.8: Backend must be "executable in a Docker runtime and brought up by docker-compose"
Questions:
How many separate services are planned? One per module (8 services)? >> Yes, one per module (8 services)
Database per service or shared database? >> Database per service
How will distributed transactions be handled (e.g., booking + payment)? >> Use Saga pattern for distributed transactions
What about service discovery, inter-service communication patterns? >> Use service mesh (e.g., Istio) for service discovery and communication

6. Authentication & Session Management
Issue: JWT is mentioned but implementation details are vague
Questions:
Where are JWT sessions stored (database, Redis, in-memory)? >> Redis
What's the JWT expiration strategy and refresh token approach? >> JWT expiration set to 15 minutes, refresh tokens valid for 7 days
How do other modules validate JWTs issued by the User System? >> Each module will have a shared library for JWT validation, ensuring consistent checks across services.
Is there a central auth service or distributed validation? >> Central auth service

7. Payment Integration
Section 2: Multiple mentions of "Payment integration (credit card / refund)"
Questions:
Which payment gateway/processor (Stripe, Square, PayPal, etc.)? >> Stripe
Will it support Canadian payment methods specifically? >> No
PCI DSS compliance requirements and approach? >> Follow PCI DSS SAQ A guidelines for handling card payments via Stripe
Handling of failed payments, chargebacks, reconciliation? >> Implement webhook listeners for payment status updates, automated retries for failed payments, and a reconciliation process using Stripe's reporting tools.
Currency support (CAD only or multi-currency)? >> Multi-currency with CAD as default

8. PIPEDA Compliance
Mentioned: "Security logging & compliance (PIPEDA)"
Questions:
Specific PIPEDA requirements to implement? >> Data minimization, purpose limitation, consent management
Data retention and deletion policies? >> Retain personal data for no longer than necessary, with automatic deletion after 2 years of inactivity
User consent management strategy? >> Implement consent banners and detailed privacy settings in user profiles
Data breach notification procedures?    >> Notify affected users within 72 hours of identifying a breach
Right to access/portability implementation?  >> Provide users with downloadable copies of their data upon request

9. Pickleball Analytics Module (OUT OF OUR SCOPE)
Section 2, Priority 5: "frontend" only mentioned, marked as "(partial)"
Questions:
Who provides the backend/AI for video analysis? >> ATTA team
Where are videos stored and processed?  >> ATTA team infrastructure
What's the data format for performance analytics? >> JSON format defined by ATTA
Is this module MVP or Phase 2? >> Phase 2

10.  ATTA App Integration.  (OUT OF OUR SCOPE)
Section 2, Priority 6: Mentions "Progress Tracker" and "T-Shots"
Questions:
Are these existing ATTA systems to integrate with? >> Yes
What APIs do they expose? >> RESTful APIs
Who maintains those systems? >> ATTA team
What's the authentication flow between systems? >> OAuth 2.0
Is ATTA providing documentation/sandbox for these integrations? >> Yes

11.  UI Layouts
User System: "Three UI layouts for different roles (Admin / Operator / Member)"
Questions:
Are these completely different UIs or same UI with different permissions? >> Same UI with different permissions
Should all three roles use the same Flutter app?  >> Yes
Any design mockups or wireframes available?   >> No
Responsive design requirements for desktop vs mobile? >> Yes, must be responsive

12.  Real-time Features
Food Ordering: "Real-time kitchen dashboard (order queue)"
Parking: "Real-time slot availability display"
Questions:
What real-time technology (WebSockets, Server-Sent Events, polling)?  >> WebSockets
Performance requirements (latency, concurrent users)?  >> Latency under 200ms, support for 1000 concurrent users
How does this work with GraphQL subscriptions? >> Use GraphQL subscriptions over WebSockets for real-time updates

13.  Notification System
Mentioned: Email and push notifications
Questions:
Push notification service (Firebase Cloud Messaging, APNs directly)? >> Firebase Cloud Messaging
Email service provider (SendGrid, AWS SES, Mailgun)? >> SendGrid
Notification preferences and opt-out management? >> User profile settings for notification preferences
In-app notification center? >> Yes, include an in-app notification center

14.  Testing & QA
Mentioned: "Internal UAT" and "full QA" but no details
Questions:
Who performs UAT (ATTA team or end users)? >> ATTA team
Test coverage requirements? >> Minimum 80% code coverage
Performance/load testing requirements? >> Yes, simulate 1000 concurrent users
Security testing and penetration testing? >> Yes, conduct regular security audits and penetration tests
Automated testing strategy (unit, integration, e2e)? >> Use automated unit and integration tests, with playwright-powered semi-automated e2e testing

15.  Deployment & DevOps
Section 1.8: Docker and docker-compose mentioned
Questions:
Production hosting platform (AWS, Google Cloud, Azure, on-premise)? >> AWS
CI/CD pipeline requirements? >> Use GitHub Actions for CI/CD
Monitoring and logging tools (Prometheus, ELK stack)? >> ELK stack
Backup and disaster recovery strategy? >> Daily backups, with a 30-day retention policy
Staging/production environment separation? >> Yes, separate staging and production environments

16.  Phase 2 Features
Scattered throughout: Various Phase 2 items mentioned
Questions:
Is Phase 2 in scope for the Aug 2026 delivery? >> No
Or is it a separate contract/timeline? >> Separate contract/timeline
Which Phase 2 features are must-have vs nice-to-have? >> Must-have: Loyalty Points System, Membership Tiers; Nice-to-have: Pickleball Analytics, ATTA App Integration

17.  Loyalty Points System
User System: "Membership tiers and loyalty point system (Phase 2)"
Questions:
Points earning rules (per booking, purchase, etc.)? >> Points earned per dollar spent and per booking
Points redemption mechanism? >> Points can be redeemed for discounts on future bookings or purchases
Points expiration policy? >> Points expire after 12 months of inactivity
Integration with payment system? >> Yes, integrate with payment system for redemption

18.  Admin Panel
Mentioned: Admin override, manual booking, reports
Questions:
Separate admin web portal or part of main app? >> Separate admin web portal
What reporting tools/dashboards are needed? >> Custom reporting dashboard with key metrics
Export formats (PDF, CSV, Excel)? >> CSV and Excel
Role hierarchy (Super Admin, Venue Admin, etc.)? >> Yes, Super Admin and Venue Admin roles

19.  Mobile App Distribution
Questions:
App Store and Google Play publishing handled by ATTA or contractor? >> Handled by ATTA
Beta testing via TestFlight/Firebase App Distribution? >> Yes, use TestFlight for iOS and Firebase App Distribution for Android
App store assets (screenshots, descriptions) - who provides? >> ATTA provides app store assets
Multiple environments in production (staging app vs prod app)? >> Yes, separate staging and production environments

20.  Source Code & Repository Access (ASSUME GITHUB)
Section 8: ATTA has right to review code and access Git repository
Questions:
Which Git platform (GitHub, GitLab, Bitbucket)? >> GitHub
Repository ownership transfer timing? >> Upon project completion
Code review process and approval workflow? >> Pull request reviews by ATTA team
Branch strategy and deployment process? >> Git flow with separate branches for features, staging, and production
ASSUME FOR NOW we do not need to care code ownership until project completion

21.  Interactive Map Feature
Facility Booking: "interactive map" mentioned
Pickleball Analytics: "Interactive map, chart, heat map, diagram"
Questions:
What map library (Google Maps, Mapbox, custom)? >> custom
What's being mapped exactly (venue layout, court positions)? >> Venue layout and court positions
Indoor/outdoor positioning requirements? >> not required

22.  Membership Management Details
User System: "Subscriptions, add, renew, expire"
Questions:
Subscription billing cycle (monthly, annual, custom)? >> Monthly and annual options 
Automatic renewal vs manual? >> Automatic renewal
Pro-rated refunds for early cancellation? >> Yes, pro-rated refunds available
Grace period for expired memberships? >>  Grace period of 30 days
Multiple membership types/tiers? >> Yes, multiple membership types and tiers available