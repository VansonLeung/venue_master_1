# Functions and Features - Overall Summary

## App & System Development Overview for Outsourcing Evaluation

**Document Status:** Updated with technical clarifications (Nov 2025)

### Key Technical Decisions Summary

This document has been updated with the following key technical decisions:

- **Timeline:** Project starts Dec 2025, final delivery Aug 1, 2026 (9 months)
- **Database:** PostgreSQL with database-per-service architecture, Sequelize ORM for migrations
- **Cloud Storage:** AWS S3 (confirmed)
- **API Architecture:** GraphQL for client-facing (gqlgen), REST for inter-service communication
- **Authentication:** JWT with Redis session storage (15-min expiration, 7-day refresh tokens)
- **Microservices:** 8 separate services with Saga pattern for distributed transactions, Istio for service mesh
- **Payment:** Stripe (multi-currency, CAD default, PCI DSS SAQ A compliance)
- **Real-time:** WebSockets with GraphQL subscriptions (<200ms latency, 1000 concurrent users)
- **Notifications:** SendGrid (email), Firebase Cloud Messaging (push)
- **PIPEDA Compliance:** Data minimization, 2-year retention, 72-hour breach notification
- **Hosting:** AWS with GitHub Actions CI/CD, ELK stack monitoring
- **Testing:** 80% code coverage, Playwright e2e tests, load testing for 1000 concurrent users
- **Admin Interface:** Separate admin web portal for administrative functions
- **Mobile Distribution:** TestFlight (iOS), Firebase App Distribution (Android), ATTA handles publishing
- **Repository:** GitHub with Git flow, code ownership transfers upon project completion
- **Phase 2 (Separate Contract):** Loyalty points system (must-have), Pickleball Analytics & ATTA Integration (nice-to-have)

### 1. Core Technologies & Architecture
1.1 **Language**
- **Server:** Go (Golang)
- **App & Web:** Flutter (Dart)

1.2 **API Layer**
- **Client-facing API:** GraphQL using gqlgen with GraphQL subscriptions over WebSockets for real-time updates
- **Inter-service Communication:** RESTful APIs for microservice-to-microservice communication
- **HTTP Framework:** gin-gonic/gin
- **GraphQL Directives:** Authorization logic implemented using GraphQL directives like `@hasRole`, keeping resolver logic focused on business tasks.
- **API Gateway:** Gateway service handles authentication, authorization, and request routing to microservices

1.3 **Database**
- **Primary Database:** PostgreSQL (SQL database)
- **Requirements:** ACID compliance, scalability
- **Architecture:** Database per service (one database per microservice module)
- **Migration Tools:** Sequelize ORM and Sequelize CLI for database migrations and versioning

1.4 **Cloud Storage**
- **Abstraction Layer:** An interface defines a clean abstraction for cloud storage operations.
- **Storage Provider:** AWS S3 (Simple Storage Service) - Confirmed

1.5 **Authentication & Authorization**
- **JWT (JSON Web Tokens):** User authentication is managed via JWTs. Upon successful login, a session is created and stored in Redis, and a JWT is issued to the client for authenticating subsequent requests.
- **Session Storage:** Redis (for JWT session management)
- **Token Strategy:**
  - JWT expiration: 15 minutes
  - Refresh tokens: Valid for 7 days
- **Architecture:** Central authentication service with shared JWT validation library across all microservices
- **Access Control (AC):** Features a comprehensive AC model. Users are assigned roles (e.g., Member, Operator, SysOp) and permissions, with access to GraphQL operations restricted accordingly.
- **Cross-Service Auth:** Gateway service handles authentication and authorization, applying @hasRole directives before routing requests to respective microservices.

