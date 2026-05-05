# Airbnb Booking Service — Microservices with Prisma ORM and MySQL

This service manages the **booking-related functionalities** of an Airbnb-style system, built on top of **Prisma ORM** and **MySQL**. The primary goals are **data consistency**, **transactional safety**, and **concurrency control** — particularly to eliminate issues like **double bookings** and **race conditions**.

---

## 🚀 Project Setup

### Clone and Install

```bash
git clone <ProjectName>
cd <ProjectName>
npm i
```

Create a `.env` file in the root directory and add the required environment variables (e.g., `PORT`).

```bash
npm run dev
```

---

## 🧠 Core Concepts

- **Microservices Architecture** — Each service is independently deployable and owns its own database.
- **Concurrency Handling** — Prevents double bookings through distributed locking.
- **Separate Databases**:
  - `airbnb_booking_dev` — handles booking data.
  - `airbnb_dev_mode` — handles hotel and room data.
  - These databases are **kept isolated**; cross-database JOINs are not permitted.

> If you delete and recreate your database, run `npx prisma db push` from within the `prisma/` directory to sync all tables. Running it from outside that directory will result in a `prisma.schema file not found` error.

---

## 🛠️ Setting Up Prisma

### Installation Steps

```bash
# Step 1: Install Prisma CLI
npm install prisma

# Step 2: Initialize Prisma inside the src folder
cd src
npx prisma init

# Step 3: Clean up extra .env and .gitignore files created by Prisma

# Step 4: Install Prisma Client
npm install @prisma/client
```

Update your `.env` with the MySQL connection string:

```env
DATABASE_URL="mysql://<user>:<password>@localhost:3306/airbnb_booking_dev"
```

> `@prisma/client` is the runtime library that interacts with your database based on the Prisma schema definition.

---

## 🧩 Prisma Schema

**File:** `prisma/schema.prisma`

```prisma
model Booking {
  id             Int           @id @default(autoincrement())
  userId         Int
  hotelId        Int
  createdAt      DateTime      @default(now())
  updatedAt      DateTime      @updatedAt
  bookingAmount  Int
  status         BookingStatus
  totalGuests    Int
}

enum BookingStatus {
  PENDING
  CONFIRMED
  CANCELLED
}
```

> Prisma now supports **modular schema files**, so large schemas can be split across multiple files for better maintainability.

---

## 🔄 Running Migrations

```bash
# Initial migration
npx prisma migrate dev --name init

# After adding a new field (e.g., totalGuest)
npx prisma migrate dev --name added_totalGuest_column_to_Booking_table
```

Every migration generates a new folder under `prisma/migrations/`, creating a clean audit trail of schema changes.

---

## 🔄 Sequelize vs Prisma — A Quick Comparison

In **Sequelize**, you extend the `Model` class and manually define field types and constraints:

```ts
class Hotel extends Model<InferAttributes<Hotel>, InferCreationAttributes<Hotel>> {
  declare id: CreationOptional<number>;
  declare name: string;
}
Hotel.init({ id: { type: 'integer', autoIncrement: true, primaryKey: true } });
```

In **Prisma**, you skip all of that. Just define your schema in `schema.prisma` and run:

```bash
npx prisma generate
```

Prisma reads your `.env` and `schema.prisma`, then auto-generates fully typed client code in `node_modules`. Types like `Booking` and `BookingStatus` become immediately available throughout your project.

---

## 📁 Folder Structure

```
src/
├── prisma/
│   └── client.ts          # PrismaClient singleton
├── repositories/
│   └── booking.ts         # DB query functions
├── services/
│   └── booking.service.ts # Business logic
```

### Prisma Client Singleton (`prisma/client.ts`)

```ts
import { PrismaClient } from "@prisma/client";
export default new PrismaClient();
```

---

## 🔑 Idempotency Key — Preventing Duplicate Bookings

### What Is an Idempotency Key?

An idempotency key is a **unique identifier (UUID)** generated at the start of a booking flow. It ensures that even if a user sends the same request multiple times (e.g., due to network issues), only **one booking is created**.

### Installing UUID

```bash
npm install uuid
```

Generate the key inside the service layer when creating a pending booking.

### The `finalized` Property

The `IdempotencyKey` model includes a `finalized` boolean (default: `false`). When `finalized` is `true`, it signals that the booking completed its full lifecycle successfully. If the same idempotency key is received again, no new booking is created — the existing one is returned.

