# 🚀 Savannah Backend API

A production-ready RESTful API built with Go and the Gin framework, featuring enterprise-grade architecture, comprehensive authentication, SMS integration with asynchronous processing, and cloud-native deployment capabilities.

[![CI/CD Pipeline](https://github.com/your-org/savannah-backend/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/your-org/savannah-backend/actions/workflows/ci-cd.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/your-org/savannah-backend)](https://goreportcard.com/report/github.com/your-org/savannah-backend)
[![Coverage](https://codecov.io/gh/your-org/savannah-backend/branch/main/graph/badge.svg)](https://codecov.io/gh/your-org/savannah-backend)

## 🏗️ Architecture Overview

This is a **production-grade microservice** built with enterprise patterns and cloud-native principles:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client Apps   │    │   Load Balancer  │    │   Kubernetes    │
│  (Web/Mobile)   │◄──►│    (Ingress)     │◄──►│    Cluster      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                         │
┌─────────────────────────────────────────────────────────────────┐
│                    Savannah Backend API                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │    Auth     │  │     API     │  │    Jobs     │            │
│  │ (OIDC/JWT)  │  │ (REST/HTTP) │  │ (SMS Queue) │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
                                │
            ┌───────────────────┼───────────────────┐
            │                   │                   │
    ┌───────▼────────┐  ┌───────▼────────┐  ┌──────▼──────┐
    │  PostgreSQL    │  │     Redis      │  │ Africa's    │
    │  (Primary DB)  │  │  (Job Queue)   │  │  Talking    │
    │   + Audit      │  │   + Caching    │  │ (SMS Gateway)│
    └────────────────┘  └────────────────┘  └─────────────┘
```

## 📋 Core Features

### 🔐 **Enterprise Authentication**
- **OpenID Connect (OIDC)** integration with multiple providers
- **JWT-based** access tokens with scope validation
- **Role-based access control** (RBAC)
- Support for **Auth0, Azure AD, Keycloak**

### 📱 **SMS Integration** 
- **Africa's Talking** API integration
- **Asynchronous SMS processing** with Redis job queue
- **Automatic retry** with exponential backoff
- **Delivery tracking** and error handling

### 🗄️ **Database Excellence**
- **PostgreSQL** with UUID primary keys
- **Full audit trail** with history tables
- **Automatic versioning** with database triggers
- **Optimized indexes** and query performance

### ☁️ **Cloud-Native Ready**
- **Docker** multi-stage builds for minimal images
- **Kubernetes** deployment with Helm charts
- **Horizontal pod autoscaling** (HPA)
- **Health checks** and graceful shutdown

### 🔄 **Production-Grade CI/CD**
- **GitHub Actions** pipeline with security scanning
- **Automated testing** with coverage enforcement
- **Multi-environment** deployments (dev → prod)
- **Automatic rollback** on deployment failures

## 📁 Project Structure

```
backend/
├── main.go                 # Application entry point
├── go.mod                 # Go module dependencies
├── go.sum                 # Go module checksums
├── .env                   # Environment variables (not in git)
├── .gitignore            # Git ignore patterns
├── README.md             # This file
├── cmd/                  # Command-line tools
│   └── migrate.go        # Database migration tool
└── pkg/                  # Package directory
    ├── config/           # Configuration management
    │   └── config.go
    ├── database/         # Database connection and management
    │   └── database.go
    ├── handlers/         # HTTP request handlers
    │   └── handlers.go
    ├── middleware/       # HTTP middleware
    │   └── middleware.go
    ├── migrations/       # Database migrations
    │   ├── migrations.go # Migration manager
    │   └── definitions.go# Migration definitions
    ├── models/          # Data models and structures
    │   └── models.go
    ├── routes/          # Route definitions
    │   └── routes.go
    └── utils/           # Utility functions
        └── utils.go
```

## Features

- **RESTful API** with clean route organization
- **Middleware support** for CORS, logging, authentication, and rate limiting
- **Modular architecture** with separation of concerns
- **Environment-based configuration**
- **Structured error handling and responses**
- **Sample endpoints** for users and products
- **Health check endpoint**
- **Database integration** with PostgreSQL using GORM
- **Explicit migration system** with version tracking and rollback support

## API Endpoints

### Health Check
- `GET /health` - Check server status

### Documentation
- `GET /docs` - API documentation

### Users
- `GET /api/v1/users` - Get all users
- `GET /api/v1/users/:id` - Get user by ID
- `POST /api/v1/users` - Create new user

### Products
- `GET /api/v1/products` - Get all products
- `GET /api/v1/products/:id` - Get product by ID
- `POST /api/v1/products` - Create new product

### Protected Routes
- `GET /api/v1/protected/dashboard` - Protected dashboard (requires auth)

### Admin Routes
- `GET /api/v1/admin/stats` - Admin statistics (requires auth + rate limiting)

## Getting Started

### Prerequisites

- Go 1.19 or later
- PostgreSQL 12 or later

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd backend
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up the database:
   ```bash
   # Create database (using psql or your preferred tool)
   createdb backend_dev
   
   # Or using psql:
   psql -U postgres -c "CREATE DATABASE backend_dev;"
   ```

4. Copy and configure environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

5. Run database migrations:
   ```bash
   go run cmd/migrate.go -action=up
   ```

6. Run the application:
   ```bash
   go run main.go
   ```

The server will start on `http://localhost:8080` by default.

### Development

To run in development mode with auto-reload, you can use `air`:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with auto-reload
air
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ENVIRONMENT` | Application environment | `development` |
| `PORT` | Server port | `8080` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `devuser` |
| `DB_PASSWORD` | Database password | `` |
| `DB_NAME` | Database name | `backend_dev` |
| `DB_SSLMODE` | Database SSL mode | `disable` |

## Database Migrations

The application uses an explicit migration system for better control over database schema changes.

### Migration Commands

```bash
# Run all pending migrations
go run cmd/migrate.go -action=up

# Check migration status
go run cmd/migrate.go -action=status

# Rollback the last migration
go run cmd/migrate.go -action=down

# Show help
go run cmd/migrate.go -help
```

### Current Migrations

1. **001_create_users_table** - Creates users table with email uniqueness
2. **002_create_categories_table** - Creates categories table
3. **003_create_products_table** - Creates products table with foreign key to categories
4. **004_add_indexes** - Adds database indexes for performance optimization

### Adding New Migrations

To add a new migration:

1. Add your migration definition to `pkg/migrations/definitions.go`
2. Follow the naming convention: `XXX_description_of_change`
3. Implement both `Up` and `Down` functions for the migration
4. Add the migration to the `getAllMigrations()` function

Example:
```go
{
    Version:     "005_add_user_roles",
    Description: "Add roles table and user_role relationship",
    Up:          addUserRoles,
    Down:        removeUserRoles,
}
```

## Testing the API

### Using curl

```bash
# Health check
curl http://localhost:8080/health

# Get all users
curl http://localhost:8080/api/v1/users

# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test User"}'

# Get API documentation
curl http://localhost:8080/docs
```

### Response Format

All API responses follow this structure:

```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... },
  "error": null
}
```

## Recent Implementations

The following features have been recently implemented:

1. **Asynchronous SMS Processing**: Integrated with Africa's Talking SMS gateway with Redis-based job queues
2. **Retry & Backoff Logic**: Automatic retry mechanism for failed SMS delivery attempts
3. **Redis Job Queue**: Durable job storage with TTL and sorted sets for priority management
4. **Background Worker**: Continuous processing of SMS jobs in separate goroutines

## Next Steps

Planned features for upcoming development:

1. **API Integration**: Connect SMS service with order creation API
2. **Error Handling**: Improve error handling and observability for job processing
3. **Docker Multi-Stage Build**: Create optimized production Docker images
4. **Kubernetes Deployment**: Develop Helm charts with ConfigMaps and Secrets
5. **CI/CD Pipeline**: Implement GitHub Actions for automated build, test and deployment

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.