# Environment Variables Guide

## 🔐 Security Principles

**CRITICAL RULES:**
1. ❌ **NEVER hardcode secrets** (passwords, API keys, JWT secrets) in source code
2. ✅ **ALL secrets MUST be in `.env.local`** which is `.gitignored`
3. ✅ **`.env` is a template** showing what variables are needed
4. ✅ **Fallback values only for non-secrets** (ports, log levels)

---

## 📊 The Flow: How Environment Variables Work

```
┌─────────────────────────────────────────────────────────────────┐
│  DEVELOPER MACHINE SETUP                                        │
└─────────────────────────────────────────────────────────────────┘

STEP 1: .env Template (TEMPLATE - COMMITTED)
═══════════════════════════════════════════════════════════════════
File: .env (in git ✅)
  DB_SOURCE=postgres://admin:change_me_in_env_local@...
  JWT_SECRET=change_me_with_random_secret_in_env_local
  REDIS_ADDR=localhost:6379

Purpose: 
  - Template showing required variables
  - Safe example values (not real credentials)
  - Team reference: "what variables does this project need?"


STEP 2: .env.local Override (SECRETS - IGNORED)
═══════════════════════════════════════════════════════════════════
File: .env.local (NOT in git ✅ - .gitignore blocks it)
  DB_SOURCE=postgres://admin:myRealPassword123@localhost:5432/jalur_berlian_db?sslmode=disable
  JWT_SECRET=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6
  REDIS_ADDR=localhost:6379

Purpose:
  - Local development secrets
  - Real passwords, real API keys
  - NEVER committed to git
  - Each developer has their own copy


STEP 3: docker-compose.yaml Loads Variables
═══════════════════════════════════════════════════════════════════
docker-compose.yaml:
  services:
    jb-api:
      environment:
        DB_SOURCE: ${DB_SOURCE}       ← Environment substitution
        JWT_SECRET: ${JWT_SECRET}
        REDIS_ADDR: ${REDIS_ADDR}

Process:
  1. Reads .env (default)
  2. Overrides with .env.local if exists
  3. Passes to container


STEP 4: main.go READS Environment Variables (REQUIRED)
═══════════════════════════════════════════════════════════════════
Code:
  // Secrets MUST be present - no fallback!
  dbSource := os.Getenv("DB_SOURCE")
  if dbSource == "" {
    log.Fatal("❌ DB_SOURCE not set!")  // FAIL
  }
  
  jwtSecret := os.Getenv("JWT_SECRET")
  if jwtSecret == "" {
    log.Fatal("❌ JWT_SECRET not set!")  // FAIL
  }
  
  // Non-secrets can have fallback
  redisAddr := os.Getenv("REDIS_ADDR")
  if redisAddr == "" {
    log.Fatal("❌ REDIS_ADDR not set!")  // FAIL
  }

STEP 5: Application Uses Real Values
═══════════════════════════════════════════════════════════════════
  database.NewPostgres(dbSource)  ← Uses real DB password from env
  usecase.NewAuthUsecase(userRepo, jwtSecret)  ← Uses real secret
```

---

## 🔒 Secrets Classification

### 🚨 REQUIRED SECRETS (Must be in .env.local)

| Variable | Purpose | Example Format | Fallback Policy |
|----------|---------|-----------------|-----------------|
| **DB_SOURCE** | PostgreSQL connection | `postgres://user:pass@host:5432/db?sslmode=disable` | ❌ NO - Must fail if not set |
| **JWT_SECRET** | Token signing key | Random 64-char string from `openssl rand -hex 32` | ❌ NO - Must fail if not set |
| **REDIS_ADDR** | Redis server address | `localhost:6379` or `redis.cloud:port` | ❌ NO - Must fail if not set |

### ⚙️ NON-SECRETS (Can have safe fallbacks or defaults)

| Variable | Purpose | Safe Default | Fallback Policy |
|----------|---------|------------------------|-----------------|
| **PORT** | Server port | `8080` | ✅ Fallback OK |
| **GIN_MODE** | Gin framework mode | `release` | ✅ Fallback OK |
| **LOG_LEVEL** | Logging verbosity | `info` | ✅ Fallback OK |

---

## 📝 Setup Instructions for New Developers

### 1. Clone Repository
```bash
git clone https://github.com/izzahnin/jb-backend.git
cd jb-backend
```

### 2. Copy Template to Local Secrets File
```bash
cp .env.local.example .env.local
```

### 3. Edit `.env.local` with Real Values
```bash
# Linux/Mac
nano .env.local
# or
vim .env.local

# Windows
notepad .env.local
```

Update each variable:
```env
# Database - use your actual password
DB_SOURCE=postgres://admin:your_actual_password@localhost:5432/jalur_berlian_db?sslmode=disable

# JWT Secret - generate random string
JWT_SECRET=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6

# Redis (if different from default)
REDIS_ADDR=localhost:6379
```

### 4. Generate Secure JWT Secret

**Windows PowerShell:**
```powershell
$bytes = New-Object byte[] 32
$rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
$rng.GetBytes($bytes)
[BitConverter]::ToString($bytes) -replace '-', ''
```

**Linux/Mac:**
```bash
openssl rand -hex 32
```

### 5. Start Services
```bash
docker-compose up -d
```

### 6. Verify Application Started
```bash
docker logs jbm_api
# Should see: "Server PT. Jalur Berlian berjalan di port :8080"
```

---

## ✅ Security Checklist

- [ ] `.env.local` is in `.gitignore`
- [ ] `.env.local` contains actual passwords/secrets
- [ ] `.env` contains only template values (change_me_*)
- [ ] `main.go` reads secrets from environment (no hardcoding)
- [ ] Secrets produce error if not set (fail-fast principle)
- [ ] `.env.local` is never committed to git
- [ ] Random JWT_SECRET is generated for each environment

---

## 🔍 Verification

To verify secrets are properly configured:

```bash
# Check what variables are being read
docker logs jbm_api | grep -i "env\|secret\|database"

# Test API endpoint
curl -s http://localhost:8080/swagger/docs

# Verify database connection
docker exec jbm_postgres psql -U admin -d jalur_berlian_db -c "\dt"
```

---

## 🚀 Production Deployment

For production, use environment variables from:
- **Cloud Platform**: Railway, Render, AWS, etc.
- **Or**: CI/CD secrets (GitHub Secrets, GitLab Variables)
- **Never commit** production secrets to repository

Example with Railway.app:
```
Set environment variables in Railway dashboard:
  DB_SOURCE=postgres://prod_user:prod_password@prod-db.railway.app:5432/jalur_berlian_prod?sslmode=require
  JWT_SECRET=<production-random-secret>
  REDIS_ADDR=redis.railway.app:6380
  GIN_MODE=release
```

---

## ❓ FAQ

**Q: Why not commit `.env.local`?**
A: It contains sensitive information (passwords, API keys). Committing it exposes secrets to anyone with git access.

**Q: Why fail if environment variable is not set?**
A: Better to fail immediately (fail-fast) than to run with default/wrong values and cause issues later.

**Q: Can I share my `.env.local`?**
A: NO. Each developer/environment should have their own `.env.local` with their own secrets.

**Q: What if I forgot to update `.env.local` before running?**
A: The application will fail with clear error message showing what's missing. Fix and retry.

**Q: Is `.env` safe to commit?**
A: YES. `.env` contains only template values and is safe. The actual secrets are in `.env.local` (ignored).
