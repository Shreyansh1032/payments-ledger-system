# PayWave

A payments wallet built as a **double-entry ledger system**: every transfer is an atomic, idempotent, row-locked database transaction that writes a debit and a credit, so money is never created or lost.

**Live demo:** https://paywave-web.onrender.com

> The demo runs on free tiers. After a period of inactivity the first request can take up to a minute while the services wake up.

## Features

- **Payments:** send money by username, deposit, QR-code payments and scanning, transaction history with CSV export
- **Requests and splitting:** request money, split a bill across several people, approve or decline
- **Planning:** scheduled (recurring) transfers, savings goals, monthly budgets by category, spending insights
- **Security:** JWT access tokens with rotating refresh tokens, TOTP two-factor authentication (enforced on password and Google sign-in), Google Sign-In, email verification, per-IP rate limiting on login and signup
- **Product polish:** favorites, in-app notifications, dark mode, toast messages, glassmorphism UI

## Architecture

```mermaid
flowchart LR
    U[Browser]

    subgraph Live["Live deployment (Render + Neon)"]
        FE1[React frontend<br/>nginx]
        API1[Go API<br/>chi, JWT]
        DB1[(Neon Postgres)]
        FE1 -.static assets.-> U
        U -->|REST, CORS| API1
        API1 --> DB1
    end

    subgraph K8s["Kubernetes (minikube)"]
        ING[Ingress<br/>nginx]
        FE2[Frontend pods]
        API2[API pods<br/>HPA]
        DB2[(Postgres<br/>StatefulSet + PVC)]
        ING -->|/| FE2
        ING -->|/api| API2
        API2 --> DB2
    end

    G[Google Identity] -.verifies ID token.-> API1
    API1 -.exposes.-> M[/metrics<br/>Prometheus/]
```

The same Docker images run in both places. On Kubernetes an Ingress serves the frontend and API from one origin. On Render they are two services and the API allows the frontend origin through CORS.

## How a transfer works

```mermaid
sequenceDiagram
    participant C as Client
    participant A as API
    participant D as Postgres

    C->>A: POST /api/v1/transfer<br/>Idempotency-Key header
    A->>D: Look up idempotency key
    alt key already used
        D-->>A: existing transaction
        A-->>C: same result, no new money movement
    else new key
        A->>D: BEGIN
        A->>D: Lock both accounts (FOR UPDATE, sorted by id)
        A->>D: Check sender balance
        A->>D: Insert transaction
        A->>D: Insert debit and credit ledger entries
        A->>D: Update both balances
        A->>D: Insert recipient notification
        A->>D: COMMIT
        A-->>C: 200 completed
    end
```

Design points worth knowing:

- **Row locking in a fixed order.** Both accounts are locked in sorted ID order, so two opposite transfers cannot deadlock.
- **Idempotency.** `transactions.idempotency_key` is `UNIQUE`; retrying a request returns the original result instead of paying twice.
- **Exact money math.** All amounts are `NUMERIC(19,4)` in Postgres and `shopspring/decimal` in Go, never floats.
- **Double-entry invariant.** Each transfer writes one debit and one credit of the same amount. Tests assert the ledger stays balanced.
- **Treasury account.** Deposits move money from a `system_treasury` account, the only account allowed a negative balance (`CHECK (balance >= 0 OR is_system = true)`), so the ledger still balances.
- **New features add code, not edits.** Features such as deposits were added as separate services instead of changing the tested transfer logic.

## Data model (core tables)

```mermaid
erDiagram
    USERS ||--|| ACCOUNTS : owns
    ACCOUNTS ||--o{ TRANSACTIONS : "sends (from)"
    ACCOUNTS ||--o{ TRANSACTIONS : "receives (to)"
    TRANSACTIONS ||--|{ LEDGER_ENTRIES : "debit + credit"
    ACCOUNTS ||--o{ LEDGER_ENTRIES : posts
    USERS ||--o{ REFRESH_TOKENS : has

    USERS {
        uuid id PK
        string username UK
        string password_hash
        string email
        bool email_verified
        bool totp_enabled
        string google_id UK
    }
    ACCOUNTS {
        uuid id PK
        uuid user_id FK
        numeric balance
        char currency
        bool is_system
    }
    TRANSACTIONS {
        uuid id PK
        string idempotency_key UK
        uuid from_account_id FK
        uuid to_account_id FK
        numeric amount
        string status
        string category
    }
    LEDGER_ENTRIES {
        bigint id PK
        uuid transaction_id FK
        uuid account_id FK
        string entry_type
        numeric amount
    }
    REFRESH_TOKENS {
        uuid id PK
        uuid user_id FK
        string token_hash
    }
```

Feature tables not drawn here: `favorites`, `payment_requests`, `scheduled_transfers`, `savings_goals`, `budgets`, `email_verifications`, `notifications`. All 15 migrations live in [`migrations/`](migrations).

## CI/CD pipeline

