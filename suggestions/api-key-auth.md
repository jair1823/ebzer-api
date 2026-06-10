# API Key Authentication

Simple API key guard to protect all `/api` routes from public access.  
No auth library needed — reads a secret from an environment variable.

---

## Risk without this

Since the API is publicly exposed on Railway with no authentication, anyone with the URL can:
- Read all orders, incomes, and statuses (`GET`)
- Create, modify, or delete any record (`POST`, `PUT`, `DELETE`)

---

## Backend — `cmd/server/main.go`

### 1. Allow `X-API-Key` in CORS headers

```go
app.Use(cors.New(cors.Config{
    AllowOrigins: "*",
    AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-API-Key",
}))
```

### 2. Add the guard on the `/api` group

Place this right after `api := app.Group("/api")`:

```go
// API key guard — set API_KEY env var to enable; skipped if unset (dev mode)
apiKey := os.Getenv("API_KEY")
if apiKey != "" {
    api.Use(func(c *fiber.Ctx) error {
        if c.Get("X-API-Key") != apiKey {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "unauthorized",
            })
        }
        return c.Next()
    })
}
```

> `/ping` and `/dbping` are intentionally left unprotected — Railway needs them for health checks.

---

## Frontend — `ebzer-web/src/services/api.ts`

```ts
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:3000/api";
const API_KEY = import.meta.env.VITE_API_KEY || "";

const authHeaders = (): HeadersInit => ({
  "Content-Type": "application/json",
  ...(API_KEY ? { "X-API-Key": API_KEY } : {}),
});
```

Replace all hardcoded `"Content-Type": "application/json"` with `...authHeaders()`, and add the header to GET and DELETE calls as well.

---

## Railway environment variables

| Service      | Variable          | Value                              |
|--------------|-------------------|------------------------------------|
| `ebzer-api`  | `API_KEY`         | strong random string (see below)   |
| `ebzer-web`  | `VITE_API_KEY`    | same value as above                |

Generate a secure key:
```bash
openssl rand -hex 32
```

> **Never commit the key to git.** Set it only in Railway's Variables UI.

---

## Behavior

| `API_KEY` env set? | Request has correct header? | Result  |
|--------------------|-----------------------------|---------|
| No                 | —                           | ✅ allowed (dev mode) |
| Yes                | Yes                         | ✅ allowed |
| Yes                | No / missing                | ❌ 401 Unauthorized |
