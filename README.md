# E-Wallet Service

A multi-currency e-wallet backend service built with Go, Gin, and PostgreSQL. Supports wallet creation, top-up, payment, transfer, suspension, and balance queries with transactional safety, decimal precision, and idempotency.

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
DB_NAME=db_ewallet
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
   docker build -f build/package/Dockerfile -t e-wallet-system .
   ```

2. **Run the container:**
   
   Ensure you have a database running and accessible. Link it via environment variables.

   ```bash
   docker run -p 8080:8080 \
     -e DB_HOST=host.docker.internal \
     -e DB_PORT=5432 \
     -e DB_USER=postgres \
     -e DB_PASSWORD=postgres \
     -e DB_NAME=db_ewallet \
     e-wallet-system
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
go test -v ./internal/app/usecase/wallets/service/
```

---

## 📂 Project Structure

- `cmd/api`: Entry point for the API server.
- `cmd/migrate`: Entry point for database migrations.
- `configs`: Configuration files (.env).
- `internal/app`: Core application logic (Clean Architecture).
  - `dto/models`: Database models (Wallet, LedgerEntry, User).
  - `dto/domains`: API request/response DTOs.
  - `routes`: Gin route definitions.
  - `usecase/user`: User-related controllers, services, and repositories.
  - `usecase/wallets`: Wallet-related controllers, services, and repositories.
  - `database`: Database connection and migration files.
- `build`: Docker and CI/CD related files.

---

## 📡 API Specification

### 1. Get All Users

Retrieve the list of all registered users.

- **URL:** `/users`
- **Method:** `GET`

#### Example cURL

```bash
curl -X GET 'http://localhost:8080/users'
```

---

### 2. Get User Wallets

Retrieve all wallets owned by a specific user.

- **URL:** `/users/:user_id/wallets`
- **Method:** `GET`

#### Example cURL

```bash
curl -X GET 'http://localhost:8080/users/{user_id}/wallets'
```

#### Response (200 OK)

```json
{
  "user_id": "...",
  "wallets": [
    {
      "wallet_id": "...",
      "currency": "USD",
      "balance": "1000.00",
      "status": "ACTIVE"
    }
  ]
}
```

---

### 3. Create Wallet

Create a new wallet for a user with a specified currency. Each user can only have one wallet per currency.

- **URL:** `/wallets`
- **Method:** `POST`
- **Content-Type:** `application/json`

#### Request Body

```json
{
  "user_id": "...",
  "currency": "USD"
}
```

#### Example cURL

```bash
curl -X POST 'http://localhost:8080/wallets' \
  --header 'Content-Type: application/json' \
  --data '{
  "user_id": "your-user-id",
  "currency": "USD"
}'
```

#### Response (201 Created)

```json
{
  "wallet_id": "...",
  "user_id": "...",
  "currency": "USD",
  "balance": "0.00",
  "status": "ACTIVE"
}
```

---

### 4. Get Wallet By ID

Retrieve the current status and balance of a wallet. Suspended wallets are still readable.

- **URL:** `/wallets/:id`
- **Method:** `GET`

#### Example cURL

```bash
curl -X GET 'http://localhost:8080/wallets/{wallet_id}'
```

#### Response (200 OK)

```json
{
  "wallet_id": "...",
  "currency": "USD",
  "balance": "1000.50",
  "status": "ACTIVE"
}
```

---

### 5. Top-Up Wallet

Add funds to a wallet. Requires a unique `reference_id` for idempotency.

- **URL:** `/wallets/:id/topup`
- **Method:** `POST`
- **Content-Type:** `application/json`

#### Request Body

```json
{
  "amount": "100.00",
  "reference_id": "topup-unique-ref-001"
}
```

#### Example cURL

```bash
curl -X POST 'http://localhost:8080/wallets/{wallet_id}/topup' \
  --header 'Content-Type: application/json' \
  --data '{
  "amount": "100.00",
  "reference_id": "topup-unique-ref-001"
}'
```

#### Response (200 OK)

```json
{
  "wallet_id": "...",
  "currency": "USD",
  "balance": "1100.00",
  "status": "ACTIVE"
}
```

---

### 6. Pay From Wallet

Deduct funds from a wallet. Requires sufficient balance and a unique `reference_id`.

- **URL:** `/wallets/:id/pay`
- **Method:** `POST`
- **Content-Type:** `application/json`

#### Request Body

```json
{
  "amount": "50.00",
  "reference_id": "pay-unique-ref-001"
}
```

#### Example cURL

```bash
curl -X POST 'http://localhost:8080/wallets/{wallet_id}/pay' \
  --header 'Content-Type: application/json' \
  --data '{
  "amount": "50.00",
  "reference_id": "pay-unique-ref-001"
}'
```

#### Response (200 OK)

```json
{
  "wallet_id": "...",
  "currency": "USD",
  "balance": "950.00",
  "status": "ACTIVE"
}
```

---

### 7. Transfer Between Wallets

Transfer funds between two wallets. Both wallets must use the same currency.

- **URL:** `/wallets/transfer`
- **Method:** `POST`
- **Content-Type:** `application/json`

#### Request Body

```json
{
  "from_wallet_id": "...",
  "to_wallet_id": "...",
  "amount": "200.00",
  "reference_id": "transfer-unique-ref-001"
}
```

#### Example cURL

```bash
curl -X POST 'http://localhost:8080/wallets/transfer' \
  --header 'Content-Type: application/json' \
  --data '{
  "from_wallet_id": "sender-wallet-id",
  "to_wallet_id": "receiver-wallet-id",
  "amount": "200.00",
  "reference_id": "transfer-unique-ref-001"
}'
```

#### Response (200 OK)

```json
{
  "from_wallet": {
    "wallet_id": "...",
    "currency": "USD",
    "balance": "800.00",
    "status": "ACTIVE"
  },
  "to_wallet": {
    "wallet_id": "...",
    "currency": "USD",
    "balance": "1200.00",
    "status": "ACTIVE"
  }
}
```

---

### 8. Suspend Wallet

Suspend a wallet. Suspended wallets cannot perform top-ups, payments, or transfers. This operation is idempotent.

- **URL:** `/wallets/:id/suspend`
- **Method:** `POST`

#### Example cURL

```bash
curl -X POST 'http://localhost:8080/wallets/{wallet_id}/suspend'
```

#### Response (200 OK)

```json
{
  "wallet_id": "...",
  "currency": "USD",
  "balance": "500.00",
  "status": "SUSPENDED"
}
```

---

## ⚠️ Error Responses

All error responses follow this format:

```json
{
  "error": "descriptive error message"
}
```

| HTTP Status | Description |
|---|---|
| `400 Bad Request` | Invalid input (bad amount, missing fields, invalid currency) |
| `403 Forbidden` | Operation on a suspended wallet |
| `404 Not Found` | Wallet or user not found |
| `409 Conflict` | Concurrent modification or duplicate reference |
| `422 Unprocessable Entity` | Insufficient balance |
| `500 Internal Server Error` | Unexpected server error |