1.6 **Other Packages & Services**
- **Email Service:** SendGrid for transactional emails (e.g., registration confirmation, login credentials).
- **Push Notifications:** Firebase Cloud Messaging (FCM) for mobile push notifications
- **In-App Notifications:** Notification center within the app
- **Real-time Communication:** WebSockets for real-time features (e.g., kitchen dashboard, parking availability)
- **Performance Requirements:** Latency under 200ms, support for 1000 concurrent users
- **Configuration:** Managed through environment variables and config files.

1.7 **Supported Platforms**
- **iOS:** iOS 16+
- **Android:** Android 13+
- **Web Browsers:**
  - iOS: Safari, Chrome
  - Android: Chrome
  - Desktop: Microsoft Edge, Chrome, Safari, Firefox

1.8 **Backend Execution Environment**
- The backend artifact must be executable in a Docker runtime and brought up by docker-compose per the defined compose configuration.
- **Microservices Architecture:** 8 separate services (one per module), each with its own database
- **Distributed Transactions:** Saga pattern for handling cross-service transactions (e.g., booking + payment)
- **Service Discovery:** Service mesh (e.g., Istio) for service discovery and inter-service communication
- **Hosting Platform:** AWS (Amazon Web Services)
- **CI/CD:** GitHub Actions for continuous integration and deployment
- **Monitoring & Logging:** ELK stack (Elasticsearch, Logstash, Kibana)
- **Backup Strategy:** Daily automated backups with 30-day retention policy
- **Environments:** Separate staging and production environments

### 2. Core Modules

