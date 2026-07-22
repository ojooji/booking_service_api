# Booking Service API

REST API for appointment scheduling — think barbershop or clinic booking. Users register, browse services and employees, check free time slots, and book appointments. Built to practice production-style Go: layered architecture, real concurrency handling, and tests.

## Stack

Go 1.25 · chi · PostgreSQL (pgx) · JWT · Swagger · testify + testcontainers

## What's interesting in here

- **Double-booking prevention under concurrency** — a `pg_advisory_xact_lock` per employee serializes the conflict-check-then-insert, and a `btree_gist` exclusion constraint (`EXCLUDE USING gist (employee_id WITH =, tstzrange(start_time, end_time) WITH &&)`) makes the database itself reject overlapping bookings even if application code is bypassed. See `internal/repository/pgx/booking_repo.go` and `migrations/005`.
- **Slot availability** — free 30-minute slots are computed per employee per day from existing bookings, within business hours.
- **Layered architecture** — handlers (HTTP) → services (business rules) → repositories (SQL), with repository interfaces so services are unit-testable with mocks.

## Getting started

```bash
cp .env.example .env          # set JWT_SECRET to something real

docker compose up -d          # local PostgreSQL

make migrate-up               # needs golang-migrate CLI
make run
```

Swagger UI: http://localhost:8080/swagger/

## API overview

```
POST   /api/v1/auth/register        POST   /api/v1/auth/login

GET    /api/v1/services             GET    /api/v1/employees
GET    /api/v1/employees/{id}/slots?date=YYYY-MM-DD

POST   /api/v1/bookings             (auth)
GET    /api/v1/bookings             (auth, own bookings)
GET    /api/v1/bookings/{id}        (auth, owner or admin)
PUT    /api/v1/bookings/{id}/cancel (auth, owner or admin)

POST/PUT/DELETE on /services, /employees, /users, GET /bookings/all  (admin)
```

Responses use a consistent envelope: `{"status": "success", "data": ...}` / `{"status": "error", "error": {"code", "message"}}`.

## Testing

```bash
go test ./tests/unit/...        # fast, mocked repositories
go test ./tests/integration/... # spins up Postgres via testcontainers (needs Docker)
```

## Known limitations (honest list)

- Single JWT per login — no refresh flow, and logout is client-side only (tokens are valid until expiry, there is no revocation)
- Business hours (09:00–18:00) are validated in the client's timezone offset, not the employee's — fine for a single-timezone deployment, wrong for a global one
- No rate limiting on auth endpoints
- No pagination on list endpoints
- No admin bootstrap — the first admin has to be promoted via SQL
- Error responses on 500s echo internal error text; should be replaced with generic messages + server-side logging

These are documented on purpose — I'd rather show I know where the edges are.
