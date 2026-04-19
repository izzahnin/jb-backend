# 🔧 Setup & Configuration Guide

Complete step-by-step instructions for setting up the Jalur Berlian Backend system.

---

## 📋 Prerequisites

- **Docker & Docker Compose** (Latest version)
  - [Install Docker Desktop](https://www.docker.com/products/docker-desktop)
  - Verify installation: `docker --version` & `docker-compose --version`
  
- **Go 1.25+** (Optional, only for local development without Docker)
  - [Download Go](https://golang.org/dl/)
  - Verify: `go version`

- **Git** (For cloning repository)
  - [Download Git](https://git-scm.com/downloads)

- **Postman** (Recommended, for API testing)
  - [Download Postman](https://www.postman.com/downloads/)

---

## 🚀 Quick Start (Docker - Recommended)

### Step 1: Clone Repository
```bash
git clone <repository-url> jb-backend
cd jb-backend
```

### Step 2: Start Services
```bash
# Start all containers (PostgreSQL, Redis, API server)
docker-compose up -d

# Verify services are running
docker-compose ps
```

Expected output:
```
NAME                 COMMAND                  STATUS
postgres             "docker-entrypoint..."   Up 10 seconds
redis                "docker-entrypoint..."   Up 10 seconds
jb-api               "./jb-api"              Up 5 seconds
```

### Step 3: Verify API is Ready
```bash
# Check logs
docker-compose logs jb-api | tail -20

# Should see: "server is running on port :8080"
```

### Step 4: Test API
```bash
# Quick health check
curl http://localhost:8080/auth/login -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'

# Expected response: Valid JSON with error or token
```

**✅ System is ready!** Proceed to [Testing API](#-testing-api) section.

---

## 🛠️ Local Development Setup (Without Docker)

### Step 1: Install Dependencies

**Go Modules**
```bash
cd jb-backend
go mod download
```

**PostgreSQL 15**
- Download and install locally: https://www.postgresql.org/download/
- Create database:
  ```bash
  createdb jalur_berlian
  ```

**Redis 7**
- Download and install: https://redis.io/download
- Or use Windows Subsystem for Linux (WSL):
  ```bash
  wsl
  sudo apt-get install redis-server
  redis-server
  ```

### Step 2: Configure Environment Variables

Create `.env` file in project root:
```env
# Database
DATABASE_URL=postgres://postgres:password@localhost:5432/jalur_berlian
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=jalur_berlian

# Redis
REDIS_URL=localhost:6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24h

# Server
PORT=8080
```

### Step 3: Run Database Migrations

```bash
# Migrations run automatically on startup
# Or manually using migrate CLI:
migrate -path migrations -database "postgres://postgres:password@localhost:5432/jalur_berlian" up
```

### Step 4: Start API Server

```bash
# Development (hot reload with air)
air

# Or direct compile and run
go build -o jb-api cmd/api/main.go
./jb-api
```

Expected output:
```
[GIN] 2026/03/16 13:30:00 | 200 |     123.456µs |             127.0.0.1 | GET      /health
Server is running on port :8080
```

---

## 🗄️ Database Setup

### Automatic Migration (Docker)
- Migrations run automatically when containers start
- Database schema created from `migrations/*.sql` files

### Manual Migration (Local Dev)

**Option 1: Using migrate tool**
```bash
# Install migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations -database "postgres://user:pass@localhost:5432/jalur_berlian" up
```

**Option 2: Manual SQL execution**
```bash
# Connect to PostgreSQL
psql -U postgres -d jalur_berlian

# Run migration files manually
\i migrations/000001_create_initial_tables.up.sql
```

### Database Schema Overview

**Tables Created**:
- `users` - Internal admin/staff accounts with role hierarchy
- `customers` - Business customer profiles (no login)
- `drivers` - Driver master data and availability status
- `trucks` - Fleet master data with truck type and status
- `orders` - Global customer order (commercial layer)
- `trips` - Operational execution/surat jalan per truck
- `locations` - GPS history per trip (partitioned by time)
- `audit_logs` - Change log for compliance

**Constraints**:
- Order status follows: pending → partial → completed/cancelled
- Trip status follows: pickup → in_transit → delivered/cancelled
- Database-level CHECK constraints prevent invalid enum values
- Foreign keys enforce referential integrity

---

## 🧪 Testing API

### Method 1: Using Postman (Recommended)

1. **Import Collection**
   - Open Postman
   - Click "Import" → Select `Jalur-Berlian-Backend.postman_collection.json`
   - All 21 endpoints loaded with examples

2. **Register Admin User**
   - Find "Auth" folder → POST "Register"
   - Body:
     ```json
     {
       "username": "admin",
       "password": "securepassword123"
     }
     ```
   - Click "Send"

3. **Login & Get Token**
   - POST "Login" with same credentials
   - Response includes `token` field
   - Copy token value

4. **Set Bearer Token**
   - Postman Collections tab → Select your environment
   - Add variable: `token` = <copied token value>
   - Admin endpoints now automatically include Authorization header

5. **Test Endpoints**
   - Try "Create Truck" (POST /admin/trucks)
   - Try "List Trucks" (GET /admin/trucks)
   - Try "Public Tracking" (no auth needed)

### Method 2: Using cURL

```bash
# 1. Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"pass123"}'

# 2. Login (saves token to variable)
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"pass123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4)

# 3. Create truck (uses token)
curl -X POST http://localhost:8080/admin/trucks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"plate_number":"AB1234CD","driver_name":"Budi Santoso"}'

# 4. List trucks
curl -X GET http://localhost:8080/admin/trucks \
  -H "Authorization: Bearer $TOKEN"
```

### Method 3: Using Thunder Client / REST Client VS Code Extension
- Install Thunder Client or REST Client extension in VS Code
- Open any .http file with REST requests
- Click "Send Request" next to each endpoint

---

## 📊 Key Endpoints for Testing

### Authentication
```
POST /auth/register        [Public] Register new admin
POST /auth/login           [Public] Get JWT token
POST /auth/logout          [Protected] Logout
```

### Fleet Management
```
POST /admin/trucks         [Protected] Create truck
GET  /admin/trucks         [Protected] List trucks (pagination available)
GET  /admin/trucks/:id     [Protected] Get single truck
PUT  /admin/trucks/:id     [Protected] Update truck
DELETE /admin/trucks/:id   [Protected] Delete truck
```

### Order Management
```
POST /admin/orders                     [Protected] Create order
GET  /admin/orders                     [Protected] List orders
GET  /admin/orders/:id                 [Protected] Get order details
PUT  /admin/orders/:id/assign          [Protected] Assign truck
PUT  /admin/orders/:id/confirm-pickup  [Protected] Start delivery
PUT  /admin/orders/:id/confirm-delivery [Protected] Complete delivery
```

### Location Tracking
```
POST /locations                        [Public] Record location
GET  /locations/:truck_id/latest       [Public] Get last position
GET  /locations/:truck_id/history      [Public] Get location history
```

### Customer Tracking
```
GET  /orders/:order_number/track       [Public] Track by order number
```

---

## 🐛 Troubleshooting

### Docker Issues

**Problem**: "Cannot connect to Docker daemon"
```bash
# Solution: Start Docker Desktop or Docker service
# Windows: Open Docker Desktop application
# Linux: sudo systemctl start docker
```

**Problem**: "Port 8080 already in use"
```bash
# Solution: Change port in docker-compose.yaml
# Or kill process using the port:
# Windows (PowerShell): Stop-Process -Id (Get-NetTCPConnection -LocalPort 8080).OwningProcess
# Linux: lsof -ti:8080 | xargs kill -9
```

**Problem**: "Database connection refused"
```bash
# Solution: Wait for PostgreSQL to start (can take 30-60 seconds)
docker-compose logs postgres  # Check logs
docker-compose ps             # Check status

# Restart if needed:
docker-compose down
docker-compose up -d
```

### API Issues

**Problem**: "Invalid token" error on protected routes
```bash
# Solution: Make sure to:
# 1. Register user first (POST /auth/register)
# 2. Login to get token (POST /auth/login)
# 3. Include in Authorization header: Bearer <token>
```

**Problem**: "Order status is invalid" on state transitions
```bash
# Solution: Follow state machine:
# pending → pickup (confirm-pickup) → in_transit → delivered (confirm-delivery)
# Cannot skip states or go backward
```

**Problem**: "Database migration failed" on startup
```bash
# Solution: Check database logs
docker-compose logs postgres

# Rebuild and restart
docker-compose down -v  # Remove volumes (WARNING: deletes data)
docker-compose up -d
```

### Development Issues

**Problem**: Hot reload not working (air)
```bash
# Solution: Install air
go install github.com/cosmtrek/air@latest

# Create .air.toml configuration
air init

# Run air
air
```

**Problem**: Go modules not found
```bash
# Solution: Download dependencies
go mod download
go mod tidy
```

---

## 🔐 Environment Variables Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | postgresql://... | PostgreSQL connection string |
| `POSTGRES_USER` | postgres | DB username |
| `POSTGRES_PASSWORD` | - | DB password (required) |
| `POSTGRES_DB` | jalur_berlian | Database name |
| `REDIS_URL` | localhost:6379 | Redis connection |
| `REDIS_PASSWORD` | (empty) | Redis password if set |
| `JWT_SECRET` | - | Secret key for JWT (required, change in production!) |
| `JWT_EXPIRY` | 24h | Token expiration time |
| `PORT` | 8080 | API server port |
| `GIN_MODE` | debug | Set to "release" for production |

---

## 📈 Performance Tips

### Database
- Enable query logging to identify slow queries: `log_statement = 'all'` in postgres.conf
- Use pagination for list endpoints (default limit: 10)
- Index frequently filtered columns

### Redis
- Monitor memory usage: `redis-cli INFO memory`
- Enable persistence: `save 900 1` in redis.conf (save every 15 min if 1+ change)

### API Server
- Enable caching headers on static responses
- Use connection pooling (automatic with sqlx and go-redis)
- Monitor goroutines: `runtime.NumGoroutine()`

---

## 🚀 Production Deployment

### Pre-Deployment Checklist
- [ ] All environment variables set (especially JWT_SECRET)
- [ ] Database backups configured
- [ ] Redis persistence enabled
- [ ] Firewall rules configured (port 8080 accessible, 5432/6379 private)
- [ ] HTTPS/TLS configured (use reverse proxy nginx)
- [ ] Log aggregation setup
- [ ] Monitoring alert configured

### Docker Production Deployment
```bash
# Build production image
docker build -t jb-api:1.0.0 -f Dockerfile .

# Run with production settings
docker run -d \
  --name jb-api \
  -p 8080:8080 \
  -e DATABASE_URL=<prod-db-url> \
  -e REDIS_URL=<prod-redis-url> \
  -e JWT_SECRET=<strong-secret> \
  -e GIN_MODE=release \
  jb-api:1.0.0
```

### Kubernetes Deployment (Optional)
- Create ConfigMap for configuration
- Create Secret for sensitive data (JWT_SECRET, passwords)
- Use Persistent Volume for database
- Deploy with multiple replicas for high availability

---

## 📚 Additional Resources

- **Go Documentation**: https://golang.org/doc
- **Gin Framework**: https://gin-gonic.com
- **PostgreSQL**: https://www.postgresql.org/docs
- **Redis**: https://redis.io/documentation
- **JWT**: https://jwt.io
- **Docker Compose**: https://docs.docker.com/compose

---

**Need help?** Check the logs:
```bash
# Docker
docker-compose logs -f jb-api

# Local dev
# Check console output or application logs
```

---

Last Updated: 2026-03-16 | Version: 1.0.0
