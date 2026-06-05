# StayFlow Track

[![CI](https://github.com/mdshoaibumer/stayflowtrack/actions/workflows/ci.yml/badge.svg)](https://github.com/mdshoaibumer/stayflowtrack/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mdshoaibumer/stayflowtrack)](https://goreportcard.com/report/github.com/mdshoaibumer/stayflowtrack)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> Production-grade multi-tenant SaaS platform for Service Apartment and Boutique Hotel Management. Built with Go, PostgreSQL, and Next.js — designed for single-VPS deployment.

---

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [Deployment](#deployment)
- [API Documentation](#api-documentation)
- [Security](#security)
- [Contributing](#contributing)
- [License](#license)

---

## Features

### Property Management
- Multi-property support per tenant
- Unit types with dynamic pricing
- Real-time availability calendar
- Maintenance blocking & room moves

### Reservation Engine
- Conflict detection with database-level enforcement
- Status machine: Pending → Confirmed → Checked-In → Checked-Out
- No-show handling with automated charges
- Stay extension workflow
- Corporate reservation support

### Billing & Invoicing
- GST-compliant folio management (CGST/SGST/IGST)
- Decimal-precision arithmetic (no floating-point errors)
- PDF invoice generation
- Payment tracking with multiple methods
- Deposit management with refund workflows

### Housekeeping & Laundry
- Task assignment and status tracking
- Integration with reservation check-in/check-out
- Laundry order management with folio posting

### Multi-Tenant SaaS
- Subscription billing with Razorpay integration
- Plan-based feature gating
- Tenant onboarding wizard
- Platform admin dashboard

### Security
- JWT authentication with token rotation
- Role-based access control (RBAC)
- Rate limiting (sliding window)
- HMAC webhook verification
- Row-Level Security readiness
- Audit logging

---

## Architecture

### Tech Stack
- **Backend**: Go 1.25+, Chi Router, PostgreSQL 16, sqlc, JWT
- **Frontend**: Next.js 15, TypeScript, Tailwind CSS, shadcn/ui
- **Infrastructure**: Docker Compose, Caddy (SSL), MinIO (S3-compatible)
- **Payments**: Razorpay

### Architecture Principles
- **Modular Monolith** with clear module boundaries
- **Domain-Driven Design** with bounded contexts per module
- **Clean Architecture** (Handler → Service → Repository)
- **Multi-Tenant** with tenant isolation at every layer
- **API Versioned** with `/api/v1/` prefix
- **Defense-in-Depth** security model

### Module Boundaries

```
┌─────────────────────────────────────────────────────────┐
│                      API Layer                          │
│  /api/v1/auth  /api/v1/properties  /api/v1/guests  ... │
├─────────────────────────────────────────────────────────┤
│                   Middleware Layer                       │
│  Authentication │ Tenant Context │ RBAC │ Rate Limit    │
├──────────┬──────────┬──────────┬────────────────────────┤
│   Auth   │ Property │  Guest   │    Reservation         │
│  Module  │  Module  │  Module  │      Module            │
├──────────┼──────────┼──────────┼────────────────────────┤
│ Billing  │ Calendar │  HK      │  Laundry │ Operations  │
│  Module  │  Module  │  Module  │  Module  │  Module     │
├──────────┴──────────┴──────────┴────────────────────────┤
│                   Shared Platform                        │
│  Database │ Storage │ Logger │ Errors │ Audit │ Money   │
├─────────────────────────────────────────────────────────┤
│                   PostgreSQL + S3                        │
└─────────────────────────────────────────────────────────┘
```

### Project Structure

```
backend/
├── cmd/server/main.go          # Application entry point & composition root
├── internal/
│   ├── config/                 # Configuration (env-based, production validation)
│   ├── platform/               # Infrastructure
│   │   ├── database/           # Connection pool, transactions, tenant context, timeouts
│   │   ├── logger/             # Structured logging (zerolog)
│   │   └── storage/            # S3-compatible file storage
│   ├── modules/
│   │   ├── auth/               # Authentication, authorization, RBAC
│   │   ├── property/           # Properties, unit types, units
│   │   ├── guest/              # Guest profiles, documents
│   │   ├── reservation/        # Reservation lifecycle
│   │   ├── calendar/           # Availability views, drag-and-drop
│   │   ├── checkinout/         # Check-in/check-out workflows
│   │   ├── billing/            # Folios, line items, payments, invoices
│   │   ├── housekeeping/       # Task management
│   │   ├── laundry/            # Order management
│   │   ├── notifications/      # Templates, providers (WhatsApp/SMS)
│   │   ├── dashboard/          # Metrics & KPIs
│   │   ├── operations/         # No-show, extend stay, room move, deposits
│   │   └── saas/               # Plans, subscriptions, Razorpay billing
│   └── shared/                 # Cross-cutting concerns
│       ├── audit/              # Audit logging
│       ├── errors/             # Structured error types
│       ├── middleware/         # Rate limiting, body size, metrics, logging
│       ├── money/              # Decimal arithmetic helpers
│       ├── pagination/         # Cursor & offset pagination
│       ├── response/           # JSON response formatters
│       └── validation/         # Input validation (go-playground/validator)
├── migrations/                 # 12 sequential migrations
├── sqlc/                       # sqlc queries & config
├── Dockerfile                  # Multi-stage build
└── Makefile                    # Dev commands

frontend/
├── src/app/                    # Next.js App Router
├── src/components/             # UI components per module
├── src/lib/                    # API client & utilities
└── Dockerfile

scripts/
├── backup.sh                   # Automated PostgreSQL backup (cron + S3)
└── restore.sh                  # Point-in-time restore
```

---

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose v2
- Node.js 22+ (for frontend)

### Quick Start

```bash
# Clone the repository
git clone https://github.com/mdshoaibumer/stayflowtrack.git
cd stayflowtrack

# Copy environment configuration
cp .env.example .env
# Edit .env with your secrets (JWT keys, DB password, etc.)

# Start all services
docker compose up -d

# Or for development (infrastructure only):
docker compose up -d postgres minio createbuckets

# Run backend migrations
cd backend && make migrate-up

# Start backend (development)
make run

# Start frontend (separate terminal)
cd ../frontend && npm install && npm run dev
```

### Development Commands

```bash
# Backend
make run               # Run server with hot reload
make test              # Run tests with race detector
make test-coverage     # Generate coverage report
make lint              # Run golangci-lint
make build             # Build production binary
make migrate-up        # Apply all migrations
make migrate-down      # Rollback last migration

# Docker
docker compose up -d             # Start (development)
docker compose -f docker-compose.prod.yml up -d  # Start (production)
```

---

## Deployment

### Single VPS (Recommended)

```
VPS (4 CPU / 8GB RAM minimum)
├── Caddy (SSL termination, auto Let's Encrypt)
├── StayFlow Backend (Go binary, stateless)
├── PostgreSQL 16 (data volume)
├── MinIO (document storage)
└── Cron (pg_dump every 6h → S3)
```

```bash
# Production deployment
cp .env.example .env
# Configure production secrets in .env
docker compose -f docker-compose.prod.yml up -d
```

Caddy automatically provisions SSL certificates for your domain.

### Environment Variables

See [`.env.example`](.env.example) for all configuration options. Critical production variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `DB_PASSWORD` | PostgreSQL password | Yes |
| `JWT_ACCESS_SECRET` | Access token signing key (≥32 chars) | Yes |
| `JWT_REFRESH_SECRET` | Refresh token signing key (≥32 chars) | Yes |
| `PLATFORM_TENANT_ID` | UUID of platform operator tenant | Yes |
| `DOMAIN_API` | API domain (for Caddy) | Production |
| `DOMAIN_APP` | Frontend domain (for Caddy) | Production |

---

## API Documentation

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new tenant |
| POST | `/api/v1/auth/login` | Login (returns JWT) |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/password-reset/request` | Request password reset |
| POST | `/api/v1/auth/password-reset/confirm` | Confirm password reset |

### Properties

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/properties` | Create property |
| GET | `/api/v1/properties` | List properties |
| GET | `/api/v1/properties/:id` | Get property details |
| PUT | `/api/v1/properties/:id` | Update property |

### Reservations

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/reservations` | Create reservation |
| GET | `/api/v1/reservations` | List reservations |
| GET | `/api/v1/reservations/:id` | Get reservation |
| POST | `/api/v1/reservations/:id/cancel` | Cancel reservation |
| GET | `/api/v1/reservations/availability` | Check availability |

### Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/operations/check-in` | Check-in guest |
| POST | `/api/v1/operations/check-out` | Check-out guest |
| POST | `/api/v1/operations/no-show` | Mark as no-show |
| POST | `/api/v1/operations/extend-stay` | Extend stay |
| POST | `/api/v1/operations/room-move` | Move room |
| POST | `/api/v1/operations/maintenance-block` | Block for maintenance |
| POST | `/api/v1/operations/refund-deposit` | Refund deposit |

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": [{"field": "email", "message": "must be a valid email"}]
  }
}
```

---

## Security

- **Authentication**: JWT with 15-min access tokens + 7-day refresh tokens (hashed, rotated)
- **Authorization**: RBAC with 4 roles (super_admin, property_admin, receptionist, housekeeping)
- **Rate Limiting**: 5 req/min on auth endpoints, 100 req/min global (sliding window per IP)
- **Body Size Limits**: 1MB global, 10MB for file uploads
- **Tenant Isolation**: Application-layer filtering + PostgreSQL RLS policies (defense-in-depth)
- **Secrets**: No hardcoded values; environment-variable based with production validation
- **Audit Logging**: All critical operations logged with actor, entity, old/new values
- **Webhooks**: HMAC-SHA256 signature verification
- **SSL**: Automatic via Caddy with HSTS headers
- **Password Storage**: bcrypt (default cost)

---

## Database

### Migrations

12 sequential migrations covering:
1. Core schema (tenants, users, roles)
2. Property management
3. Guest management
4. Reservation management
5. Billing engine
6. Housekeeping
7. Laundry
8. Notifications
9. SaaS management
10. Performance indexes (pg_trgm, partial, composite)
11. Row-Level Security policies
12. Hospitality operations (maintenance, deposits, corporate)

### Key Design Decisions

- **Decimal arithmetic** for all monetary values (`shopspring/decimal`)
- **Partial indexes** for hot query paths (active reservations, open folios)
- **Trigram indexes** for fuzzy guest search
- **RLS policies** as defense-in-depth (not enforced by default — requires restricted DB role)
- **Automated backups** every 6 hours with 7-day retention

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit with conventional commits (`feat:`, `fix:`, `docs:`, etc.)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards

- Go: `gofmt` + `golangci-lint`
- Commit messages: [Conventional Commits](https://www.conventionalcommits.org/)
- Tests required for business logic
- No `float64` for monetary values

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Built with ❤️ for the hospitality industry.
