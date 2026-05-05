# ⭐ Review Service — Backend Architecture Notes

> **Stack:** Go · Chi Router · MySQL · Goose Migrations · go-playground/validator · robfig/cron

---

## 📦 Table of Contents

1. [Project Setup](#1-project-setup)
2. [How It Fits in the System](#2-how-it-fits-in-the-system)
3. [Folder Structure](#3-folder-structure)
4. [Database Setup and Migrations (Goose)](#4-database-setup-and-migrations-goose)
5. [Database Schema](#5-database-schema)
6. [Clean Architecture in Go](#6-clean-architecture-in-go)
7. [Request Lifecycle](#7-request-lifecycle)
8. [Validation Middleware](#8-validation-middleware)
9. [DTOs — Data Transfer Objects](#9-dtos--data-transfer-objects)
10. [Soft Deletion](#10-soft-deletion)
11. [JSON Response Utilities](#11-json-response-utilities)
12. [API Endpoints](#12-api-endpoints)
13. [Review Aggregation Cron Job](#13-review-aggregation-cron-job)
14. [MySQL Distributed Lock — GET_LOCK()](#14-mysql-distributed-lock--get_lock)
15. [Weighted Average Rating Formula](#15-weighted-average-rating-formula)
16. [HotelClient — Cross-Service Communication](#16-hotelclient--cross-service-communication)
17. [Environment Variables](#17-environment-variables)
18. [Makefile Commands](#18-makefile-commands)
19. [Conceptual Q&A](#19-conceptual-qa)

---

## 1. Project Setup

### Prerequisites

- Go 1.24.1+
- MySQL 8.0+
- [Goose](https://github.com/pressly/goose) migration tool installed globally:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Steps to Get Started

```bash
# 1. Clone the repository
git clone <repository-url>git clone https://github.com/m4nu-git/Express-Typescript-Starter-Project.git  ReviewService
cd ReviewService

# 2. Install dependencies
make deps

# 3. Create a .env file in the root directory (see Environment Variables section)

# 4. Run database migrations
make migrate-up

# 5. Start the service
make run
```

---

## 2. How It Fits in the System

The **ReviewService** serves two distinct responsibilities within the Airbnb microservices system:

**Part 1 — REST API:** Accepts and manages hotel reviews from users after a booking is completed.

**Part 2 — Background Cron Job:** Periodically aggregates new review ratings and pushes the updated average rating back to the HotelService.

```
Client (Postman / Browser)
        │
        │  POST /reviews, GET /reviews, etc.
        ▼
  ReviewService (REST API)
        │
   Saves review with is_synced=FALSE
        │
        ▼
   MySQL (airbnb_reviews DB)
        │
        │  (every 30s in dev / hourly in prod)
        ▼
  ReviewService (Cron Job)
        │
        ├── GET  /api/v1/hotels/id/:id   →  HotelService (fetch current rating)
        │
        └── PUT  /api/v1/hotels/updateById/:id  →  HotelService (push new avg)
```

> The ReviewService is the **only** writer to the `reviews` table. HotelService owns the hotel rating fields and is updated via HTTP — there is no shared database between services.

---

## 3. Folder Structure

```
ReviewService/
├── main.go                                   → Entry point — loads config, boots application
├── app/
│   └── application.go                        → Wires all dependencies, starts HTTP server + cron
├── config/
│   ├── db/
│   │   └── db.go                             → MySQL connection using go-sql-driver
│   └── env/
│       └── env.go                            → .env loader with typed helpers (GetString, GetInt, GetBool)
├── controllers/
│   ├── ping.go                               → Health check handler
│   └── review.go                             → HTTP handlers for all review endpoints
├── cronJob/
│   └── cron.go                               → Schedules ProcessPendingRatings via robfig/cron
├── db/
│   ├── migrations/
│   │   └── 20260216103732_create_review_table.sql  → Goose migration (up + down)
│   └── repositories/
│       ├── reviews.repository.go             → CRUD queries for the reviews table
│       └── review_aggregate_rating.repository.go   → Aggregate queries for the cron batch
├── dto/
│   └── review.go                             → CreateReviewRequestDTO, UpdateReviewRequestDTO, ReviewResponseDTO
├── middlewares/
│   └── validator.go                          → Request body decoding + validation middleware
├── models/
│   └── review.go                             → Review struct (maps to DB row)
├── router/
│   ├── router.go                             → Root Chi router setup
│   └── review_router.go                      → Registers all /reviews routes with their middlewares
├── services/
│   ├── review_service.go                     → CRUD business logic
│   └── review_aggregate_rating_service.go    → Batch aggregation + hotel rating update logic
├── client/
│   └── hotel_client.go                       → HTTP client to communicate with HotelService
└── utils/
    └── json.go                               → WriteJsonSuccessResponse / WriteJsonErrorResponse helpers
```

---

## 4. Database Setup and Migrations (Goose)

The project uses **Goose** for database migrations — a version-controlled approach to evolving the schema. Every migration file has an `UP` section (apply the change) and a `DOWN` section (revert it), identical to the philosophy used in Sequelize migrations.

### Installing Goose

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Migration Commands

All migration commands are wrapped in the `Makefile` for convenience:

```bash
make migrate-up          # Apply all pending migrations
make migrate-down        # Roll back the most recent migration
make migrate-reset       # Roll back all migrations and reset the database
make migrate-status      # Show the current migration state
make migrate-redo        # Undo and re-apply the last migration
```

### Creating a New Migration

```bash
make migrate-create name="add_column_to_reviews"
```

This generates a new timestamped `.sql` file under `db/migrations/`. Fill in the `-- +goose Up` and `-- +goose Down` blocks.

### Migration File Structure

Every migration file uses Goose comment directives:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE reviews ( ... );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE reviews;
-- +goose StatementEnd
```

---

## 5. Database Schema

**Database:** `airbnb_reviews`

```sql
CREATE TABLE reviews (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id    BIGINT NOT NULL,
    booking_id BIGINT NOT NULL,
    hotel_id   BIGINT NOT NULL,
    comment    TEXT NOT NULL,
    rating     INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    is_synced  BOOLEAN NOT NULL DEFAULT FALSE,

    INDEX idx_user_id    (user_id),
    INDEX idx_booking_id (booking_id),
    INDEX idx_hotel_id   (hotel_id),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at)
);
```

### Column Notes

| Column | Purpose |
|--------|---------|
| `rating` | Integer 1–5, enforced by a `CHECK` constraint at the database level |
| `deleted_at` | `NULL` means the review is active. A timestamp means it was soft-deleted |
| `is_synced` | `FALSE` (default) until the cron job has pushed this review's rating to HotelService |
| `created_at` | Used as the **cutoff timestamp** during cron aggregation — only reviews created before the cutoff are processed |

### Why These Indexes?

- `idx_user_id`, `idx_hotel_id`, `idx_booking_id` — speed up the filter query endpoints (e.g., "all reviews for hotel 42")
- `idx_deleted_at` — since all queries include `WHERE deleted_at IS NULL`, this index prevents a full table scan on every read
- `idx_created_at` — the cron's aggregate query filters `WHERE created_at <= ?`, so this index makes that scan efficient at scale

---

## 6. Clean Architecture in Go

The service uses a **layered architecture** with Go interfaces to invert dependencies and keep each layer independently testable.

```
HTTP Request
     │
 [ Chi Router + Middleware ]     → Route matching, validation, body parsing
     │
 [ Controller ]                  → Extracts validated payload, calls service, writes response
     │
 [ Service (interface) ]         → Business logic — validation, ID parsing, orchestration
     │
 [ Repository (interface) ]      → Raw SQL queries against MySQL
     │
 [ MySQL ]
```

### Dependency Injection via Constructor Functions

Dependencies flow downward via constructor functions in `application.go`:

```go
db, _  := dbConfig.SetupDB()          // MySQL connection
rr     := repo.NewReviewRepository(db) // Repository gets the DB
rs     := services.NewReviewService(rr) // Service gets the Repository
rc     := controllers.NewReviewController(rs) // Controller gets the Service
```

No global state. No `init()` functions. Every dependency is explicit and traceable.

### Why Interfaces?

Both `ReviewService` and `ReviewRepository` are defined as Go interfaces:

```go
type ReviewService interface {
    CreateReview(payload *dto.CreateReviewRequestDTO) (*models.Review, error)
    GetReviewById(id string) (*models.Review, error)
    // ...
}
```

The controller only depends on the **interface**, not the concrete struct. This means:
- The service layer can be swapped out without touching the controller.
- Mock implementations can be injected in tests.

---

## 7. Request Lifecycle

When `POST /reviews` is called:

```
1. Chi Router         → Matches route, runs ReviewCreateRequestValidator middleware
2. Validator MW       → Decodes JSON body → validates struct tags → stores payload in context
3. CreateReview()     → Controller reads payload from context, calls ReviewService
4. ReviewService      → Double-checks rating range (1–5), calls repository
5. ReviewRepository   → Executes INSERT, returns new Review model
6. Controller         → Calls WriteJsonSuccessResponse → 201 Created
```

The validated payload is passed to the controller via `context.WithValue`:

```go
ctx := context.WithValue(r.Context(), "payload", payload)
next.ServeHTTP(w, r.WithContext(ctx))
```

The controller retrieves it with a type assertion:

```go
payload := r.Context().Value("payload").(dto.CreateReviewRequestDTO)
```

---

## 8. Validation Middleware

**File:** `middlewares/validator.go`

Chi supports **per-route middleware** using `.With()`. Validation is applied only to routes that accept a body — `POST` and `PUT`:

```go
r.With(middlewares.ReviewCreateRequestValidator).Post("/reviews", rc.CreateReview)
r.With(middlewares.ReviewUpdateRequestValidator).Put("/reviews/{id}", rc.UpdateReview)
```

Validation is powered by **go-playground/validator** and uses struct tags:

```go
type CreateReviewRequestDTO struct {
    UserId    int64  `json:"user_id"    validate:"required"`
    BookingId int64  `json:"booking_id" validate:"required"`
    HotelId   int64  `json:"hotel_id"   validate:"required"`
    Comment   string `json:"comment"    validate:"required,min=1,max=1000"`
    Rating    int    `json:"rating"     validate:"required,min=1,max=5"`
}
```

If validation fails, the middleware responds with `400 Bad Request` and never calls `next` — the controller never runs.

---

## 9. DTOs — Data Transfer Objects

**File:** `dto/review.go`

DTOs define the exact shape of data at the API boundary, decoupled from the internal `Review` model.

| DTO | Used By | Fields |
|-----|---------|--------|
| `CreateReviewRequestDTO` | `POST /reviews` | `user_id`, `booking_id`, `hotel_id`, `comment`, `rating` |
| `UpdateReviewRequestDTO` | `PUT /reviews/{id}` | `comment`, `rating` only — user/hotel/booking cannot be changed |
| `ReviewResponseDTO` | All responses | Full review fields including `deleted_at` (omitted from JSON if nil) and `is_synced` |

`UpdateReviewRequestDTO` intentionally restricts what can be changed. A review's `user_id`, `hotel_id`, and `booking_id` are immutable once created — they cannot be patched.

---

## 10. Soft Deletion

Reviews are never hard-deleted. The `Delete` repository method sets `deleted_at` to the current timestamp:

```go
query := "UPDATE reviews SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL"
```

Every read query includes `WHERE deleted_at IS NULL` to automatically exclude soft-deleted records:

```go
query := "SELECT ... FROM reviews WHERE id = ? AND deleted_at IS NULL"
```

This means:
- A deleted review cannot be fetched, updated, or deleted again.
- The data remains in the database for auditing and recovery.
- The `idx_deleted_at` index ensures this filter is fast even on large tables.

---

## 11. JSON Response Utilities

**File:** `utils/json.go`

All handlers use two shared response helpers to maintain a consistent API response envelope:

```go
// Success → { "success": true, "message": "...", "data": { ... } }
func WriteJsonSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{})

// Error → { "success": false, "message": "...", "error": "..." }
func WriteJsonErrorResponse(w http.ResponseWriter, statusCode int, message string, err error)
```

**Example success response:**

```json
{
    "success": true,
    "message": "Review created successfully",
    "data": {
        "id": 1,
        "user_id": 10,
        "booking_id": 123,
        "hotel_id": 456,
        "comment": "Great stay!",
        "rating": 5,
        "created_at": "2026-02-16T10:37:32Z",
        "updated_at": "2026-02-16T10:37:32Z",
        "is_synced": false
    }
}
```

---

## 12. API Endpoints

Base URL: `http://localhost:8081`

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET`  | `/ping`  | Returns pong — confirms service is alive |

### CRUD Operations

| Method   | Endpoint          | Middleware                        | Description |
|----------|-------------------|-----------------------------------|-------------|
| `POST`   | `/reviews`        | `ReviewCreateRequestValidator`    | Create a new review |
| `GET`    | `/reviews`        | —                                 | Get all active reviews |
| `GET`    | `/reviews/{id}`   | —                                 | Get a single review by ID |
| `PUT`    | `/reviews/{id}`   | `ReviewUpdateRequestValidator`    | Update comment and rating |
| `DELETE` | `/reviews/{id}`   | —                                 | Soft-delete a review |

### Filter Operations

| Method | Endpoint           | Query Param         | Description |
|--------|--------------------|---------------------|-------------|
| `GET`  | `/reviews/user`    | `?user_id={id}`     | All reviews by a user |
| `GET`  | `/reviews/hotel`   | `?hotel_id={id}`    | All reviews for a hotel |
| `GET`  | `/reviews/booking` | `?booking_id={id}`  | All reviews for a booking |

### Example cURL Requests

**Create a review:**
```bash
curl -X POST http://localhost:8081/reviews \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "booking_id": 123,
    "hotel_id": 456,
    "comment": "Excellent service and clean rooms.",
    "rating": 5
  }'
```

**Get reviews by hotel:**
```bash
curl "http://localhost:8081/reviews/hotel?hotel_id=456"
```

**Soft-delete a review:**
```bash
curl -X DELETE http://localhost:8081/reviews/1
```

---

## 13. Review Aggregation Cron Job

The cron job is the second major responsibility of this service. It runs in the background alongside the HTTP server and is responsible for keeping hotel ratings up to date across services.

### How It Works — Step by Step

```
1. Cron fires (every 30s in dev, hourly in prod)
        │
2. Acquire MySQL advisory lock → GET_LOCK("process_pending_ratings", 1)
        │  If lock not acquired → another instance is running → skip
        │
3. Set cutoff = current UTC time
        │
4. Query:  SELECT hotel_id, SUM(rating), COUNT(*)
           FROM reviews
           WHERE is_synced = FALSE AND created_at <= cutoff
           GROUP BY hotel_id
        │
5. For each hotel with unsynced reviews:
        │
        ├── a. Begin MySQL transaction
        │
        ├── b. GET /api/v1/hotels/id/:hotelId  →  fetch { rating, ratingCount } from HotelService
        │
        ├── c. Calculate new average:
        │       newAvg = (oldAvg * oldCount + sumOfNewRatings) / newCount
        │
        ├── d. PUT /api/v1/hotels/updateById/:hotelId  →  push { rating: newAvg, ratingCount: newCount }
        │
        ├── e. UPDATE reviews SET is_synced = TRUE
        │       WHERE is_synced = FALSE AND hotel_id = ? AND created_at <= cutoff
        │
        └── f. COMMIT transaction  (rollback on any step failure)
        │
6. RELEASE_LOCK("process_pending_ratings")
```

### Why the `cutoff` Timestamp?

The cutoff is set to `time.Now().UTC()` at the **start** of each batch run. All SQL queries filter `created_at <= cutoff`. Any reviews that arrive during the batch run are not included in the current cycle — they will be picked up in the next scheduled run. This prevents a race condition where a review submitted mid-batch gets marked as synced before it was actually aggregated.

### Initial Run on Startup

The cron also triggers once at startup (after a 2-second delay) so that any reviews accumulated while the service was down are processed immediately, without waiting for the first scheduled tick:

```go
go func() {
    time.Sleep(2 * time.Second)
    log.Println("Initial run of ProcessPendingRatings...")
    svc.ProcessPendingRatings(context.Background())
}()
```

### Per-Hotel Transaction Isolation

Each hotel is processed in its own independent database transaction. If updating hotel #101 fails (e.g., HotelService is temporarily down), the transaction for hotel #101 is rolled back and the loop continues to hotel #102. A single failure does not block all other hotels from being processed.

---

## 14. MySQL Distributed Lock — GET_LOCK()

**File:** `services/review_aggregate_rating_service.go`

```go
var got sql.NullInt64
if err := s.db.QueryRowContext(ctx, "SELECT GET_LOCK(?, 1)", s.lockName).Scan(&got); err != nil {
    return fmt.Errorf("GET_LOCK error: %w", err)
}
if !got.Valid || got.Int64 != 1 {
    log.Println("ProcessPendingRatings: another instance is already running")
    return nil
}
defer s.db.ExecContext(context.Background(), "SELECT RELEASE_LOCK(?)", s.lockName)
```

`GET_LOCK(name, timeout)` is a **MySQL advisory lock** — it is not tied to a table or row, it is a named session-level lock stored in MySQL's internal lock manager.

| Return Value | Meaning |
|-------------|---------|
| `1` | Lock successfully acquired |
| `0` | Timed out waiting — another session holds the lock |
| `NULL` | An error occurred |

The second argument (`1`) is the timeout in seconds — if the lock is held by another session, this call waits up to 1 second before giving up.

`defer RELEASE_LOCK(...)` uses `context.Background()` deliberately, not the passed `ctx`. This ensures the lock is released even if the main context times out or is cancelled mid-batch — preventing a deadlock where the lock is held but no one can release it.

---

## 15. Weighted Average Rating Formula

When new reviews arrive for a hotel that already has an existing average rating, a simple recalculation of the mean is not enough — the existing data must be preserved.

The service uses a **weighted average**:

```
newAvg = (oldAvg × oldCount + sumOfNewRatings) / newCount
```

Where:
- `oldAvg` = current average rating stored in HotelService
- `oldCount` = number of reviews already factored into HotelService
- `sumOfNewRatings` = sum of all `is_synced = FALSE` ratings for this hotel
- `newCount` = `oldCount + count of new reviews`

**Example:**

| State | Value |
|-------|-------|
| Old average | 4.0 |
| Old count | 50 |
| New reviews | 3 ratings: 5, 4, 3 → sum = 12 |
| New count | 50 + 3 = 53 |
| New average | (4.0 × 50 + 12) / 53 = **4.038** |

This formula guarantees that a large volume of existing ratings is not wiped out by a small batch of new ones — each rating is weighted proportionally by the total count.

---

## 16. HotelClient — Cross-Service Communication

**File:** `client/hotel_client.go`

The `HotelClient` is a thin HTTP wrapper that abstracts all communication with the HotelService:

```go
type HotelClient struct {
    BaseURL    string
    HttpClient *http.Client
}
```

It exposes two methods:

| Method | HTTP Call | Description |
|--------|-----------|-------------|
| `GetHotelRating(hotelID)` | `GET /api/v1/hotels/id/:id` | Fetches current `rating` and `ratingCount` |
| `UpdateHotelRating(hotelID, rating, count)` | `PUT /api/v1/hotels/updateById/:id` | Pushes the new average rating and count |

The base URL is injected at construction time from the `HOTEL_SERVICE_URL` environment variable:

```go
hotelClient := client.NewHotelClient(config.GetString("HOTEL_SERVICE_URL", "http://localhost:3000/api/v1"))
```

This makes it easy to point to different HotelService instances across environments (local, staging, production) without code changes.

---

## 17. Environment Variables

Create a `.env` file in the root directory:

```env
# Server
PORT=:8081
APP_MODE=dev

# MySQL
DB_USER=root
DB_PASSWORD=root1234
DB_NET=tcp
DB_ADDR=127.0.0.1:3306
DBName=airbnb_reviews

# External Services
HOTEL_SERVICE_URL=http://localhost:3000/api/v1
```

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server listen address | `:5555` |
| `APP_MODE` | `dev` (cron every 30s) or `prod` (cron hourly) | `test` |
| `DB_USER` | MySQL username | `root` |
| `DB_PASSWORD` | MySQL password | `root1234` |
| `DB_NET` | Network type for MySQL driver | `tcp` |
| `DB_ADDR` | MySQL host and port | `127.0.0.1:3306` |
| `DBName` | MySQL database name | `review_dev` |
| `HOTEL_SERVICE_URL` | Base URL for HotelService API | `http://localhost:3000/api/v1` |

---

## 18. Makefile Commands

```bash
make run              # Run the service with go run main.go
make build            # Compile a binary: ./reviewservice
make deps             # Run go mod tidy && go mod download

make migrate-up       # Apply all pending migrations
make migrate-down     # Roll back the most recent migration
make migrate-reset    # Roll back all migrations
make migrate-status   # Show current migration state
make migrate-redo     # Undo and re-apply last migration
make migrate-create name="add_x_to_reviews"  # Generate a new migration file
make migrate-to version=20260216103732        # Migrate up to a specific version
make migrate-down-to version=20260216103732   # Roll back to a specific version
```

---

## 19. Conceptual Q&A

---

**Q1. Why does the ReviewService aggregate ratings in a background cron job instead of updating HotelService immediately when a review is created?**

Updating HotelService synchronously during `POST /reviews` would couple the success of the review creation to the availability of HotelService. If HotelService is slow or down at that moment, the review creation API would fail — even though the review itself was written correctly to the database. The cron approach decouples them: a review is always created instantly, and the rating propagation happens asynchronously. If HotelService is temporarily unavailable, the unsynced reviews accumulate and are processed on the next cron tick.

---

**Q2. What does `is_synced = FALSE` on a review mean, and what happens to reviews stuck in that state?**

`is_synced = FALSE` means the review's rating has been stored locally but has not yet been factored into the HotelService's average rating. On every cron run, all `is_synced = FALSE` reviews with `created_at <= cutoff` are aggregated and pushed to HotelService. Once successfully sent, they are marked `is_synced = TRUE` inside the same database transaction. If HotelService fails for an extended period, reviews accumulate in the `is_synced = FALSE` state and are all processed together in the next successful cron run — the weighted average formula handles any batch size correctly.

---

**Q3. Why is a MySQL advisory lock (`GET_LOCK`) used instead of a simple boolean flag in the database?**

A boolean flag in the database (e.g., `is_running = TRUE`) would require careful management: setting it before the batch, unsetting it after, and handling crashes where the flag is never unset. `GET_LOCK` is session-scoped — MySQL automatically releases the lock if the session disconnects or the process crashes, which eliminates the risk of a permanent lock due to an unhandled failure. It is also atomic: only one session can acquire a named lock at a time, making it safe in environments where the service is scaled horizontally.

---

**Q4. Why does `RELEASE_LOCK` use `context.Background()` instead of the function's own context?**

The function receives a `ctx` that may have a deadline or cancellation. If the batch runs long and the context times out mid-way, using the original `ctx` in the deferred `RELEASE_LOCK` call could fail because the context would already be done. Using `context.Background()` ensures the lock release is always attempted regardless of the caller's context state — a deadlock where no one can acquire the lock again is far more damaging than a slightly delayed release.

---

**Q5. What happens if the ReviewService crashes in the middle of a cron batch — after updating HotelService but before marking reviews as synced?**

Because the `MarkReviewsAsSynced` and the HotelService update are not wrapped in a distributed transaction (HotelService is a separate system), this is a real edge case. If the service crashes after pushing the rating to HotelService but before committing the `is_synced = TRUE` update, those reviews will be re-processed on the next cron run. This means HotelService could receive a double update for the same batch of reviews.

The impact is bounded by the weighted average formula — the recalculation will include reviews that were already counted once, slightly skewing the average. A full fix would require an idempotency mechanism on the HotelService side, or a two-phase commit pattern. For the current scale of this system, the rare double-count on a crash is an acceptable trade-off.

---

**Q6. Why does the validation middleware store the parsed payload in `context.WithValue` instead of calling the service directly?**

The middleware's only job is to parse and validate the incoming request. Calling the service from inside a middleware would mix two separate responsibilities into one place — making the code harder to test and violating the Single Responsibility Principle. Storing the validated payload in the request context is a clean handoff: the middleware guarantees the payload is valid and typed; the controller retrieves it and decides what to do with it. The controller's code is simpler too — it can type-assert directly without error handling, because the middleware already verified the shape.

---

**Q7. Why are both `ReviewService` and `ReviewRepository` defined as Go interfaces?**

Go's structural typing means any struct that implements all the methods of an interface satisfies it automatically — no `implements` keyword needed. Defining these as interfaces means the controller depends on an abstraction, not a concrete type. In practice, this means you can create a `MockReviewService` in tests that returns predictable data, inject it into the controller, and test the HTTP layer in complete isolation without ever touching the database.

---

*End of Notes*
