# Airbnb Microservices – Platform Architecture Overview

This document provides an overview of the architecture and core functionality of our Airbnb-style microservices platform.

---

## 1. API Gateway (Golang)

**Role:** The API Gateway serves as the **single entry point** for all incoming client requests. It is responsible for **security enforcement, intelligent routing, and request lifecycle management**.

### Key Responsibilities

- **Centralized Authentication & Authorization**
  - Every inbound request must pass through the gateway before reaching any microservice.
  - Authentication is handled via **JWT tokens**.
  - **Role-Based Access Control (RBAC)** restricts route access based on user roles (`user` or `admin`), shielding internal services from unauthorized access.

- **Reverse Proxy**
  - The gateway transparently forwards requests to the appropriate downstream microservice, keeping internal service endpoints hidden from clients.
  - Services such as BookingService, HotelService, and ReviewService are all accessed through this gateway.

- **Rate Limiting**
  - To prevent abuse and promote fair usage, requests are capped at **5 per minute per IP address**.

- **Inter-Service Communication**
  
  The gateway supports two communication paradigms:

  **I. Synchronous (Request/Response)**
  - **REST over HTTP/HTTPS** — the approach used in this platform.
  - **gRPC** — uses HTTP/2 with Protocol Buffers (Protobuf) for high-performance communication.

  **II. Asynchronous (Message-Based)**
  - **Message Queue** — uses AMQP-based brokers (e.g., RabbitMQ, AWS SQS). This platform uses **io-redis and BullMQ**.
  - **Publish/Subscribe (Event-Driven)** — examples include Apache Kafka and Redis Pub/Sub.

  The gateway also handles **internal service-to-service calls**, authenticating them and aggregating data where required (e.g., merging user data from AuthService with ReviewService responses).

> **In essence:** The gateway functions as a **security checkpoint, traffic controller, and data aggregator** for the entire microservices ecosystem.

---

## 2. Booking Service (TypeScript)

**Role:** Manages the full hotel booking lifecycle and guarantees data consistency across the system.

### Core Logic

- Receives booking requests and validates **room availability** by making a synchronous REST API call to the HotelService.
- Employs **Redis + RedLock** for distributed locking, preventing race conditions where multiple users might attempt to book the same room simultaneously.
  ```
  await redlock.acquire([bookingResource], ttl);
  ```
  The lock is scoped to the `roomId` and held for up to 60 seconds before automatic release.
- New bookings are initially placed in a **pending state**.
- Upon booking creation, the room is marked with the `bookingId`, blocking other users from reserving it. This hold is valid for **10 minutes**. If the booking isn't confirmed within that window, a **cron job** (running every minute) detects expired pending bookings and:
  - Updates their status to `expired`.
  - Calls the HotelService's `/release` endpoint to free up the room for other users.
- Includes validation to ensure only the **original booking creator** can confirm the reservation.

### Booking Confirmation Flow

Confirmation uses **database transactions** to guarantee consistency:

1. Checks if the operation has already been processed using an **IdempotencyKey** (to prevent duplicate confirmations).
2. Updates the booking status to `CONFIRMED`.
3. Marks the IdempotencyKey as finalized.

### Key Database Tables

| Table | Purpose |
|---|---|
| `Booking` | Stores booking details — user, hotel, dates, guest count, and status. |
| `IdempotencyKey` | Prevents duplicate processing of booking or payment operations. |

---

## 3. Hotel Service (TypeScript)

**Role:** Handles all hotel and room management operations.

### Key Features

1. **Hotel CRUD** — Full create, read, update, and delete support for hotel records.

2. **Room Management**
   - Check room availability for a specified date range.
   - Associate a `bookingId` with a room upon reservation.
   - Release rooms after booking expiration to eliminate ghost bookings.

3. **Room Generation & Scheduling**
   - Supports bulk room creation via **Redis queues and background workers**.
   - Scheduled cron jobs extend room availability to maintain a rolling booking window (e.g., 90 days ahead).

4. **Elasticsearch Integration**
   
   Elasticsearch is a distributed search and analytics engine built on Apache Lucene.

   - **Hotel Indexing:** When a hotel is created, its ID is queued in Redis → a worker fetches the details → transforms the data → indexes it in Elasticsearch. Create, update, and delete operations are kept in sync with MySQL.
   - **Hotel Search:** The `/search` route dynamically builds Elasticsearch queries using filters such as `id`, `name`, `address`, and `location`, then calls `esClient.search` and returns filtered results.

---

## 4. Notification Service (TypeScript)

**Role:** Handles outbound email notifications to users.

### How It Works

- When an email needs to be sent, a job is added to a **Redis queue** containing the recipient address, email template name, and template parameters.
- Background **worker processes** consume the queue and dispatch emails using **Nodemailer** with **Handlebars templates** for dynamic content rendering.

---

## 5. Review Service (Async Rating Calculation)

**Role:** Manages user reviews and drives hotel rating calculations.

### Workflow

- A **cron job** periodically aggregates newly submitted reviews to compute updated hotel ratings.
- Combines review records with **user profile data from AuthService** to produce enriched review responses.
- Pushes the **updated average rating** back to the HotelService.

> **Note:** All internal API calls between microservices are authenticated via the API Gateway using JWT tokens.

---

## 6. Cross-Cutting Concepts

| Concept | Description |
|---|---|
| **Idempotency** | Ensures that repeated booking or payment operations don't result in duplicates. |
| **Database Transactions** | Groups related operations so they succeed or fail together, preserving consistency. |
| **Redis + RedLock** | Distributed locking mechanism to handle concurrent access to shared resources. |
| **Cron Jobs** | Scheduled background tasks for room availability extension, expired booking cleanup, and rating recalculation. |
| **DTOs & Repositories** | Enforces a clean separation between layers for maintainable, testable code. |

---

## 7. Tech Stack & Design Choices

| Layer | Technology | Reason |
|---|---|---|
| API Gateway | Golang | High concurrency, strong typing, performance |
| Microservices | Node.js + TypeScript | Expressive, typed business logic |
| ORM / Database | Sequelize, Prisma + SQL | Flexible data access patterns |
| Async Jobs | Redis + BullMQ | Reliable background processing and email delivery |
| Security | JWT + RBAC | Stateless auth with fine-grained access control |

---

## Summary

This platform is built to be **secure, scalable, and maintainable**. The **API Gateway** centralizes authentication, routing, and request management, while each microservice remains focused on a single domain responsibility — ensuring clean boundaries, independent deployability, and long-term system health.