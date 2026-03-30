# Blue Ledger Architecture Document

**Version:** 1.0.0  
**Last Updated:** March 30, 2026  
**Status:** Production Ready

---

## Table of Contents

- [System Overview](#system-overview)
- [Architecture Diagram](#architecture-diagram)
- [Technology Stack](#technology-stack)
- [Backend Architecture](#backend-architecture)
- [Frontend Architecture](#frontend-architecture)
- [Database Schema](#database-schema)
- [API Design](#api-design)
- [Authentication & Authorization](#authentication--authorization)
- [Data Flow](#data-flow)
- [Deployment Architecture](#deployment-architecture)
- [Scalability](#scalability)

---

## System Overview

Blue Ledger is a full-stack engagement platform for Phi Beta Sigma chapters. The system tracks member participation, XP accumulation, avatar progression, and gamified engagement metrics.

### Core Components

1. **Backend API** - Go (Echo framework) RESTful API
2. **Frontend** - React single-page application
3. **Database** - PostgreSQL 16 with Row Level Security
4. **Cache** - Redis for sessions and real-time data
5. **Infrastructure** - Docker Compose for local, Fly.io for cloud

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        End Users                             │
│                    (Chapter Members)                         │
└──────────────────────────┬──────────────────────────────────┘
                           │
                    ┌──────▼──────┐
                    │   Nginx     │
                    │ (Reverse    │
                    │  Proxy)     │
                    └──┬──────┬───┘
                       │      │
       ┌───────────────┘      └────────────────┐
       │                                       │
   ┌───▼─────────┐                    ┌───────▼────┐
   │ React Web   │                    │ Go API     │
   │   App       │                    │ (Echo)     │
   │  (SPA)      │                    │            │
   └────┬────────┘                    └───┬────┬───┘
        │                                 │    │
        │ HTTP/REST                       │    │ WebSocket
        │                          ┌──────┘    │ (Optional)
        │                          │           │
        └──────────────────────────┼───────────┘
                                   │
                         ┌─────────▼────────┐
                         │  Router/Handler  │
                         │   Middleware     │
                         ├──────────────────┤
                         │  • Auth          │
                         │  • Logging       │
                         │  • Error Handle  │
                         │  • CORS          │
                         └────────┬─────────┘
                                  │
                    ┌─────────────┼─────────────┐
                    │             │             │
            ┌───────▼─────┐  ┌────▼────┐  ┌────▼────┐
            │   Service   │  │Validator│  │ Helper  │
            │   Layer     │  │  Layer  │  │ Utils   │
            │ (Business   │  │         │  │         │
            │  Logic)     │  └─────────┘  └─────────┘
            └───────┬─────┘
                    │
            ┌───────▼─────────────┐
            │  Repository Layer   │
            │  (Data Access)      │
            └───────┬─────────────┘
                    │
        ┌───────────┼────────────┐
        │           │            │
    ┌───▼─┐   ┌────▼────┐  ┌───▼─┐
    │ SQL │   │ Caching │  │Auth │
    │Query│   │ (Redis) │  │JWT  │
    └─────┘   └─────────┘  └─────┘
        │
    ┌───▼──────────────────────────┐
    │   PostgreSQL Database (16)   │
    │                              │
    │  • users                     │
    │  • members                   │
    │  • engagement_log            │
    │  • badges & quests           │
    │  • events & rsvps            │
    │  • [20+ more tables]         │
    │                              │
    │  Row Level Security (RLS)    │
    └──────────────────────────────┘
```

---

## Technology Stack

### Backend
| Component | Technology | Version | Purpose |
|---|---|---|---|
| Language | Go | 1.25+ | Backend logic |
| Framework | Echo | 4.x | HTTP server |
| Database | PostgreSQL | 16 | Primary datastore |
| Cache | Redis | 7+ | Sessions, cache |
| ORM | sqlc | Latest | Type-safe SQL |
| Auth | JWT | RS256 | API authentication |
| Container | Docker | 20.10+ | Containerization |

### Frontend
| Component | Technology | Version | Purpose |
|---|---|---|---|
| Library | React | 18+ | UI framework |
| Build Tool | Vite | 4+ | Fast bundling |
| Language | TypeScript | 5+ | Type safety |
| Styling | Tailwind CSS | 3+ | Utility CSS |
| HTTP Client | Axios/Fetch | Latest | API calls |
| State | Context API | N/A | State management |

### Infrastructure
| Component | Technology | Version | Purpose |
|---|---|---|---|
| Orchestration | Docker Compose | 1.29+ | Local/staging |
| Cloud Platform | Fly.io | Latest | Production |
| Reverse Proxy | Nginx | 1.24+ | Web server |
| CI/CD | GitHub Actions | N/A | Automation |

---

## Backend Architecture

### Project Structure

```
blue-ledger-api/
├── cmd/
│   ├── server/
│   │   └── main.go              # Application entry point
│   ├── migrate/
│   │   └── main.go              # Database migrations
│   └── seed/
│       └── main.go              # Test data seeding
│
├── internal/                    # Private application code
│   ├── members/
│   │   ├── handler.go           # HTTP handlers
│   │   ├── service.go           # Business logic
│   │   ├── repository.go        # Data access
│   │   └── model.go             # Data structures
│   │
│   ├── auth/                    # Authentication
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── middleware.go
│   │
│   ├── badges/                  # Badge system
│   ├── events/                  # Event management
│   ├── notifications/           # Notification engine
│   ├── platform/                # Platform settings
│   └── [18+ more modules]
│
├── pkg/                         # Shared packages
│   ├── config/                  # Configuration
│   ├── db/                      # Database utilities
│   ├── logger/                  # Logging
│   ├── middleware/              # HTTP middleware
│   ├── errors/                  # Error handling
│   └── utils/                   # Helper functions
│
├── migrations/                  # Database migrations
│   ├── 001_users_and_chapters.up.sql
│   ├── 002_rls_policies.up.sql
│   └── [19+ more migrations]
│
├── docker-compose.yml          # Service orchestration
├── Dockerfile                  # Container image
├── Makefile                    # Build commands
├── go.mod                      # Go dependencies
└── go.sum
```

### Handler-Service-Repository Pattern

All modules follow this three-layer architecture:

```
┌────────────────────┐
│     Handler        │ HTTP layer - request/response
│  (HTTP Handlers)   │
└────────┬───────────┘
         │
         │ Call
         │
┌────────▼───────────┐
│     Service        │ Business logic - rules, validation
│  (Business Logic)  │
└────────┬───────────┘
         │
         │ Call
         │
┌────────▼───────────┐
│   Repository       │ Data access - SQL queries
│  (Data Access)     │
└────────────────────┘
```

### Example: Member Creation Flow

```go
// 1. Handler receives HTTP request
func (h *Handler) Create(c echo.Context) error {
    var req createRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(400, Error{Message: "Invalid input"})
    }
    
    // 2. Call service with business context
    member, err := h.service.Create(c.Request().Context(), req)
    if err != nil {
        return c.JSON(500, Error{Message: err.Error()})
    }
    
    // 3. Return response
    return c.JSON(201, SuccessResponse{Data: member})
}

// Service layer
func (s *Service) Create(ctx context.Context, req createRequest) (*Member, error) {
    // Validate business rules
    if !req.Email.IsValid() {
        return nil, ErrInvalidEmail
    }
    
    // Call repository to persist
    member, err := s.repo.Create(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("create member: %w", err)
    }
    
    return member, nil
}

// Repository layer
func (r *Repository) Create(ctx context.Context, req createRequest) (*Member, error) {
    // Execute SQL
    member := &Member{}
    err := r.db.QueryRow(ctx, `
        INSERT INTO members (user_id, chapter_id, first_name, last_name, email)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at
    `, req.UserID, req.ChapterID, req.FirstName, req.LastName, req.Email).
        Scan(&member.ID, &member.CreatedAt)
    
    return member, err
}
```

---

## Frontend Architecture

### React Component Structure

```
src/
├── components/
│   ├── Layout/
│   │   ├── Header.tsx
│   │   ├── Sidebar.tsx
│   │   └── Footer.tsx
│   │
│   ├── Members/
│   │   ├── MemberList.tsx
│   │   ├── MemberCard.tsx
│   │   ├── MemberForm.tsx
│   │   └── MemberDetail.tsx
│   │
│   ├── Dashboard/
│   │   ├── Dashboard.tsx
│   │   ├── XPChart.tsx
│   │   └── Leaderboard.tsx
│   │
│   └── Common/
│       ├── Button.tsx
│       ├── Modal.tsx
│       ├── Loading.tsx
│       └── Error.tsx
│
├── pages/
│   ├── Home.tsx
│   ├── Members.tsx
│   ├── Profile.tsx
│   └── Admin.tsx
│
├── hooks/
│   ├── useAuth.ts
│   ├── useMembers.ts
│   ├── useAPI.ts
│   └── useLocalStorage.ts
│
├── services/
│   ├── api.ts               # API client
│   ├── auth.ts              # Authentication
│   ├── members.ts           # Member API calls
│   └── storage.ts           # Local storage
│
├── context/
│   ├── AuthContext.tsx      # Auth state
│   ├── MembersContext.tsx   # Members state
│   └── NotificationContext.tsx
│
├── types/
│   ├── api.ts
│   ├── models.ts
│   └── errors.ts
│
├── utils/
│   ├── validators.ts
│   ├── formatters.ts
│   └── helpers.ts
│
├── styles/
│   └── index.css
│
└── App.tsx                   # Root component
```

### State Management Pattern

```
┌──────────────────────┐
│  React Context API   │
│                      │
│ AuthContext          │
│ ├─ user              │
│ ├─ token             │
│ ├─ login()           │
│ └─ logout()          │
│                      │
│ MembersContext       │
│ ├─ members[]         │
│ ├─ loading           │
│ ├─ fetchMembers()    │
│ └─ createMember()    │
└──────────────────────┘
         │
         │ useContext()
         │
   ┌─────▼─────┐
   │ Components│
   │           │
   │ ├─ Header │
   │ ├─ List   │
   │ └─ Form   │
   └───────────┘
```

---

## Database Schema

### Core Tables

**users**
- id (UUID, PK)
- email (VARCHAR, UNQ)
- password_hash (VARCHAR)
- role (ENUM: sysadmin, admin, member)
- chapter_id (UUID, FK)
- created_at, updated_at, deleted_at

**members**
- id (UUID, PK)
- user_id (UUID, FK)
- chapter_id (UUID, FK)
- first_name, last_name (VARCHAR)
- name (VARCHAR - concat of first + last)
- email (VARCHAR)
- city, job_title (VARCHAR)
- xp (INTEGER)
- level (INTEGER)
- is_active (BOOLEAN)
- created_at, updated_at, deleted_at

**engagement_log**
- id (UUID, PK)
- member_id (UUID, FK)
- activity_type (ENUM)
- xp_earned (INTEGER)
- description (TEXT)
- created_at

**badges**
- id (UUID, PK)
- chapter_id (UUID, FK)
- name, description (VARCHAR, TEXT)
- xp_required (INTEGER)
- icon_url (VARCHAR)

**member_badges**
- member_id (UUID, FK)
- badge_id (UUID, FK)
- earned_at (TIMESTAMP)

### Row Level Security (RLS)

All tables have RLS enabled:

```sql
-- Members can only see their own chapter's data
CREATE POLICY "members_chapter_isolation" ON members
    USING (chapter_id = current_setting('user.chapter_id'))
    WITH CHECK (chapter_id = current_setting('user.chapter_id'));

-- Admins see all chapter data
-- Sysadmins see everything
```

---

## API Design

### RESTful Principles

| Resource | GET | POST | PUT | DELETE |
|---|---|---|---|---|
| `/members` | List | Create | - | - |
| `/members/:id` | Get | - | Update | Delete |
| `/members/:id/reactivate` | - | Restore | - | - |
| `/events` | List | Create | - | - |
| `/events/:id` | Get | - | Update | Delete |

### Request/Response Format

**Request:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com"
}
```

**Success Response:**
```json
{
  "status": "success",
  "data": { /* resource */ },
  "timestamp": "2026-03-30T12:00:00Z"
}
```

**Error Response:**
```json
{
  "status": "error",
  "message": "Human-readable message",
  "error_code": "SPECIFIC_ERROR",
  "timestamp": "2026-03-30T12:00:00Z"
}
```

---

## Authentication & Authorization

### JWT Token Flow

```
1. User Login (POST /auth/login)
        │
        ▼
┌───────────────────┐
│ Validate email    │
│ & password        │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│ Generate JWT      │
│ RS256 signed      │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│ Return token      │
│ + expiry          │
└─────────┬─────────┘
          │
          ▼
   Client stores
   in localStorage
        │
        ▼
┌───────────────────┐
│ Include in        │
│ Authorization     │
│ header            │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│ Middleware        │
│ validates JWT     │
└─────────┬─────────┘
          │
    ┌─────▼─────┐
    │   Valid?  │
    └─────┬─────┘
      Yes │ No
        │ │
      ┌─▼─┴──────┐
      │           │
      ▼           ▼
   Process    401 Error
   Request
```

### Authorization Levels

```
┌──────────────────────────────────────┐
│          Authorization               │
├──────────────────────────────────────┤
│ Sysadmin (is_sysadmin = true)       │
│ ├─ User CRUD                        │
│ ├─ All chapter operations           │
│ ├─ System configuration             │
│ └─ Audit log access                 │
│                                      │
│ Admin (role = admin)                │
│ ├─ Member management                │
│ ├─ Event management                 │
│ ├─ Financial reports                │
│ └─ Chapter settings                 │
│                                      │
│ Chair (role = chair)                │
│ ├─ QR scanner                       │
│ ├─ Check-in verification            │
│ └─ Event attendance                 │
│                                      │
│ Member (role = member)              │
│ ├─ Own profile edit                 │
│ ├─ View public leaderboard          │
│ ├─ RSVP to events                   │
│ └─ Log service hours                │
└──────────────────────────────────────┘
```

---

## Data Flow

### Member Creation Workflow

```
1. Frontend
   └─ MemberForm.tsx calls API

2. API Handler
   ├─ Validate request
   ├─ Check authorization
   └─ Call service

3. Service
   ├─ Validate business rules
   ├─ Check duplicate email
   └─ Call repository

4. Repository
   ├─ Create user record
   ├─ Generate display_id
   ├─ Create member record
   └─ Return full member

5. Backend Response
   └─ Return 201 Created + member data

6. Frontend
   ├─ Update local state
   ├─ Show success message
   └─ Redirect to member detail
```

### XP Accumulation Flow

```
1. Event occurs
   (attendance, service, quiz pass, etc)
   │
   ├─ Chair scans QR
   ├─ Service logged
   └─ Quiz submitted

2. Handler receives event
   └─ Create engagement_log entry

3. Service calculates XP
   ├─ Look up activity XP value
   ├─ Check multipliers (e.g., officer role)
   └─ Calculate total XP

4. Repository updates
   ├─ Insert engagement_log
   ├─ UPDATE members SET xp = xp + earned
   ├─ Check for level up
   └─ Check for badge unlocks

5. Notifications
   ├─ Send badge unlock notification
   ├─ Update leaderboard cache
   └─ Real-time update to connected clients

6. Frontend updates
   ├─ Show XP gain animation
   ├─ Update avatar if level up
   └─ Refresh leaderboard
```

---

## Deployment Architecture

### Local Development

```
Developer Machine
│
├─ Docker Daemon
│  │
│  ├─ blue-ledger-api:latest
│  │  ├─ Go application
│  │  ├─ Port 8081
│  │  └─ Linked to postgres:5432
│  │
│  ├─ blue-ledger-web:latest
│  │  ├─ Nginx + React
│  │  ├─ Port 3001
│  │  └─ Proxies API calls to :8081
│  │
│  ├─ postgres:16
│  │  ├─ Port 5432
│  │  └─ Volume: ./postgres_data
│  │
│  └─ redis:7
│     └─ Port 6379
│
└─ localhost:3001 ◄─ Browser
```

### Production on Fly.io

```
Fly.io Platform
│
├─ blue-ledger-api
│  ├─ 2+ regions (replicas)
│  ├─ Load balanced
│  ├─ Health checks
│  ├─ Auto-scaling
│  └─ Logs aggregation
│
├─ blue-ledger-web
│  ├─ CDN edge caching
│  ├─ 2+ regions
│  ├─ Static asset optimization
│  └─ Automatic HTTPS
│
└─ PostgreSQL (managed)
   └─ Automated backups
```

---

## Scalability

### Horizontal Scaling

```
Load Balancer (Fly.io)
│
├─ API Instance 1
├─ API Instance 2
├─ API Instance 3
└─ API Instance N

All connected to:
└─ Shared PostgreSQL
   └─ With read replicas
```

### Caching Strategy

```
Request Flow with Cache
│
├─ Check Redis cache
│  ├─ If hit → return (fast!)
│  └─ If miss → continue
│
├─ Query database
│
├─ Store in Redis
│  └─ With TTL (time-to-live)
│
└─ Return to client
```

### Database Optimization

1. **Indexing**: Key columns indexed for fast queries
2. **Connection Pooling**: Reuse connections
3. **Row Level Security**: Efficient data filtering
4. **Query Optimization**: EXPLAIN ANALYZE on slow queries
5. **Read Replicas**: Distribute read load

---

## Monitoring & Observability

### Logging Levels

```
DEBUG   ─ Development debugging
INFO    ─ Normal operations
WARN    ─ Warning conditions
ERROR   ─ Error conditions
FATAL   ─ Fatal conditions
```

### Key Metrics

- API response time (p50, p95, p99)
- Error rate (4xx, 5xx)
- Database query time
- Cache hit rate
- Container memory/CPU usage
- Active connections

---

## Security Considerations

1. **JWT RS256** - Asymmetric signing
2. **Row Level Security** - Database-level isolation
3. **Password Hashing** - bcrypt with salt
4. **HTTPS/TLS** - All traffic encrypted
5. **CORS** - Controlled cross-origin access
6. **SQL Injection Prevention** - Parameterized queries
7. **XSS Protection** - Content Security Policy
8. **Rate Limiting** - Prevent abuse

---

## Support

For architecture questions:
- **GitHub Issues:** https://github.com/Ces1231/blue-ledger/issues
- **Documentation:** https://github.com/Ces1231/blue-ledger#readme
