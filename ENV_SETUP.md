# Environment Setup Guide

## Quick Start

### 1. Untuk Development
```bash
# Copy template ke .env.local
cp .env.local.example .env.local

# Edit .env.local dengan credentials lokal Anda
# .env.local adalah PRIVATE dan TIDAK akan dicommit ke git
```

### 2. File Structure

| File | Committed? | Purpose |
|------|-----------|---------|
| `.env` | ✅ YES | Template dengan contoh values - AMAN untuk commit |
| `.env.local` | ❌ NO | Actual credentials lokal Anda - RAHASIA (di .gitignore) |
| `.env.local.example` | ✅ YES | Template untuk membuat .env.local baru |
| `.env.production` | ❌ NO | Production config - RAHASIA |
| `.env.production.local` | ❌ NO | Production overrides - RAHASIA |

### 3. Variable yang Tersedia

#### Database
- `DB_SOURCE`: PostgreSQL connection string
  - Format: `postgres://user:password@host:port/dbname?params`
  - Example: `postgres://admin:mypassword@localhost:5432/jalur_berlian_db?sslmode=disable`

#### Redis
- `REDIS_ADDR`: Redis server address (default: localhost:6379)
- `REDIS_PASSWORD`: Redis password (empty for local development)
- `REDIS_DB`: Redis database number (default: 0)

#### Security
- `JWT_SECRET`: Secret key untuk signing JWT tokens
  - HARUS berupa string random yang panjang
  - Minimal 32 karakter untuk production
  - Example: `vqH8mK2pL9xR4jN6bZ3cW5fY7uIoAsD1`

#### Server
- `PORT`: Server port (default: 8080)
- `GIN_MODE`: Gin framework mode (debug, release)

#### Logging
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

## Best Practices

### ✅ DO
- Commit `.env` sebagai template dengan placeholder values
- Commit `.env.local.example` untuk referensi
- Commit `.gitignore` dengan daftar lengkap file sensitif
- Keep `.env.local` di local machine saja
- Use strong random strings untuk JWT_SECRET

### ❌ DON'T
- Commit `.env.local` dengan actual credentials
- Commit production passwords ke repository
- Share `.env.local` dengan orang lain
- Use hardcoded secrets dalam code
- Reuse JWT_SECRET di multiple environments

## Cara Membaca Environment Variables di Go

```go
import "os"

func init() {
    // Load dari .env atau .env.local
    dbSource := os.Getenv("DB_SOURCE")
    redisAddr := os.Getenv("REDIS_ADDR")
    jwtSecret := os.Getenv("JWT_SECRET")
}
```

Atau gunakan library `github.com/joho/godotenv`:
```go
import "github.com/joho/godotenv"

func init() {
    godotenv.Load(".env.local") // Priority: .env.local
    godotenv.Load(".env")        // Fallback: .env
}
```

## Production Deployment

1. Set environment variables di server (Docker, K8s, CI/CD):
   ```bash
   export DB_SOURCE="postgres://prod_user:strong_password@prod_host:5432/prod_db"
   export REDIS_ADDR="redis-prod.example.com:6379"
   export JWT_SECRET="$(openssl rand -hex 32)"  # Generate random secret
   export GIN_MODE="release"
   ```

2. Jangan gunakan .env.production file di production
3. Gunakan secrets management (HashiCorp Vault, AWS Secrets Manager, etc)

## Troubleshooting

### Port 8080 sudah digunakan
```bash
# Ubah di .env.local
PORT=8081
```

### Redis connection failed
```bash
# Check Redis sudah running
redis-cli ping  # Should return PONG

# Verify REDIS_ADDR di .env.local
# Example: REDIS_ADDR=localhost:6379
```

### Database connection string error
```bash
# Format yang benar: postgres://user:password@host:port/database?params
# Pastikan password tidak punya special chars yang butuh escape
# Atau gunakan URL encoding: P@ss%40word
```
