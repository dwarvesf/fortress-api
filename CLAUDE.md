# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Essential Commands

- `make dev` - Run the application locally (starts on port 8080)
- `make test` - Run the full test suite with database setup
- `make build` - Build the Go application for current OS
- `make lint` - Run golangci-lint for code quality checks
- `make init` - Set up development environment (pull Docker images, initialize databases)

### Database Management

- `make migrate-up` / `make migrate-down` - Apply/revert database migrations
- `make migrate-new name=<migration_name>` - Create new migration file
- `make reset-db` - Reset local database (migrate down, up, then seed)
- `make seed-db` - Apply seed data to database
- `make migrate-test` - Apply migrations to test database

### Documentation & Code Generation

- `make gen-swagger` - Generate Swagger/OpenAPI documentation from code annotations

### Environment Setup

- Use `make shell` for devbox isolated environment
- Docker Compose provides PostgreSQL for local development and testing
- Environment variables configured via `.env` file

## Architecture Overview

### Core Structure

This is a Go web API using **layered architecture** with clear separation of concerns:

**Entry Point**: `cmd/server/main.go`

- Initializes config, logging, database, services, background worker
- Sets up HTTP router with graceful shutdown

**Request Flow**: `Routes → Controllers → Services → Stores → Database`

### Key Packages

- **`pkg/model/`** - GORM-mapped database entities with validation and helper methods
- **`pkg/store/`** - Repository pattern implementation for database access (one per model)
- **`pkg/service/`** - Business logic layer orchestrating store calls
- **`pkg/handler/`** - HTTP request handlers (parsing, validation, response serialization)
- **`pkg/controller/`** - Intermediary layer between handlers and services
- **`pkg/routes/`** - Gin router setup with middleware and endpoint definitions

### Web Framework & Middleware

- **Framework**: Gin (github.com/gin-gonic/gin)
- **Authentication**: JWT with conditional bypass in local environment
- **Authorization**: Permission-based middleware using role permissions
- **API Versioning**: Routes organized under `/api/v1` with public/private separation

### Database Layer

- **ORM**: GORM with PostgreSQL
- **Migrations**: sql-migrate with timestamped, reversible migration files
- **Patterns**: Soft deletes (`deleted_at`), UUID primary keys, repository pattern
- **Testing**: Isolated test database with transaction rollback per test

## Development Patterns

### Migration Management

- Migration files in `migrations/schemas/` with format: `YYYYMMDDHHMMSS-description.sql`
- Each migration has `-- +migrate Up` and `-- +migrate Down` sections
- Seed data in `migrations/seed/` with modular organization via `\ir` includes
- Always run `make migrate-up` and `make migrate-test` after pulling new code

### Testing Strategy

- **Pattern**: Table-driven tests with golden file comparison
- **Database**: Transaction-wrapped tests using `testhelper.TestWithTxDB()`
- **Mocking**: gomock for external dependencies
- **Test Data**: SQL fixtures in `testdata/` directories per handler
- **CI**: GitHub Actions runs `make ci` on all PRs

### API Development Flow

1. Define route in `pkg/routes/v1.go` with appropriate middleware
2. Implement handler interface in `pkg/handler/[domain]/interface.go`
3. Create handler implementation in `pkg/handler/[domain]/[domain].go` with Swagger annotations
4. Add controller logic in `pkg/controller/[domain]/`
5. Implement store methods in `pkg/store/[domain]/` following repository pattern
6. Create/update models in `pkg/model/` with GORM tags
7. Generate migration with `make migrate-new name=description`

### Code Organization Conventions

- Domain-driven package structure (employee, project, invoice, etc.)
- Interface definitions separate from implementations
- Extensive use of GORM preloading for related data
- Enum types with validation methods (`IsValid()`, `String()`)
- Helper methods on models for common operations

### Task Provider Abstraction (NocoDB Migration - 2025-01-19)

The application uses a **provider abstraction pattern** for task management, supporting both Basecamp (legacy) and NocoDB (current):

**Configuration**: Set `TASK_PROVIDER=nocodb` or `TASK_PROVIDER=basecamp` in `.env`

**Provider Interfaces**:
- `ExpenseProvider` - Fetch expenses for payroll calculation (`pkg/service/basecamp/basecamp.go`)
- `TaskIntegration` - Invoice and accounting task operations (`pkg/service/taskintegration/`)

**NocoDB Services**:
- `pkg/service/nocodb/expense.go` - Expense fetching from NocoDB API
- `pkg/service/nocodb/accounting_todo.go` - Accounting todo fetching for payroll
- `pkg/service/taskprovider/nocodb/provider.go` - Webhook handling and task operations

**Key Patterns**:
1. **No DB persistence on webhook approval** - Expenses/todos validated only
2. **Fetch during payroll calculation** - Pull from NocoDB API when calculating payroll
3. **Persist on payroll commit** - Write to DB only when payroll is committed
4. **Status updates** - Mark NocoDB records as "completed" after commit
5. **Metadata tracking** - Store `task_provider`, `task_ref`, `task_board`, `task_attachment_url` for cross-linking

**Rollback**: Basecamp code preserved - switch `TASK_PROVIDER=basecamp` if needed

## Configuration & Infrastructure

### Dependencies

- **Core**: Gin web framework, GORM ORM, PostgreSQL driver
- **Auth**: JWT, Vault integration for secrets
- **External**: Discord, SendGrid, GitHub API, Google Cloud services
- **Testing**: testfixtures, gomock, httptest

### Environment Management

- `dbconfig.yml` defines database connections for local/test/dev/prod
- Docker Compose provides isolated PostgreSQL instances
- Conditional middleware behavior based on environment (auth bypass in local)

### Documentation Standards

- **ADRs**: Architecture Decision Records in `docs/adr/` with status/context/decision/consequences
- **Specs**: Technical specifications in `docs/specs/` with data models and API definitions
- **Changelogs**: Feature-based change tracking in `docs/changelog/`
- **Swagger**: Auto-generated API documentation from handler annotations

### Code Review Process

- CODEOWNERS file requires approval from @huynguyenh @lmquang
- All changes go through GitHub PR process
- CI runs full test suite before merge

## Running Single Tests

For targeted testing of specific functionality:

```bash
# Run specific test function
go test -run TestFunctionName ./pkg/handler/dashboard

# Run tests for specific package
go test ./pkg/handler/dashboard

# Run with verbose output
go test -v ./pkg/handler/dashboard
```

## Common Debugging

### Database Issues

- Check connection with Docker: `docker-compose ps`
- Reset environment: `make reset-db`
- View logs: `docker-compose logs postgres`

### API Issues

- Check Swagger docs at `/swagger/index.html` when running locally
- Verify JWT token in local environment (or ensure auth is bypassed)
- Check middleware order in routes setup
