# Go CMS Template

A scalable, maintainable Go REST API backend for content management systems.

## Features

- **Content Types**: Define dynamic content schemas
- **Content Posts**: Full CRUD with tags and media attachments
- **Media Management**: Track file metadata for images, videos, documents
- **Tags**: Categorize content with tags
- **Contact Submissions**: Handle contact form submissions
- **Settings**: Key-value configuration store

## Tech Stack

- **Go 1.22+**
- **Chi Router** - Lightweight, idiomatic HTTP router
- **pgx** - PostgreSQL driver with connection pooling
- **UUID** - Google's UUID library

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── database/            # Database connection
│   ├── handlers/            # HTTP request handlers
│   ├── middleware/          # HTTP middleware
│   ├── models/              # Data models and DTOs
│   ├── repository/          # Database operations
│   ├── response/            # API response helpers
│   └── router/              # Route definitions
├── .env.example             # Environment variables template
├── go.mod                   # Go modules
├── table.sql                # Database schema
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.22 or higher
- PostgreSQL 14 or higher

### Installation

1. Clone the repository:
```bash
git clone <repo-url>
cd go-cms-template
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Configure your `.env` file with your database credentials

4. Create the database and run migrations:
```bash
psql -U postgres -c "CREATE DATABASE cms_db"
psql -U postgres -d cms_db -f table.sql
```

5. Download dependencies:
```bash
go mod download
```

6. Run the server:
```bash
go run cmd/api/main.go
# or
make run
```

The server will start on `http://localhost:8080`

## API Endpoints

### Health Check
- `GET /health` - Check API health

### Content Types
- `GET /api/v1/content-types` - List content types
- `POST /api/v1/content-types` - Create content type
- `GET /api/v1/content-types/:id` - Get content type by ID
- `GET /api/v1/content-types/slug/:slug` - Get content type by slug
- `PUT /api/v1/content-types/:id` - Update content type
- `DELETE /api/v1/content-types/:id` - Delete content type

### Posts
- `GET /api/v1/posts` - List posts (with filters)
- `POST /api/v1/posts` - Create post
- `GET /api/v1/posts/:id` - Get post by ID
- `GET /api/v1/posts/slug/:slug` - Get post by slug
- `PUT /api/v1/posts/:id` - Update post
- `DELETE /api/v1/posts/:id` - Delete post
- `POST /api/v1/posts/:id/media` - Attach media to post
- `DELETE /api/v1/posts/:id/media/:mediaId` - Detach media from post

### Media
- `GET /api/v1/media` - List media
- `POST /api/v1/media` - Create media record
- `GET /api/v1/media/:id` - Get media by ID
- `PUT /api/v1/media/:id` - Update media
- `DELETE /api/v1/media/:id` - Delete media

### Tags
- `GET /api/v1/tags` - List tags
- `POST /api/v1/tags` - Create tag
- `GET /api/v1/tags/:id` - Get tag by ID
- `GET /api/v1/tags/slug/:slug` - Get tag by slug
- `PUT /api/v1/tags/:id` - Update tag
- `DELETE /api/v1/tags/:id` - Delete tag

### Contacts
- `GET /api/v1/contacts` - List contact submissions
- `POST /api/v1/contacts` - Create contact submission
- `GET /api/v1/contacts/:id` - Get contact by ID
- `GET /api/v1/contacts/unread-count` - Get unread count
- `PUT /api/v1/contacts/:id` - Update contact status
- `DELETE /api/v1/contacts/:id` - Delete contact

### Settings
- `GET /api/v1/settings` - List settings
- `POST /api/v1/settings` - Create setting
- `POST /api/v1/settings/upsert` - Create or update setting
- `POST /api/v1/settings/bulk` - Get multiple settings
- `GET /api/v1/settings/:key` - Get setting by key
- `PUT /api/v1/settings/:key` - Update setting
- `DELETE /api/v1/settings/:key` - Delete setting

## Query Parameters

### Pagination
All list endpoints support pagination:
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 20, max: 100)
- `sort_by` - Field to sort by
- `sort_dir` - Sort direction (`asc` or `desc`)

### Filtering
- **Posts**: `content_type_id`, `author_id`, `status`, `search`
- **Media**: `file_type`, `search`
- **Contacts**: `status`, `email`
- **Content Types**: `is_active`

## Response Format

All responses follow a consistent JSON structure:

```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

Error responses:
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found",
    "details": { ... }
  }
}
```

## Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 204 | No Content |
| 400 | Bad Request |
| 404 | Not Found |
| 409 | Conflict |
| 422 | Validation Error |
| 500 | Internal Server Error |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Server host | `0.0.0.0` |
| `SERVER_PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | - |
| `DATABASE_MAX_CONNS` | Max DB connections | `25` |
| `DATABASE_MIN_CONNS` | Min DB connections | `5` |
| `CORS_ALLOWED_ORIGINS` | Comma-separated origins | `http://localhost:3000` |
| `APP_ENV` | Environment (development/production) | `development` |

## Make Commands

```bash
make run            # Run the server
make build          # Build binary
make test           # Run tests
make test-coverage  # Run tests with coverage
make lint           # Run linter
make fmt            # Format code
make tidy           # Tidy modules
make clean          # Clean build artifacts
```

## License

MIT