| Priority | Module Description                                   | Key Requirements / Features                                                                                   |
|----------|-----------------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| 1        | **Cross-System Integration**                         | Core infrastructure connecting all subsystems through unified APIs. Handles authentication, payment, notifications, and analytics. |
|          |                                                     | - API Gateway setup and endpoint structure                                                                     |
|          |                                                     | - Centralized authentication (JWT / SSO)                                                                       |
|          |                                                     | - Payment service (charge, refund, transaction log)                                                           |
|          |                                                     | - Notification service (email, push)                                                                           |
|          |                                                     | - Feedback & error handling APIs                                                                                |
| 2        | **User System**                                     | Foundation for user management and access control. Provides unified identity and role-based permissions.      |
|          |                                                     | - User registration & login (email / mobile)                                                                  |
|          |                                                     | - Forgot password & verification flow                                                                           |
|          |                                                     | - Role-based access control (Admin / Operator / Member)                                                       |
|          |                                                     | - Profile management & session tracking                                                                         |
|          |                                                     | - Membership management (Subscriptions: monthly/annual, automatic renewal, add, renew, expire)                |
|          |                                                     | - Membership grace period: 30 days after expiration                                                           |
|          |                                                     | - Pro-rated refunds for early cancellation                                                                     |
|          |                                                     | - Multiple membership types and tiers                                                                          |
|          |                                                     | - UI: Same Flutter app with role-based permission views (Admin / Operator / Member)                          |
|          |                                                     | - Responsive design for desktop and mobile                                                                     |
|          |                                                     | - Security logging & compliance (PIPEDA)                                                                       |
|          |                                                     | - PIPEDA: Data minimization, purpose limitation, consent management                                           |
|          |                                                     | - Data retention: 2 years of inactivity, then automatic deletion                                               |
|          |                                                     | - User consent: Banners and privacy settings in user profiles                                                  |
|          |                                                     | - Data breach notification: Within 72 hours                                                                    |
|          |                                                     | - Right to access/portability: Downloadable data upon request                                                 |
|          |                                                     | - Payment integration: Stripe (multi-currency with CAD default)                                               |
|          |                                                     | - PCI DSS: SAQ A compliance via Stripe                                                                         |
|          |                                                     | - Payment handling: Webhook listeners, automated retries, reconciliation via Stripe reporting                 |
|          |                                                     | - Membership tiers and loyalty point system (Phase 2 - Must-have)                                             |
| 3        | **Facility Booking System (with Payment)**          | Allows members to reserve facilities and handle payments. Integrates with Membership and Payment APIs.       |
|          |                                                     | - Facility list and time-slot selection (custom interactive map for venue layout and court positions)         |
|          |                                                     | - Interactive map: Custom implementation (no Google Maps/Mapbox), indoor positioning not required             |
|          |                                                     | - Booking creation, modification & cancellation                                                                |
|          |                                                     | - Payment integration: Stripe (credit card / refund)                                                          |
|          |                                                     | - Admin override & manual booking control via separate admin web portal                                       |
|          |                                                     | - Booking reports & usage analytics (export as CSV and Excel)                                                 |
| 4        | **Food Ordering System**                            | In-app food menu and ordering system with operational dashboard for staff.                                     |
|          |                                                     | - Digital menu management (categories, prices)                                                                |
|          |                                                     | - Payment integration: Stripe (credit card / refund)                                                          |
|          |                                                     | - Order creation, update & status tracking                                                                     |
|          |                                                     | - Real-time kitchen dashboard (order queue) using WebSockets                                                  |
|          |                                                     | - Real-time performance: <200ms latency, 1000 concurrent users                                                |
|          |                                                     | - Promotional pricing & discounts                                                                                |
|          |                                                     | - Order reports & sales summary (export as CSV and Excel)                                                     |
|          |                                                     | - POS / printer integration (Phase 2)                                                                          |
| 5        | **Pickleball Analytics (frontend)**                 | Match Video Analysis and Performance Insights System - Phase 2 / OUT OF SCOPE                                 |
|          |                                                     | - List on-site captured videos (partial)                                                                        |
|          |                                                     | - Video playback & download                                                                                   |
|          |                                                     | - Tagging / categorizing match clips                                                                           |
|          |                                                     | - Visualize player performance analytics (Interactive map, chart, heat map, diagram, etc.)                   |
|          |                                                     | - Compare performance data across two matches                                                                  |
|          |                                                     | - Backend/AI: Provided by ATTA team                                                                            |
|          |                                                     | - Video storage: ATTA team infrastructure                                                                      |
|          |                                                     | - Data format: JSON format defined by ATTA                                                                     |
| 6        | **ATTA App Integration**                            | Embeds Progress Tracker & T-Shots video playback into the main app - Phase 2 / OUT OF SCOPE                   |
|          |                                                     | - Show member QR Code (serves as a launch key for ATTA's Progress Tracker and T-Shots)                       |
|          |                                                     | - API for retrieving user identity from member QR code                                                        |
|          |                                                     | - List, playback, download assessment records, video                                                          |
|          |                                                     | - Secure token-based data exchange via API                                                                     |
|          |                                                     | - Existing ATTA systems: Progress Tracker and T-Shots                                                         |
|          |                                                     | - ATTA APIs: RESTful APIs maintained by ATTA team                                                             |
|          |                                                     | - Authentication flow: OAuth 2.0                                                                               |
|          |                                                     | - ATTA documentation/sandbox: Provided by ATTA team                                                           |
|          |                                                     | - Token system for T-Shots (Phase 2)                                                                          |
| 7        | **Premium Car Parking**                             | Parking management module sharing similar logic to facility booking.                                          |
|          |                                                     | - Parking slot reservation & time-based pricing                                                                |
|          |                                                     | - Real-time slot availability display using WebSockets                                                        |
|          |                                                     | - Real-time performance: <200ms latency, 1000 concurrent users                                                |
|          |                                                     | - Booking creation / cancellation / refund                                                                      |
|          |                                                     | - Payment API integration: Stripe                                                                             |
|          |                                                     | - Usage analytics (export as CSV and Excel)                                                                   |
| 8        | **Pro Shop Ordering System**                        | Online shop for merchandise and equipment sales.                                                              |
|          |                                                     | - Product catalog (images, stock)                                                                               |
|          |                                                     | - Cart & checkout process                                                                                      |
|          |                                                     | - Order management (status, refund)                                                                            |
|          |                                                     | - Payment integration: Stripe                                                                                  |
|          |                                                     | - Order reports (export as CSV and Excel)                                                                     |
|          |                                                     | - Discount code & promotions (Phase 2)                                                                         |

### 3. Technical Overview

**Architecture:**
- **Microservices:** 8 separate backend services (one per module), each with its own PostgreSQL database
- **Client Communication:** GraphQL API (gqlgen) for all client-facing operations (web app, mobile app)
- **Inter-Service Communication:** RESTful APIs for microservice-to-microservice communication
- **API Gateway:** Central gateway service handles authentication, authorization (@hasRole directives), and request routing
- **Service Mesh:** Istio for service discovery, load balancing, and inter-service communication
- **Distributed Transactions:** Saga pattern for handling cross-service transactions (e.g., booking + payment)

**Authentication & Authorization:**
- Centralized authentication service issues JWTs stored in Redis
- JWT expiration: 15 minutes, refresh tokens: 7 days
- Shared JWT validation library across all microservices
- Permission control and operation logging (PIPEDA compliance)
- Role hierarchy: Member, Operator, Admin, Venue Admin, Super Admin

**Payment & Financial:**
- Unified Stripe payment/refund API across all modules
- Multi-currency support with CAD as default
- PCI DSS SAQ A compliance
- Automated webhook handling, retries, and reconciliation

**Reporting & Analytics:**
- Sales, facility usage, and user activity reporting
- Export formats: CSV and Excel
- Custom reporting dashboard in admin portal

**Admin Functions:**
- Separate admin web portal for all administrative operations
- Manual overrides, booking management, user management, reporting

**Performance & Scale:**
- Support 1000 concurrent users
- Real-time features: <200ms latency via WebSockets
- Daily automated backups with 30-day retention

**Phase 2 Enhancements (Separate Contract):**
- Must-have: Loyalty program with membership tiers
- Nice-to-have: Pickleball Analytics frontend, ATTA App Integration
- Additional: POS/printer integration, advanced discount engine

### 3a. Loyalty Points System (Phase 2 - Must-have)
- **Points Earning Rules:**
  - Points earned per dollar spent on bookings, food orders, parking, and pro shop purchases
  - Bonus points per booking completion
- **Points Redemption:**
  - Points can be redeemed for discounts on future bookings or purchases
  - Integration with payment system for seamless redemption at checkout
- **Points Expiration:**
  - Points expire after 12 months of inactivity
  - Notification system to alert users before expiration
- **Membership Tiers:**
  - Multiple membership tiers with different benefits
  - Tier upgrades based on points accumulation or spend thresholds

### 4. Development Scope Summary
- **Frontend:** Web and Mobile App (iOS + Android), desktop + mobile layout, including admin, operator functions.
- **Backend:** API Gateway, microservices for each module, and per-module databases.
- **Integration:** Authentication, Payment, Notification, and Analytics modules.

#### Deliverables
- **Phase 1 (Aug 2026):** Membership System, Facility Booking System, Food Ordering System, Premium Car Parking, Pro Shop Ordering System.
- **Phase 2 (Separate contract/timeline):** 
  - Must-have: Membership tiers and loyalty point system
  - Nice-to-have: Pickleball Analytics (frontend integration), ATTA App Integration
  - Additional: Food Ordering System POS/printer integration; Pro Shop Ordering System discount code & promotions.

### 5. Supported Platform
| Supported Platform | Requirements           | Deliverables                |
|--------------------|------------------------|-----------------------------|
| iOS                | iOS 16+                |                             |
| Android            | Android 13+            |                             |
| Web Browser         | Desktop: Microsoft Edge, Chrome, Safari, Firefox; iOS: Safari, Chrome; Android: Chrome |                             |

### 6. Development Environment

#### Core Technologies & Architecture
6.1 **Language**
- **Server:** Go (Golang)
- **App & Web:** Flutter (Dart)

6.2 **API Layer**
- **GraphQL API:** gqlgen
- **HTTP Framework:** gin-gonic/gin
- **GraphQL Directives:** Authorization logic implemented using GraphQL directives like `@hasRole`.

6.3 **Database**
- **Primary Database:** PostgreSQL (SQL database)
- **Requirements:** ACID compliance, scalability
- **Architecture:** Database per service (one database per microservice module)
- **Migration Tools:** Sequelize ORM and Sequelize CLI for database migrations and versioning

6.4 **Cloud Storage**
- **Abstraction Layer:** An interface defines a clean abstraction for cloud storage operations.
- **Storage Provider:** AWS S3 (Simple Storage Service) - Confirmed

6.5 **Authentication & Authorization**
- **JWT (JSON Web Tokens):** User authentication managed via JWTs. Upon successful login, a session is created and stored in Redis.
- **Session Storage:** Redis (for JWT session management)
- **Token Strategy:**
  - JWT expiration: 15 minutes
  - Refresh tokens: Valid for 7 days
- **Architecture:** Central authentication service with shared JWT validation library across all microservices
- **Access Control (AC):** Comprehensive AC model with roles (e.g., Member, Operator, SysOp, Super Admin, Venue Admin).
- **Cross-Service Auth:** Gateway service handles authentication and authorization, applying @hasRole directives before routing requests to respective microservices.

6.6 **Other Packages & Services**
- **Email Service:** SendGrid for transactional emails (e.g., registration confirmation, login credentials).
- **Push Notifications:** Firebase Cloud Messaging (FCM) for mobile push notifications
- **In-App Notifications:** Notification center within the app with user preference management
- **Real-time Communication:** WebSockets for real-time features (GraphQL subscriptions over WebSockets)
- **Performance Requirements:** Latency under 200ms, support for 1000 concurrent users
- **Payment Service:** Stripe for credit card processing, refunds, and multi-currency support (CAD as default)
- **PCI DSS Compliance:** SAQ A guidelines for handling card payments via Stripe
- **Payment Features:** Webhook listeners for payment status updates, automated retries for failed payments, reconciliation process using Stripe's reporting tools
- **Configuration:** Managed through environment variables loaded at runtime using a custom utility in osutil.

### 7. Milestone Plan

| Milestone                    | Target Date        | Deliverables / Goals                                                                    |
|------------------------------|--------------------|----------------------------------------------------------------------------------------|
| Milestone 0 – Project Kickoff| Dec 2025 - Jan 2026 | Confirm technical stack and architecture plan; set up GitHub repositories and communication channels; define API Gateway framework for Cross-System Integration; Setup AWS hosting, CI/CD pipeline (GitHub Actions), and ELK stack monitoring |
| Milestone 1A – Cross-System Integration (Backend Foundation) | Feb - Mar 2026 | API Gateway core implementation with GraphQL (client-facing) and REST (inter-service); Authentication (JWT with Redis) ready; Stripe payment API base structure; Logging and notification framework (SendGrid, FCM); WebSocket infrastructure for real-time features |
| Milestone 1B – User System   | Mar - Apr 2026     | Registration/login/password reset; Role-based access control (Member, Operator, Admin, Super Admin, Venue Admin); Profile management and data logging; PIPEDA compliance implementation (consent management, data retention policies, breach notification procedures); Integration with API Gateway; Separate admin web portal |
| Milestone 1C – Facility Booking System (with Payment) | Apr – May 2026 | Facility booking and management with custom interactive map; Stripe payment/refund integration; Admin override via admin portal; Usage reporting (CSV/Excel export); Testing with 1000 concurrent users |
| Milestone 1 – MVP Completion | May 2026           | Delivery of MVP (3 core modules): Cross-System Integration, User System, Facility Booking System (with Payment); Includes backend microservices integration, admin portal, testing environment (80% code coverage); Beta distribution via TestFlight (iOS) and Firebase App Distribution (Android) |
| Milestone 2A – Food Ordering System | May – Jun 2026 | Menu management and order workflow; Real-time kitchen dashboard with WebSockets (<200ms latency); Order dashboard for kitchen/staff; Promotion/discount features; Stripe integration; Order reports (CSV/Excel export) |
| Milestone 2B – Premium Card Parking & Pro Shop System | Jun – Jul 2026 | Parking reservation module with Stripe payment and real-time availability (WebSockets); Pro Shop ordering and checkout system with Stripe; Product catalog management; Reports and analytics (CSV/Excel export) |
| Milestone 2C – Integration Testing & QA | Jul 2026 | Internal UAT by ATTA team across all modules; Performance and load testing (1000 concurrent users); Security audits and penetration testing; Playwright-powered semi-automated e2e testing; Final integration tests; Daily backups and disaster recovery testing |
| Milestone 2 – Final Delivery (Completion) | Aug 1, 2026 | Final handover including: User System, Facility Booking System, Food Ordering System, Premium Card Parking, Pro Shop Ordering System; Performance optimization, full QA (80% code coverage), comprehensive documentation, complete source code (front-end + back-end), Docker/docker-compose configuration, deployment to AWS production environment. **Note:** Pickleball Analytics and ATTA App Integration moved to Phase 2 (out of scope for Aug 2026 delivery). |

### 8. Testing & Quality Assurance

**Testing Strategy:**
- **Unit Tests:** Minimum 80% code coverage requirement
- **Integration Tests:** Automated integration tests for all microservices and API endpoints
- **End-to-End Tests:** Playwright-powered semi-automated e2e testing for critical user flows
- **Performance Testing:** Load testing to simulate 1000 concurrent users with <200ms latency requirement
- **Security Testing:** Regular security audits and penetration testing throughout development

**User Acceptance Testing (UAT):**
- Conducted by ATTA team
- Internal UAT across all modules before final delivery
- Beta testing via TestFlight (iOS) and Firebase App Distribution (Android)

**Quality Metrics:**
- 80% minimum code coverage
- All critical bugs resolved before milestone completion
- Performance benchmarks met (1000 concurrent users, <200ms latency)
- Security vulnerabilities addressed per penetration test findings

### 9. Mobile App Distribution & Deployment

**Beta Testing:**
- iOS: TestFlight for beta distribution
- Android: Firebase App Distribution for beta testing
- Separate staging and production app builds

**Production Release:**
- App Store (iOS) and Google Play (Android) publishing handled by ATTA
- App store assets (screenshots, descriptions, keywords) provided by ATTA
- Contractor provides app binaries and release notes

**Environments:**
- Staging environment for testing (accessible to ATTA team)
- Production environment on AWS
- Separate configurations for each environment

**Branch Strategy:**
- Git flow methodology
- Feature branches for new development
- Staging branch for pre-production testing
- Production branch for live releases
- Pull request reviews by ATTA team before merging

### 10. Source Code & Repository Management

**Repository:**
- Platform: GitHub
- Access: ATTA team has full access to repository throughout development
- Ownership transfer: Upon project completion (Aug 2026)

**Code Review Process:**
- All code changes require pull request reviews
- ATTA team has right to review and request modifications
- Code must meet established quality standards and pass automated tests

**Documentation:**
- Comprehensive API documentation (GraphQL schema, REST endpoints)
- Architecture diagrams and system design documents
- Deployment and configuration guides
- Database schema documentation
- README files for each microservice

### 11. Other Terms
- During development, ATTA has the right to review the source code and make reasonable modification requests.
- Upon completion, the Contractor shall provide ATTA with all deliverables, including the application's front-end, back-end, and complete source code.
- ATTA shall have access to the project's GitHub repository and sandbox/staging environment to monitor progress, review code quality, and conduct preliminary testing.
- Repository ownership and transfer will be handled upon project completion.
- All code must be delivered with Docker and docker-compose configuration for deployment.