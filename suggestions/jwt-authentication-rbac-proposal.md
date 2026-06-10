# JWT Authentication and Role-Based Access Proposal

## Summary

This proposal expands **Priority 1: Security & Authentication / JWT Authentication System** from `api-upgrade-recommendations.md`.

The API already has a `users` table with email, password hash, and role fields. The next authentication upgrade should introduce JWT login, authenticated API routes, and role-based authorization for sensitive actions.

Recommended role names:

| Role | Purpose | Notes |
|------|---------|-------|
| `admin` | Owner/admin user with all permissions | This is the "me" role. Can manage users, orders, incomes, expenses, agenda items, and order statuses. |
| `operator` | Normal app user who performs day-to-day work | Recommended replacement for `worker` or current `employee`. The word is clearer for someone operating the business workflow without owning system settings. |
| `guest` | Read-only user | Can view records but cannot create, update, delete, or modify statuses. |

`operator` is the recommended role name instead of `worker` because it sounds less generic, maps better to application usage, and keeps the permission model focused on what the user can do in the system.

---

## Current State

The current migration defines:

```sql
role TEXT NOT NULL DEFAULT 'employee' CHECK(role IN ('admin', 'employee'))
```

Recommended target:

```sql
role TEXT NOT NULL DEFAULT 'operator' CHECK(role IN ('admin', 'operator', 'guest'))
```

If existing local data can be reset, update the existing user migration directly. If deployed data must be preserved, add a new migration that maps `employee` to `operator` and rebuilds the SQLite table with the new role constraint.

---

## Authentication Scope

### Public routes

These should stay public:

```text
GET /ping
GET /dbping
POST /api/auth/login
POST /api/auth/refresh
```

Registration should not be public by default for this app. Prefer admin-created users:

```text
POST /api/users
```

Only `admin` can create users.

### Protected routes

All business routes under `/api` should require a valid JWT:

```text
/api/orders
/api/order-statuses
/api/incomes
/api/expenses
/api/expense-categories
/api/financial-percentages
/api/agenda
/api/users
```

The current `X-API-Key` guard can remain as a temporary deployment-level lock, but JWT should become the primary user-level authentication system.

---

## Permission Matrix

| Resource | Guest | Operator | Admin |
|----------|-------|----------|-------|
| Orders | Read | Create, read, update, delete | Full access |
| Order statuses | Read | Read only | Create, read, update, deactivate, reorder |
| Incomes/payments | Read | Create, read, update, delete | Full access |
| Expenses | Read | Create, read, update, delete | Full access |
| Expense categories | Read | Create, read, update, delete | Full access |
| Financial percentages | Read | Read only | Create, read, update, delete |
| Agenda items | Read | Create, read, update, delete | Full access |
| Users | Own profile only | Own profile only | Create, read, update, deactivate users |

Key rule: **only `admin` can edit order statuses**.

---

## JWT Claims

Use short-lived access tokens and longer-lived refresh tokens.

Recommended access token claims:

```json
{
  "sub": "1",
  "email": "owner@example.com",
  "role": "admin",
  "type": "access",
  "iat": 1710000000,
  "exp": 1710000900
}
```

Recommended token lifetimes:

| Token | Lifetime | Storage |
|-------|----------|---------|
| Access token | 15 minutes | Frontend memory or secure cookie |
| Refresh token | 7-30 days | HttpOnly secure cookie preferred |

For the first implementation, bearer tokens are acceptable:

```text
Authorization: Bearer <access_token>
```

---

## Suggested Endpoints

### Auth

```text
POST /api/auth/login
POST /api/auth/logout
POST /api/auth/refresh
GET  /api/auth/me
```

Avoid public registration:

```text
POST /api/auth/register
```

Use admin-managed users instead:

```text
GET    /api/users
POST   /api/users
GET    /api/users/:id
PUT    /api/users/:id
DELETE /api/users/:id
```

`DELETE /api/users/:id` should be implemented as deactivation if preserving audit history matters.

---

## Backend Implementation Plan

### 1. Update user role model

Create a role type in the users/auth package:

```go
type Role string

const (
    RoleAdmin    Role = "admin"
    RoleOperator Role = "operator"
    RoleGuest    Role = "guest"
)
```

### 2. Add password hashing

Use bcrypt:

```text
golang.org/x/crypto/bcrypt
```

Store only `password_hash`. Never store plaintext passwords.

### 3. Add auth service

Responsibilities:

- Validate login credentials
- Hash new user passwords
- Generate access tokens
- Generate and validate refresh tokens
- Return the current authenticated user

### 4. Add JWT middleware

Responsibilities:

- Read the `Authorization` header
- Validate the bearer token
- Reject expired or invalid tokens
- Attach authenticated user context to the Fiber request

### 5. Add role middleware

Example route protection:

```go
api.Use(auth.RequireAuth(jwtService))

statusGroup := api.Group("/order-statuses")
statusGroup.Get("/", statusHandler.List)
statusGroup.Post("/", auth.RequireRole(RoleAdmin), statusHandler.Create)
statusGroup.Put("/:id", auth.RequireRole(RoleAdmin), statusHandler.Update)
statusGroup.Delete("/:id", auth.RequireRole(RoleAdmin), statusHandler.Delete)
```

Use helper middleware for common cases:

```go
auth.RequireAnyRole(RoleAdmin, RoleOperator)
auth.RequireRole(RoleAdmin)
auth.RequireReadAccess()
```

---

## Frontend Impact

The frontend should add:

- Login screen
- Auth session store
- Automatic access token attachment
- Refresh token handling
- Logout action
- Role-aware navigation and disabled controls

Role-specific UI behavior:

| Role | UI behavior |
|------|-------------|
| `admin` | Show all settings, status editing, and user management |
| `operator` | Hide status editing and user management; allow daily workflow actions |
| `guest` | Read-only views; hide create/edit/delete actions |

Backend authorization must remain the source of truth. Frontend hiding is only for user experience.

---

## Security Notes

- Use `JWT_SECRET` from environment variables.
- Require a strong secret in production.
- Do not commit secrets.
- Return generic login errors such as `invalid email or password`.
- Hash passwords with bcrypt.
- Prefer HttpOnly secure cookies for refresh tokens in production.
- Keep `/ping` and `/dbping` unauthenticated for health checks.
- Add tests for role-denied actions, especially order status edits.

---

## Implementation Priority

1. Update roles to `admin`, `operator`, and `guest`.
2. Add login and JWT middleware.
3. Protect all `/api` business routes.
4. Add RBAC middleware.
5. Restrict order status mutations to `admin`.
6. Add admin-managed users.
7. Add frontend login and role-aware controls.
