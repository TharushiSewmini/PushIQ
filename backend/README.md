# PushIQ Backend Service

Go backend service for push notification delivery via FCM and APNs.

## Project Structure

- `cmd/pushiq-api` — Main API server entrypoint
- `internal/api` — HTTP handlers and middleware
- `internal/config` — Configuration loader
- `internal/delivery` — Delivery engine and provider abstractions
- `internal/model` — Domain models
- `internal/repository` — Database repository layer
- `internal/util` — Utilities (logger, etc.)
- `migrations` — SQL schema migrations

## Local Setup

1. Create a `.env` file based on `.env.example` with your config
2. Start PostgreSQL and apply migrations:
   ```sql
   psql -U postgres -h localhost -d pushiq < migrations/0001_init.sql
   ```
3. Run the server:
   ```bash
   go mod tidy
   go run ./cmd/pushiq-api/main.go
   ```

The API will be available at `http://localhost:8080`.

## API Endpoints

### Register Device
```
POST /api/v1/devices/register
X-Api-Key: <your-api-key>

{
  "user_id": "user123",
  "platform": "android",
  "token": "device-token-xyz",
  "app_version": "1.0.0",
  "locale": "en-US"
}
```

### Send Notification
```
POST /api/v1/notifications/send
X-Api-Key: <your-api-key>

{
  "title": "Hello",
  "body": "Test notification",
  "device_id": "uuid-or-token-value",
  "platform": "android",
  "data": {"key": "value"},
  "priority": "high"
}
```
