# Holy Home

[![CI](https://github.com/Sainaif/home-app/actions/workflows/ci.yml/badge.svg)](https://github.com/Sainaif/home-app/actions/workflows/ci.yml)
[![Docker Publish](https://github.com/Sainaif/home-app/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/Sainaif/home-app/actions/workflows/docker-publish.yml)

A self-hosted household management app for shared living. Track bills, split costs fairly, manage loans, and keep everyone accountable.

## Features

- **Bill Management** - Track electricity, gas, internet, and custom bills
- **Smart Cost Splitting** - Automatically split costs based on actual usage or equally
- **Meter Readings** - Record consumption data for accurate billing
- **Loan Tracking** - Keep track of money borrowed and lent between residents
- **Balance Overview** - See who owes what at a glance
- **Household Supplies** - Track shared supplies and reimbursements
- **Chore Management** - Assign and rotate household tasks
- **Multi-auth Support** - Email, username, passkeys, and optional 2FA

## Quick Start

**Requirements:** Docker and Docker Compose

```bash
# Clone the repository
git clone https://github.com/Sainaif/home-app.git
cd home-app

# Configure environment
cp .env.example .env
# Edit .env - set ADMIN_EMAIL, ADMIN_PASSWORD, and generate JWT secrets

# Start the application
cd deploy
docker-compose up -d
```

Open http://localhost:16161 and log in with your admin credentials.

> The admin account is created automatically on first startup.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.24, Fiber v2 |
| Database | MongoDB 8.0 |
| Frontend | Vue 3, Vite, Tailwind CSS, Pinia |
| Auth | JWT, Argon2id, WebAuthn, TOTP |
| Deployment | Docker, Docker Compose |

## Configuration

All configuration is done via environment variables. Copy `.env.example` to `.env` and customize:

### Required Settings

| Variable | Description |
|----------|-------------|
| `JWT_SECRET` | Secret for access tokens (generate with `openssl rand -base64 32`) |
| `JWT_REFRESH_SECRET` | Secret for refresh tokens (must be different from JWT_SECRET) |
| `ADMIN_EMAIL` | Email for the initial admin account |
| `ADMIN_PASSWORD` | Password for the initial admin account |

### Authentication Options

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTH_ALLOW_EMAIL_LOGIN` | `true` | Allow login with email |
| `AUTH_ALLOW_USERNAME_LOGIN` | `false` | Allow login with username |
| `AUTH_REQUIRE_USERNAME` | `false` | Require username during registration |
| `AUTH_2FA_ENABLED` | `false` | Enable TOTP two-factor authentication |

### Application Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `development` | Environment (`development` or `production`) |
| `APP_DOMAIN` | `localhost` | Domain for WebAuthn (must match actual domain) |
| `APP_BASE_URL` | `http://localhost:16162` | Full URL of the application |

See `.env.example` for the complete list of options.

## How Bill Splitting Works

The app intelligently splits bills based on type:

**Metered bills (electricity):**
- Personal usage from your meter → charged to you
- Common areas (hallway, kitchen, etc.) → split equally among all residents

**Flat-rate bills (internet, gas):**
- Split equally by default
- Can be customized per bill

## Development

### Backend

```bash
cd backend
go mod tidy
go run ./cmd/api
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### Running Tests

```bash
# Backend (requires MongoDB)
cd backend
go test -v -race ./...

# Frontend
cd frontend
npm test
```

### Rebuilding Docker Images

```bash
cd deploy
docker-compose build && docker-compose up -d
```

## Architecture

```
├── backend/
│   ├── cmd/api/          # Application entrypoint
│   └── internal/
│       ├── config/       # Environment configuration
│       ├── database/     # MongoDB connection
│       ├── handlers/     # HTTP route handlers
│       ├── middleware/   # Auth, rate limiting
│       ├── models/       # Data structures
│       ├── services/     # Business logic
│       └── utils/        # JWT, password hashing, TOTP
│
├── frontend/
│   └── src/
│       ├── api/          # Axios client
│       ├── components/   # Reusable UI components
│       ├── composables/  # Vue 3 composition functions
│       ├── locales/      # Translations (Polish)
│       ├── stores/       # Pinia state management
│       └── views/        # Page components
│
└── deploy/               # Docker Compose files
```

## API

The API runs on port `16162` by default.

### Endpoints Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/auth/login` | Authenticate user |
| `POST` | `/auth/refresh` | Refresh access token |
| `GET` | `/auth/config` | Get auth configuration |
| `GET` | `/users/me` | Get current user |
| `GET` | `/bills` | List bills |
| `POST` | `/bills` | Create bill |
| `GET` | `/groups` | List groups |
| `GET` | `/balance` | Get balance summary |

All endpoints except `/auth/*` require authentication via Bearer token.

## Database Collections

| Collection | Description |
|------------|-------------|
| `users` | User accounts and credentials |
| `groups` | Household groups with cost-splitting weights |
| `bills` | Utility bills with type, amount, period |
| `consumptions` | Meter readings per user/group |
| `allocations` | Calculated cost splits |
| `payments` | Payment records |
| `loans` | Money lent/borrowed between residents |
| `chores` | Household tasks and assignments |
| `supplies` | Shared household supplies |

## Security

- **Password Hashing** - Argon2id with secure parameters
- **JWT Tokens** - Short-lived access (15min) + long-lived refresh (30 days)
- **WebAuthn/Passkeys** - Passwordless authentication support
- **TOTP 2FA** - Optional two-factor authentication
- **Rate Limiting** - 5 login attempts per 15 minutes
- **CORS** - Configurable allowed origins

## License

MIT
