# 🏨 Hotel Service — Backend Architecture Notes

> **Stack:** Express · TypeScript · Sequelize · MySQL · BullMQ · Redis · Elasticsearch

---

## 📦 Table of Contents

1. [Project Setup](#1-project-setup)
2. [Configuring the ORM (Sequelize)](#2-configuring-the-orm-sequelize)
3. [Database Migrations](#3-database-migrations)
4. [Database Models](#4-database-models)
5. [End-to-End API Architecture](#5-end-to-end-api-architecture)
6. [Soft Deletion](#6-soft-deletion)
7. [Repository Refactoring](#7-repository-refactoring)
8. [Asynchronous Room Generation](#8-asynchronous-room-generation)
9. [CRON Jobs](#9-cron-jobs)
10. [Elasticsearch Integration](#10-elasticsearch-integration)
11. [Conceptual Q&A](#11-conceptual-qa)

---

## 1. Project Setup

### Steps to Get Started

```bash
# 1. Clone the starter template
git clone https://github.com/singhsanket143/Express-Typescript-Starter-Project.git <ProjectName>

# 2. Move into the project directory
cd <ProjectName>

# 3. Install dependencies
npm i

# 4. Create a .env file in the root and define the PORT variable

# 5. Start the dev server
npm run dev
```

---

## 2. Configuring the ORM (Sequelize)

**Sequelize** is an ORM (Object Relational Mapper) that lets you interact with a MySQL database using TypeScript/JavaScript classes instead of writing raw SQL.

### Installation

```bash
npm i sequelize          # Core ORM library
npm i mysql2             # MySQL driver used internally by Sequelize
npm install -D sequelize-cli  # CLI tools to generate boilerplate
```

### Initialize Sequelize

> ⚠️ Run this command from **inside the `src/`** folder, since that is where the project structure lives.

```bash
npx sequelize-cli init
```

This creates four folders:

| Folder       | Purpose                                                                 |
|--------------|-------------------------------------------------------------------------|
| `config`     | Holds database connection settings used by the CLI                     |
| `models`     | TypeScript representations of your database tables                     |
| `migrations` | Versioned scripts to create or modify database schema                  |
| `seeders`    | Dummy/seed data to help new developers understand the database quickly  |

---

### Understanding the Config Environments

The default `config.json` defines three environments:

| Environment   | When It's Used                                                                                  |
|---------------|-------------------------------------------------------------------------------------------------|
| `development` | While actively writing and testing new features locally                                         |
| `test`        | Running automated test suites — isolated from real data                                         |
| `production`  | The live application serving real users — only stable, verified code reaches here              |

> 💡 The `dialect` field tells Sequelize which database engine to use (e.g., `mysql`, `postgres`). It activates the correct internal driver based on this value.

---

### Why We Reorganize the Generated Files

By default, `config.json` stores database credentials in plain text — which is a **security risk** in any production-grade application.

**Our approach:**
- Delete all CLI-generated folders.
- Create a `config.ts` file that reads credentials from `.env` instead.
- Move `models/`, `seeders/`, and `migrations/` inside a `db/` folder for a cleaner structure.
- Create a `.sequelizerc` file at the project root to tell the Sequelize CLI where these reorganized folders are.

---

## 3. Database Migrations

A **migration** is a versioned script that modifies your database schema. Think of it as Git for your database — you can move forward to apply changes, or roll back to a previous state.

### Generating a Migration

```bash
npx sequelize-cli migration:generate --name create-hotel-table
```

### Migration Structure: UP and DOWN

| Part   | Purpose                                                                 |
|--------|-------------------------------------------------------------------------|
| `up`   | Applies the schema change (e.g., creates a table or adds a column)      |
| `down` | Reverts the schema change (e.g., drops the table or removes the column) |

**Example — UP (Create Table):**
```js
async up(queryInterface) {
  await queryInterface.sequelize.query(`
    CREATE TABLE IF NOT EXISTS hotels(
      id           INT AUTO_INCREMENT PRIMARY KEY,
      name         VARCHAR(255) NOT NULL,
      address      VARCHAR(255) NOT NULL,
      location     VARCHAR(255) NOT NULL,
      created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    );
  `);
}
```

**Example — DOWN (Drop Table):**
```js
async down(queryInterface) {
  await queryInterface.sequelize.query(`DROP TABLE IF EXISTS hotels;`);
}
```

### Running and Reverting Migrations

```bash
npx sequelize-cli db:migrate        # Apply pending migrations
npx sequelize-cli db:migrate:undo   # Roll back the most recent migration
```

Add these as npm scripts in `package.json` for convenience:
```json
"migrate": "sequelize-cli db:migrate",
"rollback": "sequelize-cli db:migrate:undo"
```

Then simply run `npm run migrate` or `npm run rollback`.

---

### TypeScript Support for Migrations

Since the project uses TypeScript, a `sequelize.config.js` file is created inside `config/`:

```js
require('ts-node/register');
const config = require('./db.config.js');
module.exports = config;
```

**How it works:**
When you run `npx sequelize-cli db:migrate`, Sequelize reads `.sequelizerc` → which points to `sequelize.config.js` → which calls `ts-node/register` to compile TypeScript on-the-fly → so Sequelize can execute it as JavaScript.

This way:
- You write migrations in **TypeScript**.
- Sequelize executes them as **JavaScript**.

---

### Benefits of Migrations

1. **Versioned schema evolution** — Every change is tracked, making it easy to see what changed and when.
2. **Safe experimentation** — Development and test environments absorb mistakes before they affect production.
3. **Rollback support** — If a new version of the schema causes issues, you can safely revert to a previous, stable state.

---

## 4. Database Models

A **model** is the TypeScript class that maps to a MySQL table. Since JavaScript doesn't understand SQL tables natively, models provide an object-oriented interface to interact with the database.

### Defining the Hotel Model

```ts
class Hotel extends Model<InferAttributes<Hotel>, InferCreationAttributes<Hotel>> {
  declare id: CreationOptional<number>;
  declare name: string;
  declare address: string;
  declare location: string;
  declare rating: number;
  declare ratingCount: number;
  declare createdAt: CreationOptional<Date>;
  declare updatedAt: CreationOptional<Date>;
}
```

### Mapping the Model to the MySQL Table

The `init()` function creates the connection between the TypeScript class and the actual `hotels` table in MySQL:

- **First parameter** — Maps each class property to the corresponding table column.
- **Second parameter** — Provides `tableName` and the Sequelize connection config (dialect, credentials, etc.).

---

### Common Sequelize Methods

| Category    | Method                  | Description                                              |
|-------------|-------------------------|----------------------------------------------------------|
| **Read**    | `findAll()`             | Returns all matching records                             |
|             | `findOne()`             | Returns the first matching record                        |
|             | `findByPk(id)`          | Finds a record by primary key                            |
|             | `findOrCreate()`        | Fetches a record or creates it if it doesn't exist       |
|             | `findAndCountAll()`     | Returns records along with the total count (pagination)  |
| **Create**  | `create(data)`          | Inserts a single new record                              |
|             | `bulkCreate(data[])`    | Inserts multiple records in one operation                |
|             | `build(data)`           | Creates an instance in memory without saving to DB       |
| **Update**  | `update(values, opts)`  | Updates records matching a condition                     |
|             | `save()`                | Persists changes made on an instance                     |
| **Delete**  | `destroy(opts)`         | Removes matching records from the table                  |
| **Utility** | `count()`               | Returns the number of matching records                   |
|             | `increment('field')`    | Increments a numeric column value                        |
|             | `aggregate()`           | Runs SQL aggregates like SUM, AVG, MAX, etc.             |

---

## 5. End-to-End API Architecture

The project uses a **bottom-up layered architecture**, where each layer has a single, well-defined responsibility.

```
Client (Postman / Browser)
        │
    [ Router ]          → Defines routes and wires middleware
        │
  [ Validation ]        → Validates incoming request body (Zod)
        │
  [ Controller ]        → Handles HTTP request/response cycle
        │
   [ Service ]          → Contains business logic
        │
  [ Repository ]        → Handles all direct DB interactions
        │
   [ Model / DB ]       → Sequelize Model ↔ MySQL Table
```

### Request Lifecycle — `createHotel`

When a `POST` request hits `http://localhost:3001/api/v1/hotels`:

1. **`server.ts`** — `express.json()` parses the JSON request body.
2. **Correlation ID Middleware** — Attaches a unique `x-correlation-id` header to every request for debugging and tracing.
3. **Router** — The `/api/v1` prefix routes to `v1Router`, which forwards `/hotels` to `hotel.router.ts`.
4. **Validation Layer** — `validateRequestBody(hotelSchema)` uses Zod to check required/optional fields. If valid, `next()` is called.
5. **Controller** — `createHotelHandler` extracts `req.body` and calls `createHotelService`.
6. **Service Layer** — Applies business logic (e.g., blacklisted address checks) and calls `createHotel` in the repository.
7. **Repository Layer** — Uses `Hotel.create()` to persist the record to MySQL.
8. **Response** — Travels back up through each layer and is sent to the client.

---

### DTOs (Data Transfer Objects)

A **DTO** defines the shape of data coming into the system. For example, `createHotelDto` defines what a valid hotel creation payload looks like. This enforces type safety at the boundary between the client and the application.

---

## 6. Soft Deletion

In production systems, records are rarely hard-deleted. Instead, a `deleted_at` timestamp is set on the row — a pattern called a **tombstone**. The record stays in the database for auditing and recovery, but the application treats it as gone.

> Sequelize has a built-in `paranoid: true` option for soft deletes, but it behaves inconsistently in complex real-world scenarios. We implement soft deletion manually for full control.

### Migration: Add `deleted_at` Column

```bash
npx sequelize-cli migration:generate --name add-deleted-at-to-hotels
```

```ts
async up(queryInterface: QueryInterface) {
  await queryInterface.addColumn('hotels', 'deleted_at', {
    type: DataTypes.DATE,
    allowNull: true,
    defaultValue: null,
  });
}

async down(queryInterface: QueryInterface) {
  await queryInterface.removeColumn('hotels', 'deleted_at');
}
```

After running the migration, update the Hotel model to include `deleted_at`, and implement `softDeleteHotel` across the repository, service, controller, and router layers.

---

## 7. Repository Refactoring

Looking at the hotel repository, methods like `getHotelById`, `updateHotelById`, `softDeleteById`, and `hardDeleteById` are not hotel-specific — they are **generic database operations** that any resource will need.

### Solution: Base Repository with Inheritance

A `base.repository.ts` file is created with all shared, reusable DB methods. Model-specific repositories (like `hotel.repository.ts`) extend this base class, inheriting common functionality and only adding resource-specific logic on top.

This approach follows:
- **DRY Principle** — Don't Repeat Yourself
- **Single Responsibility Principle (SRP)** — Each class does one thing well
- **Open/Closed Principle** — Base is open for extension, closed for modification

---

### CRUD Decisions

| Resource       | CRUD via REST API?                                                                         |
|----------------|-------------------------------------------------------------------------------------------|
| `RoomCategory` | Yes — owners need full CRUD (create, read, update, delete room categories)                |
| `Room`         | Partial — Room creation is handled by a CRON job; update and delete use REST APIs          |

---

## 8. Asynchronous Room Generation

Rooms are generated automatically for the next **90 days** using a job queue instead of doing it synchronously within the HTTP request. This prevents the API from blocking while waiting for potentially hundreds of room records to be created.

### Technology Stack

| Tool       | Role                                                       |
|------------|------------------------------------------------------------|
| **Redis**  | Acts as the message broker and job store                   |
| **ioredis**| Node.js client for connecting to Redis                     |
| **BullMQ** | Queue and worker management library built on top of Redis  |

### Flow

1. An API call triggers a room generation job.
2. The job is pushed to a **Redis queue** (via BullMQ producer).
3. A **worker/processor** picks up the job from the queue.
4. The worker creates room records for the next 90 days one by one.

This pattern decouples the HTTP request from the heavy lifting, improving response times and reliability.

---

## 9. CRON Jobs

To maintain a **rolling 90-day window** of available rooms, a CRON job runs daily and generates rooms for the following day.

### Tool: `node-cron`

A CRON expression is used to schedule this job to run automatically every day after 12:00 PM:

```
0 12 * * *   →   Run at noon, every day
```

The CRON job calls the same room generation logic used by the BullMQ worker, ensuring the 90-day availability window is always maintained without any manual intervention.

---

## 10. Elasticsearch Integration

### Why Elasticsearch?

A traditional SQL `LIKE '%name%'` query becomes very slow at scale and cannot handle typos. Elasticsearch solves both problems by providing **fast, fuzzy, full-text search** across millions of records.

For a system like Airbnb, users often search with incomplete or misspelled names (e.g., *"Chnai"* instead of *"Chennai"*). Elasticsearch handles this gracefully using **fuzzy matching** and **relevance scoring**.

---

### High-Level Flow

#### When a Hotel is Created:

```
createHotel API
      │
  Save to MySQL
      │
  Push job to Redis Queue (BullMQ)
      │
  Elasticsearch Worker picks up job
      │
  Fetches full hotel data from MySQL
      │
  Transforms data → ES document
      │
  Indexes document in ES index: "hotels"
```

#### When a User Searches:

```
GET /api/v1/hotel/search?name=Crowne%20Plaza
      │
  SearchController → SearchService
      │
  Build Elasticsearch query (multi_match + fuzziness + pagination)
      │
  Execute query against ES
      │
  Return ranked results to client
```

---

### Core Components

| Component                     | Role                                                                       |
|-------------------------------|----------------------------------------------------------------------------|
| `ElasticsearchRepository`     | Handles index, delete, and search operations against Elasticsearch         |
| `hotelIndexingProcessor`      | BullMQ worker that listens for indexing jobs and sends data to ES          |
| `SearchService`               | Builds and executes Elasticsearch queries                                  |
| `transformHotelToESDoc()`     | Converts a Sequelize model instance into an ES-compatible document format  |
| `hotelIndex.queue.ts`         | Defines the BullMQ queue for indexing jobs                                 |
| `hotelIndex.producer.ts`      | Pushes indexing/deletion jobs to the Redis queue                           |

---

### Key Elasticsearch Concepts

| Concept          | Description                                                                                |
|------------------|--------------------------------------------------------------------------------------------|
| **Index**        | Equivalent to a SQL table. We use an index called `hotels`.                                |
| **Document**     | Equivalent to a row. Each hotel is one document.                                           |
| **Field**        | Equivalent to a column (e.g., `name`, `address`, `location`).                             |
| **Analyzer**     | Breaks text into tokens to enable flexible and fast full-text search.                      |
| **Fuzziness**    | Allows typo-tolerant matching — e.g., "Chenai" correctly matches "Chennai".               |
| **Multi-Match**  | Searches across multiple fields simultaneously (e.g., `name` and `address` together).     |
| **`_score`**     | A relevance score assigned by ES to rank results by how closely they match the query.      |
| **Pagination**   | Controlled using `from` and `size` parameters to avoid loading too many results at once.   |

---

### Example Elasticsearch Query

```json
{
  "query": {
    "bool": {
      "must": [
        {
          "multi_match": {
            "query": "Chennai Marina",
            "fields": ["name^3", "address^2"],
            "fuzziness": "AUTO"
          }
        }
      ],
      "filter": [
        { "match": { "location": "Chennai" } }
      ]
    }
  },
  "from": 0,
  "size": 5,
  "sort": [{ "_score": { "order": "desc" } }]
}
```

> `name^3` means the `name` field is weighted **3x more** than other fields when calculating relevance.

---

### Environment Variables

| Variable             | Description                           | Example                    |
|----------------------|---------------------------------------|----------------------------|
| `ES_NODE`            | Elasticsearch server URL              | `https://localhost:9200`   |
| `ES_USERNAME`        | Elasticsearch username                | `elastic`                  |
| `ES_PASSWORD`        | Elasticsearch password                | `mypassword`               |
| `ES_HOTEL_INDEX`     | Name of the ES index for hotels       | `hotels`                   |
| `ES_BULK_CHUNK_SIZE` | Number of documents per bulk request  | `500`                      |

---

## 11. Conceptual Q&A

---

**Q1. What is an ORM and why do we use Sequelize in this project?**

An ORM (Object Relational Mapper) is a layer that translates between the object-oriented world of JavaScript/TypeScript and the relational world of SQL databases. Instead of writing raw SQL queries, you work with classes and method calls. Sequelize is used here because it supports MySQL, provides a rich API for querying, and integrates well with TypeScript through model definitions and type inference.

---

**Q2. Why are there three environments (development, test, production) in the Sequelize config?**

Each environment serves a different stage of the software lifecycle. Development is for writing and experimenting with code without risking real data. Test is for running automated tests in isolation. Production is the live system serving users. Keeping them separate ensures that bugs introduced during development or testing never affect real users or business-critical data.

---

**Q3. What is the purpose of a database migration and why not just modify the database directly?**

A migration is a version-controlled script that describes a database schema change — both how to apply it (`up`) and how to undo it (`down`). Modifying a database directly is risky because there is no record of what changed, no way to share the change with other developers, and no way to roll back if something goes wrong. Migrations solve all three problems by treating schema changes like code.

---

**Q4. What is the difference between the `models` folder and the `migrations` folder?**

The `models` folder contains TypeScript class definitions that represent the current shape of database tables. It is what the application uses at runtime to interact with the database. The `migrations` folder contains the historical scripts that built and evolved that structure over time. Models reflect the present state; migrations document the journey to get there.

---

**Q5. Why is soft deletion preferred over hard deletion in production systems?**

Hard deletion permanently removes a record from the database, making it impossible to recover if it was deleted by mistake. Soft deletion instead sets a `deleted_at` timestamp, keeping the record in the database but hiding it from normal queries. This enables data auditing, recovery if needed, and preserves historical integrity — all of which are important in any serious production application.

---

**Q6. Why is the repository layer separated from the service layer?**

The repository layer is responsible for only one thing: talking to the database. The service layer is responsible for business logic — rules, decisions, and transformations specific to the application. Keeping them separate means you can change the database or the ORM without touching business logic, and you can update business rules without touching database queries. This is the **Separation of Concerns** principle in practice.

---

**Q7. What is a DTO and why is it useful?**

A DTO (Data Transfer Object) is a TypeScript type or interface that defines the exact shape of data coming into or going out of a layer. For example, `createHotelDto` defines what fields are expected when creating a hotel. DTOs act as contracts between layers, catching shape mismatches at compile time rather than at runtime, and making the intent of each function immediately clear.

---

**Q8. Why is room generation done asynchronously using a queue instead of inside the HTTP request?**

If room generation for 90 days happened synchronously, the HTTP request would stay open until all records were created — potentially several seconds. This degrades the user experience and risks timeouts. By pushing the work to a BullMQ queue, the API responds immediately and the heavy lifting is done in the background by a worker. This is the foundation of non-blocking, scalable backend design.

---

**Q9. Why is Elasticsearch used for hotel search instead of a simple SQL LIKE query?**

SQL `LIKE` queries do a character-by-character pattern match. They are slow on large datasets and completely fail on typos. Elasticsearch tokenizes and indexes text fields at write time, enabling fast, flexible, relevance-ranked full-text search. It also supports fuzzy matching, multi-field queries, and scoring — capabilities that a SQL query cannot replicate efficiently.

---

**Q10. What does the `_score` field represent in an Elasticsearch response?**

The `_score` is a numerical relevance score that Elasticsearch assigns to each matching document based on how closely it matches the search query. Factors like term frequency, field weighting (e.g., `name^3` means name is three times more important than other fields), and fuzzy match closeness all influence the score. Results are typically sorted by `_score` in descending order so the most relevant hotels appear first.

---

**Q11. Why is hotel indexing in Elasticsearch done via a BullMQ job and not inline with the hotel creation API?**

Indexing a document in Elasticsearch is a network call to an external service. If it were done inline, a failure in Elasticsearch would cause the hotel creation API to fail too — even though the MySQL record was saved successfully. By using a background job, the concerns are decoupled. The hotel is created in MySQL, and if the ES indexing fails, the job can be retried automatically without affecting the user-facing response.

---

**Q12. What is the role of the `.sequelizerc` file?**

The `.sequelizerc` file is a configuration file read by the Sequelize CLI before executing any command. It allows you to override the default paths for the config file, models, migrations, and seeders folders. In this project, it is used to point the CLI to the reorganized folder structure inside `src/`, so commands like `npx sequelize-cli db:migrate` still work correctly.

---

*End of Notes*