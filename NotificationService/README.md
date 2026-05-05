# 📬 Notification Service — Backend Architecture Notes

> **Stack:** Express · TypeScript · BullMQ · Redis · Nodemailer · Handlebars · Winston

---

## 📦 Table of Contents

1. [Project Setup](#1-project-setup)
2. [How It Fits in the System](#2-how-it-fits-in-the-system)
3. [Folder Structure](#3-folder-structure)
4. [Redis Configuration](#4-redis-configuration)
5. [BullMQ Queue](#5-bullmq-queue)
6. [Email Producer](#6-email-producer)
7. [Email Processor (Worker)](#7-email-processor-worker)
8. [Nodemailer — Sending Emails via Gmail](#8-nodemailer--sending-emails-via-gmail)
9. [Handlebars Email Templating](#9-handlebars-email-templating)
10. [Winston Logger with Correlation ID](#10-winston-logger-with-correlation-id)
11. [Correlation ID Middleware](#11-correlation-id-middleware)
12. [Error Handling](#12-error-handling)
13. [Environment Variables](#13-environment-variables)
14. [Conceptual Q&A](#14-conceptual-qa)

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

# 4. Create a .env file in the root and define required variables (see Environment Variables section)

# 5. Start the dev server
npm run dev
```

---

## 2. How It Fits in the System

The **NotificationService** is a dedicated **consumer microservice**. It does not expose business APIs — its primary job is to **listen to a shared BullMQ queue** and send transactional emails.

```
BookingService (Producer)
        │
        │  mailerQueue.add("payload:mail", payload)
        ▼
  Redis Queue ("queue-mailer")
        │
        │  Worker picks up the job
        ▼
NotificationService (Consumer)
        │
        ├── renderMailTemplate(templateId, params)   → Handlebars → HTML
        │
        └── sendEmail(to, subject, html)             → Nodemailer → Gmail → Recipient
```

> Both **BookingService** and **NotificationService** must connect to the **same Redis instance** and use the **exact same queue name** (`"queue-mailer"`). A mismatch — even in casing — causes jobs to be silently dropped.

---

## 3. Folder Structure

```
src/
├── config/
│   ├── index.ts              # Loads .env, exports serverConfig
│   ├── logger.config.ts      # Winston logger with daily rotation
│   ├── mailer.config.ts      # Nodemailer transporter (Gmail)
│   └── redis.config.ts       # Redis singleton via ioredis
├── controllers/
│   └── ping.controller.ts    # Health check endpoint
├── dto/
│   └── notification.dto.ts   # NotificationDto shape
├── middlewares/
│   ├── correlation.middleware.ts  # Attaches x-correlation-id per request
│   └── error.middleware.ts        # AppError and generic error handlers
├── processors/
│   └── email.processor.ts    # BullMQ Worker — consumes queue jobs
├── producers/
│   └── email.producer.ts     # Adds email jobs to the queue
├── queues/
│   └── mailer.queue.ts       # BullMQ Queue instance
├── routers/
│   ├── v1/                   # v1 API routes
│   └── v2/                   # v2 API routes
├── services/
│   └── mailer.service.ts     # Calls Nodemailer transporter
├── templates/
│   ├── mailer/
│   │   └── welcome.hbs       # Handlebars email template
│   └── templates.handler.ts  # Reads .hbs file and compiles it
├── utils/
│   ├── errors/
│   │   └── app.error.ts      # Custom error classes
│   └── helpers/
│       └── request.helpers.ts # AsyncLocalStorage for correlation ID
└── server.ts                 # App entry point
```

---

## 4. Redis Configuration

Redis is used exclusively as the **message broker** for BullMQ. The connection is managed as a **singleton** to avoid creating multiple idle connections.

**File:** `config/redis.config.ts`

```ts
function connectToRedis() {
    let connection: Redis;

    const redisConfig = {
        port: serverConfig.REDIS_PORT,
        host: serverConfig.REDIS_HOST,
        maxRetriesPerRequest: null, // Required by BullMQ — disables automatic reconnection
    };

    return () => {
        if (!connection) {
            connection = new Redis(redisConfig);
            return connection;
        }
        return connection;
    };
}

export const getRedisConnObject = connectToRedis();
```

### Why `maxRetriesPerRequest: null`?

BullMQ requires this setting. Without it, ioredis will automatically retry failed commands for a limited number of times and then throw — which conflicts with how BullMQ manages its own retry logic internally.

---

## 5. BullMQ Queue

**File:** `queues/mailer.queue.ts`

```ts
export const MAILER_QUEUE = "queue-mailer";

export const mailerQueue = new Queue(MAILER_QUEUE, {
    connection: getRedisConnObject(),
});
```

The `Queue` instance is used by the **producer** to push jobs. The **Worker** on the consumer side connects to the same queue name to pull and process jobs.

> The queue name constant `MAILER_QUEUE` is exported and imported wherever it is needed — both in the producer and in the worker — to guarantee the name is always consistent and never hardcoded in two places.

---

## 6. Email Producer

**File:** `producers/email.producer.ts`

```ts
export const MAILER_PAYLOAD = "payload:mail";

export const addEmailToQueue = async (payload: NotificationDto) => {
    await mailerQueue.add(MAILER_PAYLOAD, payload);
    console.log(`Email added to queue: ${JSON.stringify(payload)}`);
};
```

### NotificationDto

**File:** `dto/notification.dto.ts`

```ts
export interface NotificationDto {
    to: string;           // Recipient email address
    subject: string;      // Email subject line
    templateId: string;   // Identifies which .hbs template to render
    params: Record<string, any>; // Dynamic values injected into the template
}
```

**Example call from BookingService:**

```ts
mailerQueue.add("payload:mail", {
    to: "user@example.com",
    subject: "Booking Confirmed",
    templateId: "welcome",
    params: { name: "John", appName: "Booking App" }
});
```

---

## 7. Email Processor (Worker)

**File:** `processors/email.processor.ts`

```ts
export const setupMailerWorker = () => {
    const emailProcessor = new Worker<NotificationDto>(
        MAILER_QUEUE,
        async (job: Job) => {
            if (job.name !== MAILER_PAYLOAD) {
                throw new Error("Invalid job name");
            }

            const payload = job.data;
            const emailContent = await renderMailTemplate(payload.templateId, payload.params);
            await sendEmail(payload.to, payload.subject, emailContent);

            logger.info(`Email sent to ${payload.to} with subject "${payload.subject}"`);
        },
        { connection: getRedisConnObject() }
    );

    emailProcessor.on("failed", () => console.error("Email processing failed"));
    emailProcessor.on("completed", () => console.log("Email processing completed successfully"));
};
```

### Processing Flow

| Step | Action |
|------|--------|
| 1 | Worker receives a job from `"queue-mailer"` |
| 2 | Validates job name is `"payload:mail"` |
| 3 | Calls `renderMailTemplate(templateId, params)` to compile the Handlebars template |
| 4 | Calls `sendEmail(to, subject, html)` to dispatch via Nodemailer |
| 5 | Logs success or failure via Winston |

> The worker is initialized at server startup inside `app.listen()`. This ensures the consumer is always running and connected before it can receive any jobs.

---

## 8. Nodemailer — Sending Emails via Gmail

**File:** `config/mailer.config.ts`

```ts
const transporter = nodemailer.createTransport({
    service: 'gmail',
    auth: {
        user: serverConfig.MAIL_USER,
        pass: serverConfig.MAIL_PASS
    }
});
```

**File:** `services/mailer.service.ts`

```ts
export async function sendEmail(to: string, subject: string, body: string) {
    try {
        await transporter.sendMail({
            from: serverConfig.MAIL_USER,
            to,
            subject,
            html: body
        });
        logger.info(`Email sent to ${to} with subject "${subject}"`);
    } catch (error) {
        throw new InternalServerError(`Failed to send email`);
    }
}
```

### Gmail App Password

When using Gmail as the SMTP provider, your regular account password will **not** work if 2-Factor Authentication is enabled. You need to generate an **App Password**:

1. Go to **Google Account → Security → 2-Step Verification → App passwords**.
2. Create a new app password for "Mail".
3. Use that generated password as `MAIL_PASS` in your `.env`.

> Never commit `MAIL_USER` or `MAIL_PASS` to version control. Always load them from `.env`.

---

## 9. Handlebars Email Templating

**File:** `templates/templates.handler.ts`

```ts
export async function renderMailTemplate(
    templateId: string,
    params: Record<string, any>
): Promise<string> {
    const templatePath = path.join(__dirname, 'mailer', `${templateId}.hbs`);
    const content = await fs.readFile(templatePath, 'utf-8');
    const finalTemplate = Handlebars.compile(content);
    return finalTemplate(params);
}
```

**Example template:** `templates/mailer/welcome.hbs`

```handlebars
Hi {{name}},

Welcome to {{appName}}!
We are excited to have you on board.

Thanks for joining us!
If you have any questions, feel free to reach out to our support team.
Best regards,
```

### How It Works

1. The `templateId` (e.g., `"welcome"`) maps directly to a file at `templates/mailer/welcome.hbs`.
2. The file is read from disk with `fs.readFile`.
3. `Handlebars.compile()` returns a function that accepts the `params` object.
4. Calling that function with `params` replaces all `{{placeholder}}` expressions with real values.
5. The resulting HTML string is passed directly to `sendEmail`.

### Adding a New Template

1. Create `templates/mailer/<templateId>.hbs`.
2. Use `{{variableName}}` placeholders where dynamic content should appear.
3. Pass the matching `templateId` and `params` object when calling `addEmailToQueue`.

---

## 10. Winston Logger with Correlation ID

**File:** `config/logger.config.ts`

```ts
const logger = winston.createLogger({
    format: winston.format.combine(
        winston.format.timestamp({ format: "MM-DD-YYYY HH:mm:ss" }),
        winston.format.json(),
        winston.format.printf(({ level, message, timestamp, ...data }) => {
            const output = {
                level,
                message,
                timestamp,
                correlationId: getCorrelationId(),
                data
            };
            return JSON.stringify(output);
        })
    ),
    transports: [
        new winston.transports.Console(),
        new DailyRotateFile({
            filename: "logs/%DATE%-app.log",
            datePattern: "YYYY-MM-DD",
            maxSize: "20m",
            maxFiles: "14d",
        })
    ]
});
```

### Key Features

| Feature | Details |
|---------|---------|
| **Structured JSON** | Every log line is a JSON object — easy to parse and pipe into log aggregators |
| **Timestamp** | Human-readable format: `MM-DD-YYYY HH:mm:ss` |
| **Correlation ID** | Each log entry carries the active request's correlation ID for traceability |
| **Daily Rotation** | Log files rotate daily, capped at 20 MB per file and retained for 14 days |
| **Dual output** | Logs are written to the console **and** to rotating files simultaneously |

---

## 11. Correlation ID Middleware

**File:** `middlewares/correlation.middleware.ts`

```ts
export const attachCorrelationIdMiddleware = (req: Request, res: Response, next: NextFunction) => {
    const correlationId = uuidV4();
    req.headers['x-correlation-id'] = correlationId;

    asyncLocalStorage.run({ correlationId }, () => {
        next();
    });
};
```

Every incoming request is automatically assigned a unique **UUID as a Correlation ID**. This ID is:

- Attached to the `x-correlation-id` request header.
- Stored in Node's **`AsyncLocalStorage`**, making it accessible from anywhere in the same async call stack — including the logger — without passing it through every function argument.

### Why AsyncLocalStorage?

In a traditional approach, you would thread the correlation ID through every function call: `sendEmail(to, subject, body, correlationId)`. `AsyncLocalStorage` removes this friction. Any code running within the same async context (controller → service → logger) can retrieve the correlation ID via `getCorrelationId()` without any parameter drilling.

---

## 12. Error Handling

**File:** `utils/errors/app.error.ts`

Custom error classes implement the `AppError` interface:

| Class | HTTP Status | Use Case |
|-------|-------------|----------|
| `InternalServerError` | 500 | Unexpected failures (e.g., email send failure) |
| `BadRequestError` | 400 | Invalid client input |
| `NotFoundError` | 404 | Requested resource not found |
| `UnauthorizedError` | 401 | Authentication required |
| `ForbiddenError` | 403 | Authenticated but not authorized |
| `ConflictError` | 409 | Request conflicts with current state |
| `NotImplementedError` | 501 | Feature not yet implemented |

**File:** `middlewares/error.middleware.ts`

Two error-handling middlewares are registered at the end of the Express middleware chain:

```ts
// Handles known AppError instances
export const appErrorHandler = (err: AppError, req, res, next) => {
    res.status(err.statusCode).json({ success: false, message: err.message });
};

// Catches everything else as a generic 500
export const genericErrorHandler = (err: Error, req, res, next) => {
    res.status(500).json({ success: false, message: "Internal Server Error" });
};
```

---

## 13. Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
PORT=3001

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Gmail SMTP
MAIL_USER=your-email@gmail.com
MAIL_PASS=your-gmail-app-password
```

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Port the Express server listens on | `3001` |
| `REDIS_HOST` | Hostname of the Redis server | `localhost` |
| `REDIS_PORT` | Port of the Redis server | `6379` |
| `MAIL_USER` | Gmail address used as the sender | — |
| `MAIL_PASS` | Gmail App Password (not your login password) | — |

---

## 14. Conceptual Q&A

---

**Q1. Why is the NotificationService a separate microservice and not part of the BookingService?**

Separating notification logic from booking logic follows the **Single Responsibility Principle** at the service level. The BookingService owns the business rules around creating and confirming bookings. Mixing email-sending logic into it would create coupling — a change to the email provider, template engine, or retry strategy would require touching the booking code. As a separate service, the NotificationService can be scaled, deployed, and updated independently.

---

**Q2. What happens if the NotificationService is down when BookingService pushes a job?**

Nothing is lost. BullMQ persists jobs in **Redis**, not in memory. When the NotificationService comes back online and the worker reconnects, it will pick up and process all queued jobs that accumulated during the downtime. This is the key advantage of using a queue over a direct HTTP call — the producer and consumer are **temporally decoupled**.

---

**Q3. What happens if the email fails to send (e.g., Gmail rejects it)?**

The `sendEmail` function throws an `InternalServerError`. BullMQ catches this exception from the worker's process function and marks the job as **failed**. By default, BullMQ will not retry failed jobs unless you configure `attempts` and `backoff` when adding the job to the queue:

```ts
mailerQueue.add(MAILER_PAYLOAD, payload, {
    attempts: 3,
    backoff: { type: 'exponential', delay: 2000 }
});
```

The `"failed"` event listener on the worker can also be used to log or alert on persistent failures.

---

**Q4. Why use Handlebars for email templates instead of plain string concatenation?**

String concatenation for HTML emails becomes unreadable and unmaintainable very quickly. Handlebars provides a clear separation between the email **structure** (the `.hbs` file) and the **dynamic data** (the `params` object). Adding a new variable to a template only requires updating the `.hbs` file — no code changes needed. It also eliminates the risk of accidentally breaking HTML structure while inserting dynamic content.

---

**Q5. Why is the Redis connection wrapped in a singleton factory?**

Creating a new `Redis` instance is expensive — it opens a TCP socket and performs a handshake. If `new Redis()` were called every time the queue or worker needed a connection, the service would open many redundant connections. The singleton pattern ensures only one connection is created for the entire lifetime of the process, which is reused everywhere `getRedisConnObject()` is called.

---

**Q6. What is the purpose of the `job.name` check inside the worker?**

```ts
if (job.name !== MAILER_PAYLOAD) {
    throw new Error("Invalid job name");
}
```

A single BullMQ queue can hold jobs of multiple named types. The `job.name` check acts as a **guard** — it ensures the worker only processes jobs it was designed to handle. If the queue is ever reused for a different type of job (e.g., a future SMS notification), an unrelated worker won't accidentally try to send emails for it.

---

**Q7. How does `AsyncLocalStorage` work for correlation IDs in the logger?**

`AsyncLocalStorage` is a Node.js built-in that provides a scoped key-value store tied to an async execution context. When the correlation middleware calls `asyncLocalStorage.run({ correlationId }, () => next())`, every callback, promise, and async function that executes downstream — including database calls and logger invocations — inherits that same store. The `getCorrelationId()` helper simply reads from it:

```ts
export const getCorrelationId = (): string | undefined => {
    const store = asyncLocalStorage.getStore();
    return store?.correlationId;
};
```

This means the Winston logger can embed the correlation ID in every log line with zero coupling to the request object.

---

*End of Notes*
