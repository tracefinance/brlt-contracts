# Vault0 Go Server with React SPA

A secure, modular, and deployable foundation for a Go server application that serves a React Single Page Application (SPA), sets up a SQLite database, and applies initial database migrations.

## Features

- Go server built with Gin web framework
- SQLite database integration with libSQL driver
- Database migrations using golang-migrate
- Precompiled React SPA serving with client-side routing support
- Secure configuration management
- Modular codebase with clean separation of concerns

## Prerequisites

- Go 1.16 or higher
- SQLite installed on your system

## Project Structure

```
go-server/
├── cmd/                # Application entrypoints
│   └── server/         # Server command
│       └── main.go     # Main server application
├── internal/           # Private application code
│   ├── api/            # API handlers and routing
│   ├── config/         # Configuration management
│   └── db/             # Database access and management
├── migrations/         # Database migration files
│   ├── 000001_create_users_table.up.sql
│   └── 000001_create_users_table.down.sql
└── ui/                 # Frontend related files
    └── dist/           # Compiled React SPA
```

## Getting Started

### Running the Server

1. Clone this repository
2. Navigate to the project directory
3. Run the server:

```bash
cd go-server
go run cmd/server/main.go
```

The server will start on port 8080 by default (or the port specified in the `SERVER_PORT` environment variable).

### Configuration

The server can be configured using environment variables:

- `APP_BASE_DIR`: Base directory for relative paths (defaults to current working directory)
- `DB_PATH`: Path to SQLite database file (defaults to `vault0.db` in the base directory)
- `SERVER_PORT`: Port for the server to listen on (defaults to `8080`)
- `UI_PATH`: Path to the compiled React UI files (defaults to `ui/dist` in the base directory)
- `MIGRATIONS_PATH`: Path to migration files (defaults to `migrations` in the base directory)

Example:

```bash
SERVER_PORT=3000 DB_PATH=/path/to/database.db go run cmd/server/main.go
```

## Database Migrations

The server automatically applies migrations at startup. Migrations are stored in the `migrations` directory.

If you want to manually apply migrations, you can use the golang-migrate CLI tool:

```bash
# Install golang-migrate CLI
go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Apply migrations
migrate -database "sqlite3://vault0.db" -path ./migrations up
```

To create new migrations:

```bash
migrate create -ext sql -dir ./migrations -seq migration_name
```

## API Endpoints

- `GET /api/health`: Health check endpoint that returns `{"status": "ok"}`

## Frontend

The React SPA is served from the `ui/dist` directory. The server is configured to handle client-side routing by serving `index.html` for all non-API routes.

## Development

### Adding New API Endpoints

Add new API endpoints in the `internal/api/server.go` file:

```go
apiGroup.GET("/your-endpoint", yourHandlerFunc)
```

### Adding New Database Migrations

Create new migration files in the `migrations` directory following the naming convention: `{version}_{name}.{up|down}.sql`.

For example:
- `000002_add_email_to_users.up.sql`
- `000002_add_email_to_users.down.sql`

## Security Notes

- Sensitive configuration is stored in environment variables
- The database connection is securely managed
- The `private_key` field in the users table is intended for future encryption use

## Future Enhancements

- User authentication
- API documentation with Swagger
- Enhanced security features
- HTTPS support