```mermaid
flowchart LR
    P[git push or tag] --> GH[GitHub Actions]
    GH --> L[Lint<br/>go vet, gofmt]
    L --> T[Test<br/>Postgres service container]
    T --> B[Build<br/>multi-arch images<br/>amd64 + arm64]
    B --> GHCR[GHCR]
    B --> DH[Docker Hub]
    DH -->|pull| K[Kubernetes]
    DH -->|pull| R[Render]
```

Images are versioned from git tags (`v2.0.0`) and the version is compiled into the binary and exposed at `GET /version`.

## Tech stack

| Layer | Technology |
|---|---|
| Backend | Go, chi router, pgx, golang-migrate, zerolog, shopspring/decimal |
| Auth | JWT, bcrypt, refresh-token rotation, TOTP (pquerna/otp), Google ID-token verification |
| Database | PostgreSQL 16 (Neon in production) |
| Frontend | React 19, Vite, Tailwind CSS v4, React Router, Axios |
| Observability | Structured JSON logs, request IDs, Prometheus `/metrics` |
| Testing | Go integration tests against real Postgres, k6 load test |
| DevOps | Docker (multi-stage, multi-arch), GitHub Actions, Kubernetes (Ingress, HPA, StatefulSet, migration Job) |

## API overview

Base path `/api/v1`. All routes except auth require `Authorization: Bearer <token>`. Transfers also require an `Idempotency-Key` header.

| Area | Endpoints |
|---|---|
| Auth | `POST /auth/signup`, `/auth/login`, `/auth/refresh`, `/auth/logout`, `/auth/google`, `/auth/verify-email` |
| Account | `GET /account/balance`, `GET /account/profile`, `PUT /account/password`, `PUT /account/email`, `POST /account/deposit`, `POST /account/2fa/{setup,confirm,disable}` |
| Money | `POST /transfer`, `GET /transactions`, `GET /transactions/export`, `PATCH /transactions/{id}/category` |
| People | `GET /users/search`, `GET/POST /favorites`, `DELETE /favorites/{username}` |
| Requests | `POST /requests`, `POST /requests/split`, `GET /requests/{incoming,outgoing}`, `POST /requests/{id}/{approve,decline}` |
| Planning | `/scheduled-transfers`, `/goals` (+ `/{id}/contribute`), `/budgets`, `GET /insights/monthly` |
| Notifications | `GET /notifications`, `POST /notifications/read-all` |
| Ops | `GET /health`, `GET /version`, `GET /metrics` |

An OpenAPI draft is in [`docs/openapi.yaml`](docs/openapi.yaml).

## Run it locally

**Prerequisites:** Go, Node 20+, Docker, and the [`migrate` CLI](https://github.com/golang-migrate/migrate) built with the `postgres` tag.

```bash
# 1. Start Postgres and the API
docker compose up -d

# 2. Apply migrations
migrate -path migrations \
  -database "postgres://postgres:postgres@localhost:5432/paywave?sslmode=disable" up

# 3. Start the frontend
cd frontend
npm install
npm run dev
```

To run the API outside Docker, copy `.env.example` to `.env`, export it and run `go run ./cmd/api`.

### Configuration

| Variable | Used by | Purpose |
|---|---|---|
| `DATABASE_URL` | API | Postgres connection string |
| `JWT_SECRET` | API | Signing key for access tokens |
| `PORT` | API | Listen port (default 8080) |
| `CORS_ORIGIN` | API | Allowed frontend origin |
| `APP_BASE_URL` | API | Base URL used in email-verification links |
| `GOOGLE_CLIENT_ID` | API | Audience for Google ID-token verification |
| `VITE_API_URL` | Frontend (build time) | API base URL; empty for same-origin |
| `VITE_GOOGLE_CLIENT_ID` | Frontend (build time) | Google Sign-In client ID |

Email verification currently logs the verification link instead of sending a real email.

## Testing

```bash
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/paywave_test?sslmode=disable"
migrate -path migrations -database "$TEST_DATABASE_URL" up
go test ./...
```

The service tests run against a real Postgres database and cover transfers (including concurrent transfers and idempotent replays), deposits, ledger balance, auth, refresh-token rotation, 2FA, Google sign-in, requests, scheduled transfers, goals, budgets and the rate limiter. A k6 scenario for transfers is in [`loadtest/`](loadtest).

## Deployment

**Kubernetes** (manifests in [`k8s/`](k8s)): namespace, Postgres StatefulSet with a persistent volume, a migration Job, API Deployment with an HPA, frontend Deployment, and an nginx Ingress that routes `/` to the frontend and `/api` to the API.

```bash
kubectl apply -f k8s/
```

**Render + Neon** (the live demo): two web services run the published Docker images, and Neon provides Postgres. Migrations run against Neon with the `migrate` CLI before the API is deployed.

## Roadmap

- Grafana dashboard for the Prometheus metrics
- Automatic deploys to Render from CI
- Real email delivery for verification links
