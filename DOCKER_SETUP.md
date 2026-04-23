# 🚀 PushIQ - Docker Setup Guide

This guide shows how to run PushIQ locally using Docker for PostgreSQL.

## Prerequisites

- **Docker Desktop** installed ([download here](https://www.docker.com/products/docker-desktop))
- **Go 1.22+** installed
- **Node.js 16+** installed

## ✅ Quick Setup (3 Steps)

### Step 1: Start PostgreSQL in Docker (2 minutes)

```bash
# From the project root directory
cd /Users/mac/Documents/my-projects/PushIQ

# Start PostgreSQL container
docker-compose up -d

# Verify it's running
docker-compose ps
```

You should see:
```
CONTAINER ID   IMAGE              STATUS
xxxxx          postgres:15-alpine Up 30 seconds
```

### Step 2: Start Backend API (1 minute)

```bash
cd backend

# Build (if not already built)
go build -o pushiq-api ./cmd/pushiq-api

# Run with environment variables from .env file
set -a
source .env
set +a
./pushiq-api
```

Expected output:
```
DEBU[0000] Config loaded successfully
INFO[0000] Server starting on :8080
```

**Backend is now running at:** `http://localhost:8080`

### Step 3: Start Frontend (1 minute)

```bash
cd frontend

# Install dependencies (first time only)
npm install

# Start dev server
npm run dev
```

Expected output:
```
VITE v4.x.x  ready in 234 ms

➜  Local:   http://localhost:3000/
```

**Dashboard is now running at:** `http://localhost:3000`

---

## 🧪 Test Everything Works

### Test Backend Health
```bash
curl http://localhost:8080/health
```

Should respond:
```json
{"status":"healthy","version":"1.0.0"}
```

### Test Frontend
1. Open browser to `http://localhost:3000`
2. Enter API key: `test-key-12345`
3. Click Login
4. You should see the analytics dashboard!

### Test Notifications Endpoint
```bash
curl -X POST http://localhost:8080/api/v1/devices/register \
  -H "X-Api-Key: test-key-12345" \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "test-device-1",
    "platform": "android"
  }'
```

---

## 📁 Environment Variables Explained

The `.env` file contains:

| Variable | Purpose | Default |
|----------|---------|---------|
| `API_KEY` | Security key for API endpoints | `test-key-12345` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost/pushiq` |
| `ENVIRONMENT` | Deploy mode (development/production) | `development` |
| `PORT` | API server port | `8080` |

**Important:** In production, change `API_KEY` to a secure random value!

---

## 🛑 Stopping Everything

### Stop Backend
Press `Ctrl+C` in the backend terminal

### Stop Frontend
Press `Ctrl+C` in the frontend terminal

### Stop PostgreSQL
```bash
# Stop the container
docker-compose down

# Remove container and data
docker-compose down -v
```

---

## 🔧 Common Issues & Solutions

### Issue: `docker: command not found`
**Solution:** Install Docker Desktop from https://www.docker.com/products/docker-desktop

### Issue: `Database connection refused`
**Solution:** 
```bash
# Check if container is running
docker-compose ps

# If not running, start it
docker-compose up -d

# Check logs
docker-compose logs postgres
```

### Issue: `API_KEY is required`
**Solution:** Make sure you've exported the `.env` file:
```bash
cd backend
export $(cat .env | xargs)
./pushiq-api
```

### Issue: `Port 5432 already in use`
**Solution:** Either stop the other PostgreSQL, or change port in docker-compose.yml:
```yaml
ports:
  - "5433:5432"  # Use 5433 instead
```

Then update `DATABASE_URL` to use port 5433:
```
DATABASE_URL=postgres://postgres:postgres@localhost:5433/pushiq
```

### Issue: `Port 8080 already in use`
**Solution:** Change `PORT` in `.env` file:
```
PORT=3001
```

Then access backend at `http://localhost:3001`

### Issue: Frontend can't connect to backend
**Solution:** 
1. Make sure backend is running (`curl http://localhost:8080/health`)
2. Check API key in login matches `API_KEY` in `.env`
3. Check browser console for errors (F12)

---

## 📝 Development Workflow

### Making Backend Changes

1. Edit code in `backend/internal/`
2. Stop backend (`Ctrl+C`)
3. Rebuild: `go build -o pushiq-api ./cmd/pushiq-api`
4. Restart: `set -a; source .env; set +a; ./pushiq-api`

### Making Frontend Changes

1. Edit code in `frontend/src/`
2. Changes auto-reload in browser (Vite hot reload)

### Making Database Changes

1. Create new migration file: `backend/migrations/0004_your_migration.sql`
2. Stop backend
3. Stop PostgreSQL: `docker-compose down`
4. Start PostgreSQL: `docker-compose up -d`
5. Restart backend

---

## 🔍 Accessing Database Directly

If you have `psql` installed:

```bash
psql -h localhost -U postgres -d pushiq

# Inside psql Shell, try:
\dt                    # List tables
SELECT * FROM devices; # Query data
\q                     # Exit
```

Or use Docker to run psql:

```bash
docker exec -it pushiq-postgres psql -U postgres -d pushiq
```

---

## 📊 Monitoring Logs

### Backend Logs
```bash
# In backend terminal, you'll see logs automatically
```

### PostgreSQL Logs
```bash
docker-compose logs -f postgres
```

### All Services Logs
```bash
docker-compose logs -f
```

---

## 🚀 Next Steps

1. ✅ Start PostgreSQL: `docker-compose up -d`
2. ✅ Run backend: `set -a; source .env; set +a; ./pushiq-api`
3. ✅ Run frontend: `npm run dev`
4. ✅ Login and explore dashboard
5. ✅ Test API endpoints with curl

---

## 📚 More Documentation

- **[START_HERE.md](../START_HERE.md)** - Navigation guide
- **[README.md](../README.md)** - API reference
- **[PHASE_1_COMPLETION.md](../PHASE_1_COMPLETION.md)** - Technical details

---

**You're all set!** Run the commands above and you'll have PushIQ running locally in minutes.

Questions? Check the troubleshooting section above or review the main documentation files.
