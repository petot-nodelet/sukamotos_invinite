# SUKAMOTO Wedding Invitation — Backend

Go backend implementing **API Contract v1.0**.

## Stack

- Go 1.22+
- SQLite (`modernc.org/sqlite` — pure Go, no CGO)
- chi router
- bcrypt password hashing
- Cookie-based admin sessions (HttpOnly)

## Quick start

```bash
cd backend
go mod tidy
go run ./cmd/server
```

Server listens on `:8080` by default.

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ADDR` | `:8080` | Listen address |
| `DATABASE_PATH` | `./data/sukamoto.db` | SQLite file |
| `PROOF_STORAGE_PATH` | `./storage/proofs` | Private payment proof storage |
| `ADMIN_USERNAME` | `admin` | Bootstrap admin username |
| `ADMIN_PASSWORD` | `changeme` | Bootstrap admin password (hashed on first run) |
| `PAYMENT_METHOD` | `bank_transfer` | Default payment method label |
| `PAYMENT_DESC` | (see config) | Payment instructions text |
| `MAX_UPLOAD_MB` | `5` | Max payment proof size |
| `BASE_URL` | `http://localhost:8080` | Used to build invitation URLs |
| `ALLOWED_ORIGIN` | `http://localhost:3000` | CORS allowed origin |

**Change the default admin password before any production use.**

## API overview

### Public

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/packages` | Active packages |
| GET | `/api/v1/templates?package=silver` | Templates (optional package filter) |
| GET | `/api/v1/templates/{slug}` | Template detail + supported fields |
| GET | `/demo/{package}/{file}` | Original static template preview from the registered catalog |
| GET | `/templates/personalize.js` | Shared safe personalization runtime used by templates |
| POST | `/api/v1/orders` | Create order (idempotent) |
| GET | `/api/v1/orders/{reference}` | Minimal order status |
| POST | `/api/v1/orders/{reference}/payment` | Upload payment proof (multipart) |
| GET | `/api/v1/invitations/{slug}` | Published invitation JSON |
| GET | `/i/{slug}` | Render invitation HTML |

### Admin (session cookie required)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/admin/auth/login` | Login → sets session cookie |
| POST | `/api/v1/admin/auth/logout` | Logout |
| GET | `/api/v1/admin/orders` | List orders (`?status=&page=&limit=`) |
| GET | `/api/v1/admin/orders/{id}` | Order + payment detail |
| POST | `/api/v1/admin/orders/{id}/verify-payment` | Approve / reject payment |
| POST | `/api/v1/admin/orders/{id}/publish` | Publish invitation with slug |
| POST | `/api/v1/admin/orders/{id}/reject` | Reject order |
| GET | `/api/v1/admin/templates` | All templates |
| PATCH | `/api/v1/admin/templates/{id}` | Toggle availability |

## Order status flow

```
pending_payment
    → payment_submitted   (customer uploads proof)
    → payment_verified    (admin approves)
    → published           (admin publishes with unique slug)
```

Alternative paths: `rejected`, `cancelled`.

Payment statuses: `unpaid` → `submitted` → `verified` | `rejected`.

## Notes

- Prices are always taken from the database at order creation time.
- Order references are high-entropy (`SKM-XXXXXXXXXXXX`); never sequential.
- Payment proofs are stored outside the public web root with random filenames.
- Draft invitations are never served on `/i/{slug}` or `/api/v1/invitations/{slug}`.
- The six original files under `frontend/templates/` are registered by the database seed. `/demo/` serves their unchanged static previews; `/i/{slug}` loads the registered source and injects published invitation JSON for the shared runtime.
- CSRF protection for cookie-based admin routes should be added when the admin UI is built (SameSite=Lax is set).

## Project layout

```
backend/
├── cmd/server/main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── response/
│   ├── services/
│   └── storage/
├── migrations/001_init.sql
├── storage/proofs/          # runtime, not committed
├── go.mod
└── README.md
```
