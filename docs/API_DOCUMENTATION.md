# Blue Ledger API Documentation

**Version:** 1.0.0  
**Base URL:** `http://localhost:8081/v1` (local) | `https://api.blue-ledger.com/v1` (production)  
**Authentication:** JWT Bearer Token

---

## Table of Contents

- [Authentication](#authentication)
- [Core Endpoints](#core-endpoints)
- [Members API](#members-api)
- [Authentication API](#authentication-api)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Code Examples](#code-examples)

---

## Authentication

### Overview
All protected endpoints require a valid JWT token in the `Authorization` header.

```bash
Authorization: Bearer <access_token>
```

### Token Structure
JWT tokens contain the following claims:
```json
{
  "iss": "blue-ledger",
  "sub": "user-id",
  "user_id": "8f0c8f9e-f8c2-4dab-80a2-913b883315f8",
  "chapter_id": "8f22b21e-624a-413a-a6d9-67756233ca7f",
  "member_id": "optional-member-id",
  "role": "sysadmin|admin|member",
  "email": "user@example.org",
  "is_sysadmin": false,
  "exp": 1234567890,
  "iat": 1234567890
}
```

### Token Expiration
- **Access Token:** 15 minutes (900 seconds)
- **Refresh Token:** 7 days

---

## Core Endpoints

### Health Check
**No authentication required**

```
GET /healthz
```

**Response:**
```json
{
  "status": "ok"
}
```

---

## Members API

### List Members

```
GET /members
```

**Query Parameters:**
| Parameter | Type | Description |
|---|---|---|
| `page` | integer | Page number (default: 1) |
| `limit` | integer | Items per page (default: 50, max: 100) |
| `sort` | string | Sort field (default: "created_at") |
| `order` | string | Sort order: "asc" or "desc" (default: "desc") |

**Example:**
```bash
curl -X GET "http://localhost:8081/v1/members?page=1&limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": [
    {
      "id": "9a54536d-e67b-402e-8e5d-6f6cc954d80d",
      "user_id": "8f0c8f9e-f8c2-4dab-80a2-913b883315f8",
      "chapter_id": "8f22b21e-624a-413a-a6d9-67756233ca7f",
      "name": "Readiness Test",
      "email": "readiness@test.com",
      "first_name": "Readiness",
      "last_name": "Test",
      "city": "Bangalore",
      "job_title": "QA Engineer",
      "xp": 0,
      "level": 0,
      "role": "member",
      "is_active": true,
      "created_at": "2026-03-30T00:00:00Z",
      "updated_at": "2026-03-30T00:01:00Z",
      "deleted_at": null
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 26,
    "pages": 2
  }
}
```

---

### Get Member by ID

```
GET /members/:id
```

**Parameters:**
| Parameter | Type | Required | Description |
|---|---|---|---|
| `id` | UUID | Yes | Member ID |

**Example:**
```bash
curl -X GET "http://localhost:8081/v1/members/9a54536d-e67b-402e-8e5d-6f6cc954d80d" \
  -H "Authorization: Bearer $TOKEN"
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "id": "9a54536d-e67b-402e-8e5d-6f6cc954d80d",
    "user_id": "8f0c8f9e-f8c2-4dab-80a2-913b883315f8",
    "chapter_id": "8f22b21e-624a-413a-a6d9-67756233ca7f",
    "name": "Readiness Test",
    "email": "readiness@test.com",
    "first_name": "Readiness",
    "last_name": "Test",
    "city": "Bangalore",
    "job_title": "QA Engineer",
    "xp": 0,
    "level": 0,
    "role": "member",
    "is_active": true,
    "created_at": "2026-03-30T00:00:00Z",
    "updated_at": "2026-03-30T00:01:00Z",
    "deleted_at": null
  }
}
```

**Error Response (404 Not Found):**
```json
{
  "status": "error",
  "message": "member not found"
}
```

---

### Create Member

```
POST /members
```

**Authorization:** Admin or Sysadmin required

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "chapter_id": "8f22b21e-624a-413a-a6d9-67756233ca7f",
  "city": "San Francisco",
  "job_title": "Software Engineer",
  "role": "member"
}
```

**Example:**
```bash
curl -X POST http://localhost:8081/v1/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "chapter_id": "8f22b21e-624a-413a-a6d9-67756233ca7f"
  }'
```

**Response (201 Created):**
```json
{
  "status": "success",
  "data": {
    "id": "new-member-uuid",
    "user_id": "associated-user-uuid",
    "chapter_id": "8f22b21e-624a-413a-a6d9-67756233ca7f",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "city": "San Francisco",
    "job_title": "Software Engineer",
    "xp": 0,
    "level": 0,
    "role": "member",
    "is_active": true,
    "created_at": "2026-03-30T12:00:00Z",
    "updated_at": "2026-03-30T12:00:00Z",
    "deleted_at": null
  }
}
```

**Error Response (400 Bad Request):**
```json
{
  "status": "error",
  "message": "invalid request: missing required field 'email'"
}
```

---

### Update Member

```
PUT /members/:id
```

**Authorization:** Admin, Chair, or self-update

**Request Body (all fields optional):**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "city": "San Francisco",
  "job_title": "Senior Engineer",
  "xp": 100
}
```

**Example:**
```bash
curl -X PUT http://localhost:8081/v1/members/9a54536d-e67b-402e-8e5d-6f6cc954d80d \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "city": "San Francisco",
    "job_title": "QA Engineer"
  }'
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "id": "9a54536d-e67b-402e-8e5d-6f6cc954d80d",
    "city": "San Francisco",
    "job_title": "QA Engineer",
    "updated_at": "2026-03-30T12:01:00Z"
  }
}
```

**Error Response (403 Forbidden):**
```json
{
  "status": "error",
  "message": "unauthorized: insufficient permissions"
}
```

---

### Delete Member (Soft Delete)

```
DELETE /members/:id
```

**Authorization:** Admin required

**Example:**
```bash
curl -X DELETE http://localhost:8081/v1/members/9a54536d-e67b-402e-8e5d-6f6cc954d80d \
  -H "Authorization: Bearer $TOKEN"
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "member deleted"
}
```

---

### Reactivate Member

```
POST /members/:id/reactivate
```

**Authorization:** Admin required

**Example:**
```bash
curl -X POST http://localhost:8081/v1/members/9a54536d-e67b-402e-8e5d-6f6cc954d80d/reactivate \
  -H "Authorization: Bearer $TOKEN"
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "member reactivated",
  "data": {
    "id": "9a54536d-e67b-402e-8e5d-6f6cc954d80d",
    "is_active": true,
    "deleted_at": null
  }
}
```

---

## Authentication API

### Login

```
POST /auth/login
```

**Request Body:**
```json
{
  "email": "admin@tausigmasigma.org",
  "password": "BlueLedger2026!"
}
```

**Example:**
```bash
curl -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@tausigmasigma.org",
    "password": "BlueLedger2026!"
  }'
```

**Response (200 OK):**
```json
{
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "24f3afaefe2621e7608a0f648ece5654742be70a...",
    "expires_in": 900,
    "user": {
      "id": "8f0c8f9e-f8c2-4dab-80a2-913b883315f8",
      "email": "admin@tausigmasigma.org",
      "first_name": "Admin",
      "last_name": "User",
      "chapter_id": "8f22b21e-624a-413a-a6d9-67756233ca7f",
      "role": "sysadmin",
      "is_sysadmin": true
    }
  }
}
```

**Error Response (401 Unauthorized):**
```json
{
  "status": "error",
  "message": "invalid credentials"
}
```

---

## Error Handling

### Error Response Format

All errors follow a consistent format:

```json
{
  "status": "error",
  "message": "Human-readable error message",
  "error_code": "ERROR_CODE",
  "details": {}
}
```

### HTTP Status Codes

| Code | Meaning |
|---|---|
| 200 | OK |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 500 | Internal Server Error |
| 503 | Service Unavailable |

### Common Errors

**400 Bad Request**
```json
{
  "status": "error",
  "message": "invalid request: field validation failed",
  "details": {
    "email": "invalid email format"
  }
}
```

**401 Unauthorized**
```json
{
  "status": "error",
  "message": "missing or invalid authorization token"
}
```

**403 Forbidden**
```json
{
  "status": "error",
  "message": "insufficient permissions for this action"
}
```

**404 Not Found**
```json
{
  "status": "error",
  "message": "resource not found"
}
```

---

## Rate Limiting

Rate limits are applied per user/token:

- **General API:** 1,000 requests per minute
- **Auth endpoints:** 10 requests per minute per IP
- **Search endpoints:** 100 requests per minute

Rate limit information is provided in response headers:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1234567890
```

When rate limit exceeded:
```json
{
  "status": "error",
  "message": "rate limit exceeded",
  "retry_after": 60
}
```

---

## Code Examples

### Python

```python
import requests
import json

BASE_URL = "http://localhost:8081/v1"

# Login
login_response = requests.post(
    f"{BASE_URL}/auth/login",
    json={
        "email": "admin@tausigmasigma.org",
        "password": "BlueLedger2026!"
    }
)
token = login_response.json()["data"]["access_token"]
headers = {"Authorization": f"Bearer {token}"}

# Get members
members = requests.get(f"{BASE_URL}/members", headers=headers)
print(json.dumps(members.json(), indent=2))

# Create member
new_member = requests.post(
    f"{BASE_URL}/members",
    headers=headers,
    json={
        "first_name": "John",
        "last_name": "Doe",
        "email": "john@example.com",
        "chapter_id": "8f22b21e-624a-413a-a6d9-67756233ca7f"
    }
)
print(json.dumps(new_member.json(), indent=2))
```

### JavaScript

```javascript
const BASE_URL = "http://localhost:8081/v1";

async function login() {
  const response = await fetch(`${BASE_URL}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      email: "admin@tausigmasigma.org",
      password: "BlueLedger2026!"
    })
  });
  return response.json();
}

async function getMembers(token) {
  const response = await fetch(`${BASE_URL}/members`, {
    headers: { "Authorization": `Bearer ${token}` }
  });
  return response.json();
}

async function createMember(token, memberData) {
  const response = await fetch(`${BASE_URL}/members`, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${token}`,
      "Content-Type": "application/json"
    },
    body: JSON.stringify(memberData)
  });
  return response.json();
}

// Usage
const loginData = await login();
const token = loginData.data.access_token;
const members = await getMembers(token);
console.log(members);
```

### cURL

```bash
#!/bin/bash

BASE_URL="http://localhost:8081/v1"

# Login
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@tausigmasigma.org",
    "password": "BlueLedger2026!"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.access_token')

# Get members
curl -s -X GET "$BASE_URL/members" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Create member
curl -s -X POST "$BASE_URL/members" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "chapter_id": "8f22b21e-624a-413a-a6d9-67756233ca7f"
  }' | jq '.'
```

---

## Support

For API issues or questions:
- **GitHub Issues:** https://github.com/Ces1231/blue-ledger/issues
- **Documentation:** https://github.com/Ces1231/blue-ledger#readme
