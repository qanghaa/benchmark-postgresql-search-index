# Log Performance Testing Project

A high-performance logging system built with **Go**, **PostgreSQL**, **sqlc**, and **goose** to test and verify PostgreSQL search performance with GIN indexes at scale.

## 🎯 Project Goals

- Test PostgreSQL full-text search performance with **GIN indexes** on JSONB data
- Verify query performance at different data scales (1K, 10K, 100K, 1M, 10M records)
- Optimize bulk insert operations using **sqlc's CopyFrom** feature
- Provide a clean UI to interact with and test the database

## 🏗️ Architecture

### Tech Stack

- **Backend**: Go 1.21+
- **Database**: PostgreSQL 15+
- **SQL Generation**: sqlc (type-safe SQL with CopyFrom optimization)
- **Migrations**: goose (embedded migrations)
- **Web Framework**: Gin
- **Frontend**: HTML, CSS, JavaScript (Bootstrap 5)

### Database Schema

```sql
CREATE TABLE logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    domain VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    content JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Indexes

1. **B-tree indexes** for `user_id` and `domain` (exact match queries)
2. **GIN index** on `content` JSONB field (JSON queries)
3. **GIN index** for full-text search on `content::text`
4. **BRIN index** on `created_at` (time-series optimization)
5. **Composite B-tree index** on `(user_id, domain, created_at)`

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make (optional, for convenience)

### Installation

1. **Clone the repository**
   ```bash
   cd /home/quanghaa/projects/golang/log
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Install required tools** (sqlc and goose)
   ```bash
   make install-tools
   ```

4. **Start PostgreSQL**
   ```bash
   make docker-up
   ```

5. **Run the application**
   ```bash
   make run
   ```

6. **Access the application**
   - Web UI: http://localhost:8080
   - Swagger API: http://localhost:8080/swagger/index.html

## 📊 Features

### Data Initialization

Generate test data with configurable parameters:

- **Record counts**: 1K, 10K, 100K, 1M, 10M
- **Content sizes**:
  - **Small**: ~10 fields
  - **Medium**: 50-100 fields
  - **Large**: 300-500 fields (with nested objects and arrays)

### Search & Filter

- Filter by `user_id` (UUID)
- Filter by `domain` (exact match)
- Filter by `created_at` (date range)
- **Full-text search** on JSONB content (uses GIN index)

### Performance Metrics

The API returns query performance metrics:
- Query execution time
- Records per second during bulk insert
- Total records count

## 🔧 Development

### Project Structure

```
.
├── cmd/                    # Command-line tools (if needed)
├── config/                 # Configuration management
├── database/               # Database initialization and migrations
├── handlers/               # HTTP handlers (using sqlc)
├── internal/
│   └── db/                # Generated sqlc code
├── migrations/            # Goose migration files
├── models/                # Data models
├── sqlc/                  # SQL queries for sqlc
├── utils/                 # Utility functions
├── web/
│   ├── static/           # JavaScript, CSS
│   └── templates/        # HTML templates
├── docker-compose.yml    # PostgreSQL setup
├── Makefile              # Development commands
├── sqlc.yaml             # sqlc configuration
└── main.go               # Application entry point
```

### Available Make Commands

```bash
make help              # Show all available commands
make deps              # Download and update dependencies
make install-tools     # Install sqlc and goose
make sqlc              # Generate sqlc code
make run               # Run the application
make build             # Build binary
make docker-up         # Start PostgreSQL
make docker-down       # Stop PostgreSQL
make docker-logs       # View PostgreSQL logs
make dev               # Start PostgreSQL + App
make clean             # Clean build artifacts
make test              # Run tests
```

### Regenerate sqlc Code

After modifying `sqlc/queries.sql`:

```bash
make sqlc
```

### Database Migrations

Migrations are automatically run on application startup. To manually manage migrations:

```bash
# Check migration status
export DATABASE_URL="postgres://loguser:logpassword@localhost:5432/logdb?sslmode=disable"
make migrate-status

# Rollback last migration
make migrate-down
```

## 📡 API Endpoints

### Initialize Data
```http
POST /api/initialize
Content-Type: application/json

{
  "record_count": 100000,
  "content_size": "medium"
}
```

### List Logs with Filters
```http
GET /api/logs?user_id=<uuid>&domain=example.com&created_at=2024-01-01&created_at_to=2024-12-31&content_like=search+terms&page=1&limit=50
```

### Truncate Database
```http
DELETE /api/truncate
```

## 🧪 Testing Performance

### Test Scenario 1: Bulk Insert Performance

1. Navigate to http://localhost:8080
2. Select record count (e.g., 1M)
3. Select content size (e.g., medium)
4. Click "Initialize"
5. Observe:
   - Total duration
   - Records per second
   - Progress updates

### Test Scenario 2: Full-Text Search Performance

1. Initialize data (e.g., 1M records with medium content)
2. Use the "Content Search" filter
3. Enter search terms (e.g., "premium gold")
4. Click "Apply Filters"
5. Check the `query_duration` in the response

### Test Scenario 3: Index Performance Comparison

To verify GIN index performance:

1. Run a full-text search query and note the duration
2. Drop the GIN index:
   ```sql
   DROP INDEX idx_logs_content_fts;
   ```
3. Run the same query again
4. Compare execution times
5. Recreate the index:
   ```sql
   CREATE INDEX idx_logs_content_fts ON logs USING GIN (to_tsvector('english', content::text));
   ```

## 🔍 Performance Optimization

### Bulk Insert Optimization

This project uses **sqlc's CopyFrom** feature, which leverages PostgreSQL's `COPY` protocol for maximum insert performance:

```go
// Generated by sqlc
func (q *Queries) BulkInsertLogs(ctx context.Context, arg []BulkInsertLogsParams) (int64, error) {
    return q.db.CopyFrom(ctx, []string{"logs"}, 
        []string{"user_id", "domain", "action", "content", "created_at"}, 
        &iteratorForBulkInsertLogs{rows: arg})
}
```

**Performance**: Typically achieves 50,000-100,000+ inserts/second depending on hardware and content size.

### Query Optimization

- **GIN indexes** enable fast full-text search on JSONB content
- **BRIN indexes** optimize time-series queries on `created_at`
- **Composite indexes** improve multi-column filter performance

## 📝 Environment Variables

Create a `.env` file (copy from `.env.example`):

```bash
DATABASE_URL=postgres://loguser:logpassword@localhost:5432/logdb?sslmode=disable
PORT=8080
LOG_LEVEL=info
```

## 🐳 Docker Deployment

Build and run with Docker Compose:

```bash
docker-compose up --build
```

This will start:
- PostgreSQL on port 5432
- Application on port 8080

## 📚 Additional Resources

- [sqlc Documentation](https://docs.sqlc.dev/)
- [goose Documentation](https://github.com/pressly/goose)
- [PostgreSQL GIN Indexes](https://www.postgresql.org/docs/current/gin.html)
- [PostgreSQL Full-Text Search](https://www.postgresql.org/docs/current/textsearch.html)

## 🤝 Contributing

This is a performance testing project. Feel free to:
- Add new test scenarios
- Optimize queries
- Improve the UI
- Add benchmarks

## 📄 License

MIT License

## 🎓 Learning Objectives

By working with this project, you'll learn:

1. How to use **sqlc** for type-safe SQL in Go
2. How to optimize bulk inserts with **CopyFrom**
3. How to use **goose** for database migrations
4. PostgreSQL indexing strategies (B-tree, GIN, BRIN)
5. Full-text search implementation
6. Performance testing at scale

---

**Happy Testing! 🚀**