---

## 🔁 Two Approaches to Creating a Booking + Idempotency Key

### Approach 1 — Sequential (Slower)

1. Create the `Booking` (pending state).
2. Then create the `IdempotencyKey` referencing that booking.

**Downside:** Both DB calls happen back-to-back, and the key is only available after the booking is created.

### Approach 2 — Parallel (Faster) ✅ Preferred

1. Create the `Booking` (pending state).
2. **Simultaneously** generate and store the `IdempotencyKey` asynchronously.
3. In a later request, confirm the booking and finalize the key.

**Advantage:** The user gets a faster response. The asynchronous update can be retried (2–3 attempts) if it fails.

---

## 🔗 Relationship Between Booking and IdempotencyKey

To link these two models, a **one-to-one relationship** is established from the `IdempotencyKey` side (keeping the `Booking` model clean).

After updating `schema.prisma`, run:

```bash
npx prisma migrate dev --name Added_relationship_for_booking_to_idmpotency_key
```

When creating a booking with `BookingCreateInput`, you can now optionally attach a nested `IdempotencyKey`:

```ts
{
  userId: number;
  hotelId: number;
  bookingAmount: number;
  totalGuests: number;
  idempotencyKey?: Prisma.IdempotencyKeyCreateNestedOneWithoutBookingInput;
}
```

---

## ✅ Solving the Concurrency Problem in `confirmBookingService`

### The Problem

```ts
export async function confirmBookingService(idempotencyKey: string) {
  const idempotencyKeyData = await getIdemPotencyKey(idempotencyKey);
  if (!idempotencyKeyData) throw new NotFoundError("Idempotency key not found");
  if (idempotencyKeyData.finalized) throw new BadRequestError("Booking already finalized");

  const booking = await confirmBooking(idempotencyKeyData.bookingId);
  await finalizeIdempotencyKey(idempotencyKey);
  return booking;
}
```

If a user sends **two rapid concurrent requests** (e.g., double-tap due to slow network):

- Request A reads the key → not finalized.
- Context switches to Request B → also reads the key → not finalized.
- B confirms and finalizes the booking.
- Context switches back to A → also tries to finalize → **duplicate operation**.

### The Solution

Two mechanisms combined:

1. **Wrap everything in a Prisma `$transaction`** — if any step fails, the entire operation rolls back atomically.
2. **Apply a pessimistic lock (`SELECT ... FOR UPDATE`)** on the idempotency key row — only one request proceeds; others are blocked until the first transaction completes.

```ts
export async function confirmBookingService(idempotencyKey: string) {
  return await PrismaClient.$transaction(async (tx) => {
    const idempotencyKeyData = await getIdemPotencyKeyWithLock(tx, idempotencyKey);
    if (!idempotencyKeyData) throw new NotFoundError("Idempotency key not found");
    if (idempotencyKeyData.finalized) throw new BadRequestError("Booking already finalized");

    const booking = await confirmBooking(tx, idempotencyKeyData.bookingId);
    await finalizeIdempotencyKey(tx, idempotencyKey);
    return booking;
  });
}
```

### Implementing the Lock

```ts
export async function getIdemPotencyKeyWithLock(tx: Prisma.TransactionClient, key: string) {
  if (!isValidUUID(key)) throw new BadRequestError("Invalid idempotency key format");

  const idempotencyKey: Array<IdempotencyKey> = await tx.$queryRaw(
    Prisma.sql`SELECT * FROM IdempotencyKey WHERE idemKey = ${key} FOR UPDATE`
  );

  if (!idempotencyKey || idempotencyKey.length === 0) {
    throw new BadRequestError("Idempotency key not found");
  }

  return idempotencyKey[0];
}
```

> **Note:** Use `Prisma.sql` with tagged template literals (parameterized queries) instead of `Prisma.raw` with string interpolation. This prevents SQL injection vulnerabilities.

---

## 🔒 Solving Double Booking Across Users with Redlock

### The Problem

Even with idempotency keys, two **different users** could concurrently try to book the **same hotel room**. Without coordination, both might see the room as available and both succeed — a classic **race condition**.

### The Solution: Distributed Locking with Redis + Redlock

```bash
npm install ioredis redlock
```

