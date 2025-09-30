# Getting Started with Holy Home

Quick guide to get the Holy Home application running locally or in Docker.

## Prerequisites

- **Docker & Docker Compose** (recommended)
- **OR** for local development:
  - Go 1.25+
  - Python 3.13+
  - MongoDB 8.0+
  - Node.js (current)

## Quick Start with Docker

### 1. Configure Environment

```bash
# Copy example environment file
cp .env.example .env

# Generate a secure admin password hash
# You can use any Argon2 tool or online generator
# Example password "ChangeMe123!" hash:
# $argon2id$v=19$m=65536,t=3,p=1$CHANGE$THIS$HASH

# Edit .env and set:
# - ADMIN_EMAIL
# - ADMIN_PASSWORD_HASH
# - JWT_SECRET (random string, 32+ chars)
# - JWT_REFRESH_SECRET (different random string, 32+ chars)
```

### 2. Start Services

```bash
cd deploy
docker-compose up -d
```

### 3. Verify Services

```bash
# Check health
curl http://localhost:8080/healthz
# Should return: {"status":"ok","time":"..."}

curl http://localhost:8000/healthz
# Should return: {"status":"ok","service":"ml","time":"..."}
```

### 4. Test Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.pl","password":"ChangeMe123!"}'
```

You should receive access and refresh tokens.

## Local Development

### Backend (Go API)

```bash
cd backend

# Install dependencies
go mod tidy

# Run locally
export MONGO_URI=mongodb://localhost:27017
export MONGO_DB=holyhome
export ADMIN_EMAIL=admin@example.pl
export ADMIN_PASSWORD_HASH='$argon2id$v=19$m=65536,t=3,p=1$...'
export JWT_SECRET=your-secret-here
export JWT_REFRESH_SECRET=your-refresh-secret-here
export ML_BASE_URL=http://localhost:8000

go run ./cmd/api

# Or build first
go build ./cmd/api
./api
```

API will be available at: http://localhost:8080

### ML Sidecar (Python)

```bash
cd ml

# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run
python -m app.main
```

ML service will be available at: http://localhost:8000

### Frontend (Vue 3) - TODO

```bash
cd frontend
npm install
npm run dev
```

## API Examples

### 1. Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.pl",
    "password": "ChangeMe123!"
  }'
```

Save the returned `access` token for subsequent requests.

### 2. Create a User

```bash
TOKEN="your-access-token-here"

curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "email": "user@example.pl",
    "role": "RESIDENT",
    "tempPassword": "TempPass123!"
  }'
```

### 3. Create a Group

```bash
curl -X POST http://localhost:8080/groups \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Couple 1",
    "weight": 2.0
  }'
```

### 4. Create a Bill

```bash
curl -X POST http://localhost:8080/bills \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "type": "electricity",
    "periodStart": "2025-09-01T00:00:00Z",
    "periodEnd": "2025-09-30T23:59:59Z",
    "totalAmountPLN": 450.00,
    "totalUnits": 300.0,
    "notes": "September 2025"
  }'
```

### 5. Record Consumption

```bash
BILL_ID="bill-id-from-previous-response"
USER_ID="user-id-from-user-creation"

curl -X POST http://localhost:8080/consumptions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "billId": "'$BILL_ID'",
    "userId": "'$USER_ID'",
    "units": 150.5,
    "meterValue": 12345.5,
    "recordedAt": "2025-09-30T12:00:00Z"
  }'
```

### 6. Allocate Bill Costs

```bash
curl -X POST http://localhost:8080/bills/$BILL_ID/allocate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "strategy": "proportional"
  }'
```

### 7. Generate Forecast (ML Sidecar)

```bash
curl -X POST http://localhost:8000/forecast \
  -H "Content-Type: application/json" \
  -d '{
    "target": "electricity",
    "historical_dates": [
      "2025-01-01T00:00:00Z",
      "2025-02-01T00:00:00Z",
      "2025-03-01T00:00:00Z",
      "2025-04-01T00:00:00Z",
      "2025-05-01T00:00:00Z"
    ],
    "historical_values": [250, 280, 265, 270, 285],
    "horizon_months": 3,
    "cost_per_unit": 1.5
  }'
```

## Directory Structure

```
home-app/
├── backend/              # Go API
│   ├── cmd/api/         # Main application entry point
│   ├── internal/
│   │   ├── config/      # Configuration management
│   │   ├── database/    # MongoDB connection
│   │   ├── handlers/    # HTTP handlers
│   │   ├── middleware/  # Auth, logging, etc.
│   │   ├── models/      # Data models
│   │   ├── services/    # Business logic
│   │   └── utils/       # Utilities (JWT, crypto, etc.)
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── ml/                   # Python ML Service
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py      # FastAPI app
│   │   ├── models.py    # Pydantic models
│   │   └── forecaster.py # Forecasting engine
│   ├── requirements.txt
│   └── Dockerfile
│
├── frontend/            # Vue 3 Frontend (TODO)
│   ├── src/
│   ├── Dockerfile
│   └── package.json
│
├── deploy/              # Docker Compose
│   └── docker-compose.yml
│
├── .env.example
├── .gitignore
├── README.md
├── IMPLEMENTATION_STATUS.md
└── GETTING_STARTED.md
```

## Useful Commands

### Docker

```bash
# Start all services
cd deploy && docker-compose up -d

# View logs
docker-compose logs -f api
docker-compose logs -f ml

# Stop services
docker-compose down

# Rebuild after code changes
docker-compose build
docker-compose up -d
```

### MongoDB

```bash
# Connect to MongoDB
docker exec -it deploy-mongo-1 mongosh

# In mongosh:
use holyhome
db.users.find()
db.bills.find()
```

### Backend Development

```bash
cd backend

# Run tests
go test ./...

# Format code
go fmt ./...

# Lint
golangci-lint run

# Build
go build ./cmd/api
```

### ML Development

```bash
cd ml

# Run with auto-reload
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

# Test forecast endpoint
python -c "import requests; print(requests.post('http://localhost:8000/forecast', json={...}).json())"
```

## Troubleshooting

### "Failed to connect to MongoDB"
- Ensure MongoDB is running: `docker ps | grep mongo`
- Check connection string in `.env`
- Wait for MongoDB to be fully initialized (30s on first start)

### "Invalid credentials"
- Verify `ADMIN_PASSWORD_HASH` in `.env`
- Ensure password matches the hash
- Check admin was created: `docker-compose logs api | grep "Admin bootstrap"`

### "ML service not responding"
- Check ML service health: `curl http://localhost:8000/healthz`
- View logs: `docker-compose logs ml`
- Verify Python dependencies installed correctly

### Port already in use
- Check what's using the port: `lsof -i :8080`
- Change port in `.env` (`APP_PORT=8081`)
- Update docker-compose.yml port mapping

## Security Notes

- **Never commit `.env` files** (already in .gitignore)
- Change default passwords immediately
- Use strong random strings for JWT secrets
- Enable TOTP 2FA for all users in production
- Use HTTPS in production (via reverse proxy)
- Regularly update dependencies

## Next Steps

1. ✅ Backend API is fully functional
2. ✅ ML sidecar is ready
3. ⏳ Build the Vue 3 frontend
4. ⏳ Add SSE support for real-time updates
5. ⏳ Implement CSV/PDF exports

## Support

- Check [README.md](README.md) for detailed documentation
- See [IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md) for current progress
- Review prompt.txt for full specification