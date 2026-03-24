# User Service — REST API

Base URL: `http://localhost:8080`
All request and response bodies are JSON (`Content-Type: application/json`).

---

## Authentication

Protected endpoints require a valid access token in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

Access tokens are short-lived (15 min). Obtain them via [Register](#post-v1users) or [Login](#post-v1authlogin).

---

## Error responses

All errors follow this shape:

```json
{
  "error": "user not found",
  "message": "",
  "code": 404
}
```

| HTTP status | Meaning |
|-------------|---------|
| 400 | Invalid request body or validation failure |
| 401 | Missing, invalid, or expired token / wrong credentials |
| 403 | Account blocked or not yet active |
| 404 | Resource not found |
| 409 | Conflict (e.g. email already registered) |
| 500 | Internal server error |

---

## Endpoints

### POST /v1/users

Register a new user account.

**Auth:** none

**Request**

```json
{
  "email": "alice@example.com",
  "phone": "1234567890",
  "password": "S3cret!pw",
  "first_name": "Alice",
  "last_name": "Smith"
}
```

| Field | Type | Rules |
|-------|------|-------|
| `email` | string | valid email format |
| `phone` | string | 10–15 digits |
| `password` | string | 8–128 chars, must contain uppercase, lowercase, and a digit |
| `first_name` | string | required |
| `last_name` | string | required |

**Response `201 Created`**

```json
{
  "user_id": "01929b3e-f1a2-7c3d-8e4f-5a6b7c8d9e0f",
  "email": "alice@example.com",
  "access_token": "<jwt>",
  "refresh_token": "<jwt>",
  "created_at": "2026-03-24T10:00:00Z"
}
```

**Errors**

| Status | Trigger |
|--------|---------|
| 400 | Validation failure (invalid email, weak password, …) |
| 409 | Email already registered |

---

### POST /v1/auth/login

Authenticate with email and password.

**Auth:** none

**Request**

```json
{
  "email": "alice@example.com",
  "password": "S3cret!pw"
}
```

**Response `200 OK`**

```json
{
  "user_id": "01929b3e-f1a2-7c3d-8e4f-5a6b7c8d9e0f",
  "access_token": "<jwt>",
  "refresh_token": "<jwt>"
}
```

**Errors**

| Status | Trigger |
|--------|---------|
| 401 | Wrong email or password |
| 403 | Account blocked or not active |

---

### GET /v1/users/me

Return the profile of the currently authenticated user.

**Auth:** bearer token

**Response `200 OK`**

```json
{
  "id": "01929b3e-f1a2-7c3d-8e4f-5a6b7c8d9e0f",
  "email": "alice@example.com",
  "phone": "1234567890",
  "first_name": "Alice",
  "last_name": "Smith",
  "status": "active",
  "kyc_status": "none",
  "created_at": "2026-03-24T10:00:00Z",
  "updated_at": "2026-03-24T10:00:00Z"
}
```

**Errors**

| Status | Trigger |
|--------|---------|
| 401 | Missing or invalid token |
| 404 | User no longer exists |

---

### GET /v1/users/:id

Return a user profile by ID.

**Auth:** bearer token

**Path params**

| Param | Description |
|-------|-------------|
| `id` | User UUID |

**Response `200 OK`** — same shape as [GET /v1/users/me](#get-v1usersme)

**Errors**

| Status | Trigger |
|--------|---------|
| 401 | Missing or invalid token |
| 404 | User not found |

---

### GET /v1/users/email/:email

Return a user profile by email address.

**Auth:** bearer token

**Path params**

| Param | Description |
|-------|-------------|
| `email` | URL-encoded email address |

**Response `200 OK`** — same shape as [GET /v1/users/me](#get-v1usersme)

**Errors**

| Status | Trigger |
|--------|---------|
| 401 | Missing or invalid token |
| 404 | User not found |

---

### GET /v1/users/:id/validate

Check whether a user is active and eligible to transact.

**Auth:** bearer token

**Path params**

| Param | Description |
|-------|-------------|
| `id` | User UUID |

**Response `200 OK`**

```json
{
  "valid": true,
  "status": "active",
  "kyc_status": "verified"
}
```

When the user is not found, returns `valid: false` with empty status fields (no error status code).

---

### POST /v1/users/:id/change-password

Change a user's password. Requires the current password for verification.

**Auth:** bearer token

**Path params**

| Param | Description |
|-------|-------------|
| `id` | User UUID |

**Request**

```json
{
  "old_password": "S3cret!pw",
  "new_password": "NewP@ss99"
}
```

**Response `204 No Content`**

**Errors**

| Status | Trigger |
|--------|---------|
| 400 | New password fails validation rules |
| 401 | Missing or invalid token / wrong old password |
| 404 | User not found |

---

## User object reference

```json
{
  "id":         "string (UUID)",
  "email":      "string",
  "phone":      "string",
  "first_name": "string",
  "last_name":  "string",
  "status":     "pending | active | blocked | deleted",
  "kyc_status": "none | pending | verified | rejected",
  "created_at": "string (RFC3339)",
  "updated_at": "string (RFC3339)"
}
```
