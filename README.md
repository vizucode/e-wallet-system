# Domain Checker Service

A concurrent domain availability checker service built with Go, Gin, and PostgreSQL.

## 🚀 Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL (if running manually)

### ⚙️ Configuration (.env)

Create a `.env` file in the `configs/` directory. You can copy the example below:

```bash
# configs/.env

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=domain_checker
SSL_MODE=disable

# App Configuration
GIN_MODE=debug
```

---

## 🐳 Running with Docker Compose

The easiest way to run the application is using Docker Compose. This will start both the PostgreSQL database and the API service.

1. **Start the services:**

   ```bash
   docker-compose -f build/package/docker-compose.yml up --build
   ```

2. **Run Migrations:**
   
   Once the database container is running, you need to apply migrations. Run this command from your local machine:

   ```bash
   # Ensure configs/.env is set up (DB_HOST=localhost)
   go run cmd/migrate/main.go
   ```

   *Note: The application container might restart a few times until the database is ready.*

3. **Access the API:**
   The API will be available at `http://localhost:8080`.

---

## 📦 Running with Dockerfile (Manual Build)

If you prefer to build the Docker image manually:

1. **Build the image:**

   ```bash
   docker build -f build/package/Dockerfile -t domain-checker .
   ```

2. **Run the container:**
   
   Ensure you have a database running and accessible. Link it via environment variables.

   ```bash
   docker run -p 8080:8080 \
     -e DB_HOST=host.docker.internal \
     -e DB_PORT=5432 \
     -e DB_USER=postgres \
     -e DB_PASSWORD=postgres \
     -e DB_NAME=domain_checker \
     domain-checker
   ```

---

## 🛠 Running Manually (Local Development)

1. **Start PostgreSQL:**
   Ensure you have a PostgreSQL instance running locally on port 5432.

2. **Setup Configuration:**
   Create the `configs/.env` file as described in the Configuration section.

3. **Run Migrations:**
   Apply database schema changes:

   ```bash
   go run cmd/migrate/main.go
   ```

   To run with seeder (optional):
   ```bash
   go run cmd/migrate/main.go -seed
   ```

4. **Start the Application:**

   ```bash
   go run cmd/api/main.go
   ```

---

## 🧪 Running Unit Tests

To run the unit tests for the project:

```bash
go test -v ./...
```

or for a specific package:

```bash
go test -v ./internal/app/usecase/domain_checker/service/
```

---

## 📂 Project Structure

- `cmd/api`: Entry point for the API server.
- `cmd/migrate`: Entry point for database migrations.
- `configs`: Configuration files (.env).
- `internal/app`: Core application logic (Clean Architecture).
  - `dto`: Data Transfer Objects.
  - `routes`: Gin route definitions.
  - `usecase`: Business logic (Service) and Controllers.
- `build`: Docker and CI/CD related files.

## 📡 API Specification

### 1. Check Domains

Checks the availability and status of a list of domains.

- **URL:** `/domain-checker`
- **Method:** `POST`
- **Content-Type:** `application/json`

#### Request Body

```json
{
  "name": "Batch-1",
  "domains": [
    "google.com",
    "example.com",
    "github.com",
    "stackoverflow.com",
    "openai.com",
    "golang.org",
    "reddit.com",
    "amazon.com",
    "netflix.com",
    "microsoft.com",
    "apple.com",
    "facebook.com",
    "twitter.com",
    "linkedin.com",
    "instagram.com",
    "cloudflare.com",
    "bbc.com",
    "cnn.com",
    "yahoo.com",
    "wikipedia.org"
  ]
}
```

#### Example cURL

```bash
curl -X POST 'http://localhost:8080/domain-checker' \
  --header 'Content-Type: application/json' \
  --data '{
  "name": "Batch-1",
  "domains": [
    "google.com",
    "example.com",
    "github.com",
    "stackoverflow.com",
    "openai.com",
    "golang.org",
    "reddit.com",
    "amazon.com",
    "netflix.com",
    "microsoft.com",
    "apple.com",
    "facebook.com",
    "twitter.com",
    "linkedin.com",
    "instagram.com",
    "cloudflare.com",
    "bbc.com",
    "cnn.com",
    "yahoo.com",
    "wikipedia.org"
  ]
}'
```

#### Response (200 OK)

```json
{
  "message": "Success",
  "success": true
}
```