### Redis Configuration (`config/redis.config.ts`)

```ts
import IORedis from 'ioredis';
import Redlock from 'redlock';
import { serverConfig } from '.';

export const redisClient = new IORedis(serverConfig.REDIS_SERVER_URL);

export const redlock = new Redlock([redisClient], {
  driftFactor: 0.01,  // Accounts for Redis clock drift
  retryCount: 10,     // Retry up to 10 times
  retryDelay: 200,    // Wait 200ms between retries
  retryJitter: 200,   // Adds randomness to avoid thundering herd
});
```

### Booking Service with Locking (`booking.service.ts`)

```ts
export async function createBookingService(createBookingDTO: CreateBookingDTO) {
  const ttl = serverConfig.LOCK_TTL;
  const bookingResource = `booking:${createBookingDTO.hotelId}`;

  const lock = await redlock.acquire([bookingResource], ttl);

  try {
    const booking = await createBooking({
      userId: createBookingDTO.userId,
      hotelId: createBookingDTO.hotelId,
      bookingAmount: createBookingDTO.bookingAmount,
      totalGuests: createBookingDTO.totalGuests,
    });

    const idempotencyKey = generateIdempotencyKey();
    await createIdempotencyKey(idempotencyKey, booking.id);

    return { bookingId: booking.id, idempotencyKey };

  } finally {
    await lock.release();  // Always release the lock
  }
}
```

---

## 🧠 How Redlock Works Internally

### Step-by-Step

| Step | Action | Why It Matters |
|------|--------|----------------|
| 1 | Generate a unique `lock_id` | Ensures only the lock holder can release it |
| 2 | `SET resource lock_id NX PX ttl` on Redis | Atomic — either succeeds or fails, no partial state |
| 3 | Require majority consensus across Redis nodes | Guarantees distributed safety |
| 4 | Check elapsed time vs. TTL | Rejects lock if clock drift consumed too much time |
| 5 | Release with ID check before `DEL` | Prevents accidental unlock by another process |
| 6 | TTL auto-expires stale locks | Prevents deadlocks if the service crashes |

### Concurrency Scenario

| Step | User A | User B |
|------|--------|--------|
| 1 | Calls `redlock.acquire(["booking:101"], ttl)` | Calls `redlock.acquire(["booking:101"], ttl)` |
| 2 | Key is free → **lock acquired** ✅ | Key is taken → **fails or retries** ❌ |
| 3 | Creates booking safely | Waits or receives an error |
| 4 | Releases lock | Can now attempt acquisition |

### Best Practices

- Keep TTL short — just long enough to complete the critical section.
- Always use `try...finally` to guarantee the lock is released.
- Use multiple Redis instances (3–5) for true distributed reliability.
- Handle `LockError` explicitly to provide meaningful feedback to the client.

---

## 📬 Notification Service via BullMQ

### Architecture

The **BookingService** and **NotificationService** communicate asynchronously through a **Redis-backed BullMQ queue**:

- **BookingService (Producer):** Pushes email jobs into the `"email-producer"` queue.
- **NotificationService (Consumer/Worker):** Listens to the same queue and processes jobs.

```ts
// Producer (BookingService)
mailerQueue.add("email-producer", payload);

// Consumer (NotificationService)
const worker = new Worker("email-producer", async (job) => {
  // process the job
});
```

> Both services must connect to the **same Redis instance** and use the **exact same queue name**. Even a minor mismatch in queue names will cause jobs to be silently ignored.

### Folder Structure

```
src/
├── queues/      # mailerQueue BullMQ instance
├── producer/    # Adds jobs to the queue
├── dto/         # NotificationDTO shape definition
```

---

## 🏨 Room Availability Check (Cross-Service Communication)

Before creating a booking, the **BookingService** calls the **HotelService** API to verify room availability:

**Logic:** A room is considered available if, for the given `roomCategoryId` and date range (`checkInDate` → `checkOutDate`), no `bookingId` exists in the rooms table.

The **HotelService** exposes a dedicated API endpoint for this check. The **BookingService** calls it before proceeding with lock acquisition and booking creation.

---

## 📌 Summary

