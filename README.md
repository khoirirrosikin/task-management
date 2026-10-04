# Task Management API 🚀

A production-grade, collaborative Task Management REST API built with **Go** and **PostgreSQL**. Designed with a **Modular Monolith** architecture (Package by Feature) to demonstrate clean code, idiomatic Go practices, role-based access control (RBAC), and high-performance database interactions.

---

## 📌 Project Overview

This project is a streamlined yet robust collaboration platform inspired by tools like Trello and Asana. It is engineered to showcase real-world backend engineering practices, moving beyond basic CRUD tutorials into multi-tenancy, granular project authorization, relational data modeling, and asynchronous task processing.

### Key Highlights
- **Modular Monolith Architecture**: High cohesion and loose coupling by organizing packages by domain/feature (`auth`, `project`, `task`).
- **Idiomatic & Type-Safe Database Layer**: Powered by [`sqlc`](https://sqlc.dev/) and [`pgx/v5`](https://github.com/jackc/pgx) with zero ORM overhead and 100% type-safe compile-time SQL queries.
- **Database Migrations**: Version-controlled with [`goose`](https://github.com/pressly/goose).
- **Authentication & Security**: Stateless JWT (HS256) authentication and bcrypt password hashing.
- **Reliable Server Lifecycle**: Native Graceful Shutdown handling OS signals (`SIGINT`, `SIGTERM`).
- **Comprehensive Unit Testing**: High-coverage unit tests with interface mocking via [`testify`](https://github.com/stretchr/testify), executed without requiring a live database.

---

## 🛠️ Tech Stack

| Component | Technology | Rationale |
| :--- | :--- | :--- |
| **Language** | Go 1.22+ | High concurrency, type safety, and minimal memory footprint |
| **Web Framework** | [Gin Gonic](https://github.com/gin-gonic/gin) | High performance, robust middleware ecosystem, idiomatic HTTP handling |
| **Database** | PostgreSQL 16 | Relational integrity, ACID compliance, native UUID support |
| **Driver & Pooling**| `pgx/v5` (`pgxpool`) | Industry-standard PostgreSQL driver with connection pooling |
| **Query Generator**| `sqlc` | Type-safe Go code generated from raw SQL queries |
| **Migration Tool** | `goose` | Transactional, single-file (`Up`/`Down`) SQL migrations |
| **Auth & Crypto** | `golang-jwt/jwt/v5`, `bcrypt` | Secure authentication and password hashing |
| **Validation** | `go-playground/validator/v10` | Declarative request payload validation |
| **Testing** | `testing`, `testify` | Standard Go testing with assertion and mock support |
| **Containerization**| Docker & Docker Compose | Containerized local PostgreSQL development environment |

---

## 🏗️ Architecture & Project Structure

The project follows a **Modular Monolith** structure where each feature is self-contained inside `internal/<module>`:

```text
task-management/
├── cmd/
│   └── api/
│       └── main.go                 # Application entrypoint & dependency injection
├── internal/
│   ├── auth/                       # Authentication module (Register, Login, JWT)
│   │   ├── dto.go                  # Request/Response data transfer objects
│   │   ├── handler.go              # Gin HTTP handlers
│   │   ├── handler_test.go         # Unit tests for HTTP handlers
│   │   ├── repository.go           # Database operations interface & implementation
│   │   ├── repository_test.go      # Unit tests for database repository with pgxmock
│   │   ├── service.go              # Business logic & JWT signing
│   │   └── service_test.go         # Unit tests with mock repository
│   ├── project/                    # Project management module
│   │   ├── dto.go                  # Request/Response data transfer objects
│   │   ├── handler.go              # Gin HTTP handlers
│   │   ├── handler_test.go         # Unit tests for HTTP handlers
│   │   ├── repository.go           # Database operations interface & implementation
│   │   ├── repository_test.go      # Unit tests for database repository with pgxmock
│   │   ├── service.go              # Business logic & authorization checks
│   │   └── service_test.go         # Unit tests with mock repository
│   ├── task/                       # Task management module
│   │   ├── dto.go                  # Request/Response data transfer objects
│   │   ├── handler.go              # Gin HTTP handlers (Shallow nesting)
│   │   ├── handler_test.go         # Unit tests for HTTP handlers
│   │   ├── repository.go           # Database operations interface & implementation
│   │   ├── service.go              # Business logic & access verification
│   │   └── service_test.go         # Unit tests with mock repositories
│   ├── middleware/                 # Shared middlewares (Auth JWT, CORS, Logger)
│   │   ├── auth.go                 # JWT authentication middleware
│   │   └── auth_test.go            # Unit tests for JWT auth middleware
│   ├── response/                   # Standardized JSON response envelope & centralized error handling
│   │   ├── error.go                # Custom AppError and centralized HandleError
│   │   ├── response.go             # Success and Error response builders
│   │   └── validator.go            # Validation error formatter
│   └── database/                   # Database connection pool (pgxpool)
│       ├── postgres.go
│       └── db/                     # Auto-generated code by sqlc (DO NOT EDIT)
│           ├── db.go
│           ├── models.go
│           ├── users.sql.go
│           ├── projects.sql.go
│           └── tasks.sql.go
├── sql/
│   ├── migrations/                 # Versioned DDL migrations for goose
│   │   ├── 00001_create_users_table.sql
│   │   ├── 00002_create_projects_table.sql
│   │   └── 00003_create_tasks_table.sql
│   └── queries/                    # Raw SQL queries for sqlc
│       ├── users.sql
│       ├── projects.sql
│       └── tasks.sql
├── docker-compose.yml              # Local PostgreSQL container definition
├── sqlc.yaml                       # sqlc code generation configuration
├── go.mod
└── go.sum
```

---

## 🗺️ Roadmap & Features

### Phase 1 — Foundation (Current)
- [x] Docker & PostgreSQL 16 environment setup
- [x] Database migration system with `goose`
- [x] Type-safe query generation with `sqlc`
- [x] User Registration with bcrypt password hashing
- [x] User Login with JWT token generation
- [x] Standardized API JSON response envelope
- [x] Graceful shutdown handling
- [x] Unit test suite for Auth Service, Handler, & Repository (100% Mock-driven, 94%+ coverage)
- [x] Unit test suite for Project Service, Handler, & Repository (100% Mock-driven, 81%+ coverage)
- [x] JWT Authentication Middleware
- [x] Centralized AppError & Error Handling
- [x] CRUD Project (Create, List, Detail, Update, Delete)
- [x] CRUD Task (Title, Description, Status, Due Date, Priority)
- [x] Unit test suite for Task Service & Handler (100% Mock-driven, 84%+ coverage)
- [ ] Unit test suite for Task Repository with `pgxmock`

### Phase 2 — Collaboration & RBAC
- [ ] Project member invitation (via email/username)
- [ ] Project-level Role-Based Access Control (`owner` vs `member`)
- [ ] Task assignment to project members
- [ ] Task comments & discussions

### Phase 3 — Advanced Value (Portfolio Standout)
- [ ] Asynchronous task notification system using Go goroutines & channels
- [ ] Audit / Activity log (who modified what and when)
- [ ] Search & filtering for tasks (`status`, `priority`, `assignee`, `due_date`)
- [ ] Offset / Cursor pagination for list endpoints

---

## 🗄️ Database Schema

```mermaid
erDiagram
    USERS ||--o{ PROJECTS : "owns"
    USERS ||--o{ PROJECT_MEMBERS : "joins"
    PROJECTS ||--o{ PROJECT_MEMBERS : "contains"
    PROJECTS ||--o{ TASKS : "has"
    USERS ||--o{ TASKS : "assigned to"
    TASKS ||--o{ COMMENTS : "has"
    USERS ||--o{ COMMENTS : "authors"

    USERS {
        uuid id PK "gen_random_uuid()"
        varchar name
        varchar email UK
        varchar password_hash
        timestamptz created_at
        timestamptz updated_at
    }

    PROJECTS {
        uuid id PK "gen_random_uuid()"
        varchar name
        text description
        uuid owner_id FK
        timestamptz created_at
        timestamptz updated_at
    }

    PROJECT_MEMBERS {
        uuid id PK
        uuid project_id FK
        uuid user_id FK
        varchar role "owner | member"
        timestamptz joined_at
    }

    TASKS {
        uuid id PK
        uuid project_id FK
        varchar title
        text description
        varchar status "todo | in_progress | completed"
        varchar priority "low | medium | high | urgent"
        uuid assigned_to FK
        timestamptz due_date
        timestamptz created_at
        timestamptz updated_at
    }

    COMMENTS {
        uuid id PK
        uuid task_id FK
        uuid user_id FK
        text content
        timestamptz created_at
    }
```

---

## 🚀 Getting Started

### Prerequisites
- [Go](https://go.dev/dl/) 1.22 or later
- [Docker](https://www.docker.com/) & Docker Compose
- [Goose](https://github.com/pressly/goose) (`go install github.com/pressly/goose/v3/cmd/goose@latest`)
- [SQLC](https://sqlc.dev/) (`brew install sqlc` or `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

### 1. Clone & Setup Environment
```bash
git clone git@github.com:khoirirrosikin/task-management.git
cd task-management
```

Create a `.env` file in the root directory:
```env
DATABASE_URL=postgres://postgres:password@localhost:5432/task_management?sslmode=disable
JWT_SECRET=supersecretjwtkey_portfolio_12345
PORT=8080
GIN_MODE=debug
```

### 2. Start PostgreSQL
```bash
docker compose up -d
```

### 3. Run Database Migrations
```bash
goose -dir sql/migrations postgres "postgres://postgres:password@localhost:5432/task_management?sslmode=disable" up
```

### 4. Generate SQLC Code (Optional, if modifying queries)
```bash
sqlc generate
```

### 5. Run the Server
```bash
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`.

---

## 🧪 Running Tests

Run all unit tests across modules:
```bash
go test -v ./...
```

Run unit tests for specific modules:
```bash
go test -v ./internal/auth
go test -v ./internal/project
go test -v ./internal/task
```

Check test code coverage:
```bash
go test -cover ./internal/auth
go test -cover ./internal/project
go test -cover ./internal/task
```

---

## 📡 API Reference

### Health Check
- **Endpoint:** `GET /health`
- **Response:** `200 OK`
```json
{
  "status": "OK",
  "timestamp": "2026-09-19T22:33:58.476164+07:00"
}
```

### Register User
- **Endpoint:** `POST /api/v1/auth/register`
- **Request Body:**
```json
{
  "name": "Khoirir Rosikin",
  "email": "khoirir.rosikin@gmail.com",
  "password": "secretpassword123"
}
```
- **Response:** `201 Created`
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
      "name": "Khoirir Rosikin",
      "email": "khoirir.rosikin@gmail.com",
      "created_at": "2026-09-19T22:37:40.507475+07:00"
    }
  }
}
```

### Login User
- **Endpoint:** `POST /api/v1/auth/login`
- **Request Body:**
```json
{
  "email": "khoirir.rosikin@gmail.com",
  "password": "secretpassword123"
}
```
- **Response:** `200 OK`
```json
{
  "success": true,
  "message": "User logged in successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
      "name": "Khoirir Rosikin",
      "email": "khoirir.rosikin@gmail.com",
      "created_at": "2026-09-19T22:37:40.507475+07:00"
    }
  }
}
```

### Create Project
- **Endpoint:** `POST /api/v1/projects`
- **Headers:** `Authorization: Bearer <token>`
- **Request Body:**
```json
{
  "name": "Task Management API",
  "description": "Backend API built with Go and PostgreSQL"
}
```
- **Response:** `201 Created`
```json
{
  "success": true,
  "message": "Project created successfully",
  "data": {
    "id": "c1f7a012-6874-4ec5-b286-d24eef6ff2aa",
    "name": "Task Management API",
    "description": "Backend API built with Go and PostgreSQL",
    "owner_id": "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
    "created_at": "2026-09-22T08:00:00Z",
    "updated_at": "2026-09-22T08:00:00Z"
  }
}
```

### List Projects
- **Endpoint:** `GET /api/v1/projects`
- **Headers:** `Authorization: Bearer <token>`
- **Response:** `200 OK`
```json
{
  "success": true,
  "message": "Projects retrieved successfully",
  "data": [
    {
      "id": "c1f7a012-6874-4ec5-b286-d24eef6ff2aa",
      "name": "Task Management API",
      "description": "Backend API built with Go and PostgreSQL",
      "owner_id": "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
      "created_at": "2026-09-22T08:00:00Z",
      "updated_at": "2026-09-22T08:00:00Z"
    }
  ]
}
```

### Get Project Details
- **Endpoint:** `GET /api/v1/projects/:id`
- **Headers:** `Authorization: Bearer <token>`
- **Response:** `200 OK`
```json
{
  "success": true,
  "message": "Project retrieved successfully",
  "data": {
    "id": "c1f7a012-6874-4ec5-b286-d24eef6ff2aa",
    "name": "Task Management API",
    "description": "Backend API built with Go and PostgreSQL",
    "owner_id": "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
    "created_at": "2026-09-22T08:00:00Z",
    "updated_at": "2026-09-22T08:00:00Z"
  }
}
```

### Update Project
- **Endpoint:** `PUT /api/v1/projects/:id`
- **Headers:** `Authorization: Bearer <token>`
- **Request Body:**
```json
{
  "name": "Updated Project Name",
  "description": "Updated project description"
}
```
- **Response:** `200 OK`
```json
{
  "success": true,
  "message": "Project updated successfully",
  "data": {
    "id": "c1f7a012-6874-4ec5-b286-d24eef6ff2aa",
    "name": "Updated Project Name",
    "description": "Updated project description",
    "owner_id": "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
    "created_at": "2026-09-22T08:00:00Z",
    "updated_at": "2026-09-22T08:30:00Z"
  }
}
```

### Delete Project
- **Endpoint:** `DELETE /api/v1/projects/:id`
- **Headers:** `Authorization: Bearer <token>`
- **Response:** `200 OK`
```json
{
  "success": true,
  "message": "Project deleted successfully"
}
```

### Create Task
- **Endpoint:** `POST /api/v1/projects/:projectId/tasks`
- **Headers:** `Authorization: Bearer <token>`
- **Request Body:**
```json
{
  "title": "Setup database migrations",
  "description": "Create goose migration files for tasks table",
  "status": "todo",
  "priority": "high",
  "assigned_to": "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
  "due_date": "2026-10-10T12:00:00Z"
}
```
- **Response:** `201 Created`
```json
{
  "success": true,
  "message": "Task created successfully",
  "data": {
    "id": "e4b52c01-7fa1-42cb-b72e-c1214a1a679e",
    "project_id": "c1f7a012-6874-4ec5-b286-d24eef6ff2aa",
    "title": "Setup database migrations",
    "description": "Create goose migration files for tasks table",
    "status": "todo",
    "priority": "high",
    "assigned_to": "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
    "due_date": "2026-10-10T12:00:00Z",
    "created_at": "2026-10-03T12:00:00Z",
    "updated_at": "2026-10-03T12:00:00Z"
  }
}
```

### List Tasks by Project
- **Endpoint:** `GET /api/v1/projects/:projectId/tasks`
- **Headers:** `Authorization: Bearer <token>`
- **Response:** `200 OK`
```json
{
  "success": true,
  "message": "Tasks retrieved successfully",
  "data": [
    {
      "id": "e4b52c01-7fa1-42cb-b72e-c1214a1a679e",
      "project_id": "c1f7a012-6874-4ec5-b286-d24eef6ff2aa",
      "title": "Setup database migrations",
      "description": "Create goose migration files for tasks table",
      "status": "todo",
      "priority": "high",
      "assigned_to": "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
      "due_date": "2026-10-10T12:00:00Z",
      "created_at": "2026-10-03T12:00:00Z",
      "updated_at": "2026-10-03T12:00:00Z"
    }
  ]
}
```

### Get Task Details
- **Endpoint:** `GET /api/v1/tasks/:id`
- **Headers:** `Authorization: Bearer <token>`
- **Response:** `200 OK`
```json
{
  "success": true,
  "message": "Task retrieved successfully",
  "data": {
    "id": "e4b52c01-7fa1-42cb-b72e-c1214a1a679e",
    "project_id": "c1f7a012-6874-4ec5-b286-d24eef6ff2aa",
    "title": "Setup database migrations",
    "description": "Create goose migration files for tasks table",
    "status": "todo",
    "priority": "high",
    "assigned_to": "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
    "due_date": "2026-10-10T12:00:00Z",
    "created_at": "2026-10-03T12:00:00Z",
    "updated_at": "2026-10-03T12:00:00Z"
  }
}
```

### Update Task
- **Endpoint:** `PUT /api/v1/tasks/:id`
- **Headers:** `Authorization: Bearer <token>`
- **Request Body:**
```json
{
  "status": "in_progress",
  "priority": "medium"
}
```
- **Response:** `200 OK`
```json
{
  "success": true,
  "message": "Task updated successfully",
  "data": {
    "id": "e4b52c01-7fa1-42cb-b72e-c1214a1a679e",
    "project_id": "c1f7a012-6874-4ec5-b286-d24eef6ff2aa",
    "title": "Setup database migrations",
    "description": "Create goose migration files for tasks table",
    "status": "in_progress",
    "priority": "medium",
    "assigned_to": "26d367a0-c964-4958-a9c5-9c2ad422ff8e",
    "due_date": "2026-10-10T12:00:00Z",
    "created_at": "2026-10-03T12:00:00Z",
    "updated_at": "2026-10-03T12:30:00Z"
  }
}
```

### Delete Task
- **Endpoint:** `DELETE /api/v1/tasks/:id`
- **Headers:** `Authorization: Bearer <token>`
- **Response:** `200 OK`
```json
{
  "success": true,
  "message": "Task deleted successfully"
}
```

---

## 📄 License
This project is open-source and available under the [MIT License](LICENSE).