| Concept | Tool/Approach |
|---------|--------------|
| ORM | Prisma |
| Database | MySQL |
| Duplicate request prevention | Idempotency Key (UUID) |
| Transaction safety | Prisma `$transaction` |
| Row-level locking | `SELECT ... FOR UPDATE` |
| Distributed concurrency | Redlock + Redis |
| Async job queue | BullMQ + Redis |
| Cross-service communication | HTTP API calls |

---

## ❓ Scenario-Based Q&A

### Q1. What happens if a user accidentally taps "Confirm Booking" twice very rapidly, and both requests reach the server within milliseconds of each other?

**Answer:**
Without protection, both requests would read the idempotency key, find it not finalized, and both attempt to confirm the booking — resulting in a duplicate operation.

With the solution in place:
- Both requests enter `confirmBookingService` and try to begin a Prisma `$transaction`.
- Inside the transaction, `getIdemPotencyKeyWithLock` runs `SELECT ... FOR UPDATE` on the idempotency key row.
- Only **one request acquires the row lock**. The other is **blocked at the database level** until the first transaction commits or rolls back.
- Once the first request finalizes the key (`finalized = true`) and commits, the second request is unblocked, reads the key, sees `finalized = true`, and throws a `BadRequestError("Booking already finalized")`.
- Result: **exactly one booking is confirmed** — no duplicates.

---

### Q2. What happens if the Redis server goes down while a Redlock lock is being held during a booking creation?

**Answer:**
This depends on the TTL configuration:

- If Redis crashes while User A holds the lock, the lock entry is lost from memory (since Redis stores it in RAM by default).
- When Redis restarts, the key no longer exists — meaning **any process can acquire the lock again immediately**, even if User A's booking creation is still in progress.
- This creates a window where two processes could simultaneously hold the lock — a **split-brain scenario**.

Mitigations:
- Use **multiple Redis instances** (Redlock requires a majority of nodes to agree). Even if one node fails, the lock remains safe if the majority still hold it.
- Enable **Redis persistence** (`AOF` or `RDB`) so keys survive restarts.
- Keep TTL short so stale locks expire quickly after recovery.
- Implement retry logic with backoff on `LockError` in the application layer.

---

### Q3. What happens if the BookingService crashes after the Redlock is acquired but before the lock is released?

**Answer:**
This is exactly the failure mode that **TTL** is designed to handle.

- The lock is stored in Redis with an expiry (`PX ttl`).
- If the BookingService crashes mid-operation, it cannot call `lock.release()`.
- However, Redis will **automatically delete the key** once the TTL expires.
- After expiry, other processes can acquire the same lock and proceed.

This is why TTL should be set to slightly longer than the expected maximum duration of the critical section — not too short (to avoid premature expiry) and not too long (to avoid extended blocking on crash).

---

### Q4. What happens if the BookingService creates a booking successfully but the async call to create the IdempotencyKey fails?

**Answer:**
Under **Approach 2** (parallel/async creation), this scenario is a real risk:

- The booking exists in the database in `PENDING` state.
- But no idempotency key is linked to it.
- When the user tries to confirm the booking, `getIdemPotencyKey()` returns nothing, and the service throws `NotFoundError("Idempotency key not found")`.
- The booking is now "orphaned" — it exists but can never be confirmed through the normal flow.

Recommended safeguards:
- Implement a **retry mechanism** (2–3 attempts with exponential backoff) for the async idempotency key creation.
- Add a **background cleanup job** (e.g., via BullMQ) that periodically detects and either relinks or cancels orphaned `PENDING` bookings older than a threshold.
- Alternatively, fall back to **Approach 1** (sequential creation) wrapped in a single `$transaction` for stricter atomicity, accepting the slight performance tradeoff.

---

### Q5. What happens if two users try to book different room categories in the same hotel at the exact same time?

**Answer:**
With the current locking strategy (`booking:${hotelId}`), **both users are competing for the same lock**, even though they want different room categories.

- User A books room category `DELUXE` → acquires lock `booking:101`.
- User B books room category `SUITE` → also tries to acquire lock `booking:101` → **blocked unnecessarily**.

This creates a performance bottleneck — the system serializes bookings at the hotel level when it only needs to serialize them at the room category level.

**Better approach:** Change the lock key to include the room category:

```ts
const bookingResource = `booking:${createBookingDTO.hotelId}:${createBookingDTO.roomCategoryId}`;
```

This allows concurrent bookings for different room categories in the same hotel, significantly improving throughput under load while still preventing double-booking within the same category.