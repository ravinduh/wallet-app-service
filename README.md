# Wallet App Service

A RESTful API service for a centralized wallet application that allows users to manage their digital wallets.

## Features

- User balance management
- Deposits and withdrawals
- Money transfers between users
- Transaction history

## Tech Stack

- Language: Go
- Database: PostgreSQL
- Architecture: Clean Architecture

## Project Structure

```
wallet-app-service/
├── cmd/                    # Application entry points
│   ├── api/                # API server
│   └── seed/               # Database seeder
├── internal/               # Private application code
│   ├── config/             # Configuration
│   ├── domain/             # Domain models and interfaces
│   ├── handler/            # HTTP handlers
│   ├── repository/         # Data access layer
│   ├── usecase/            # Business logic
│   └── middleware/         # HTTP middleware
├── pkg/                    # Public libraries
│   ├── database/           # Database helpers
│   ├── logging/            # Logging utilities
│   └── errors/             # Error handling
└── migrations/             # SQL migrations
```

## API Documentation

### Base URL

All API endpoints are relative to:

```
http://localhost:8080/api/v1
```

### Request Headers

| Header | Description |
|--------|-------------|
| `Request-Id` | Optional unique identifier for the request. If not provided, a UUID will be generated |

### Response Format

All API responses follow this standard format:

**Success Response:**
```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "data": { /* response data varies by endpoint */ }
}
```

**Error Response:**
```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "error": "Error message"
}
```

### Endpoints

#### 1. Deposit Money

**Endpoint:** `POST /deposit`

Adds money to a user's wallet.

**Request Body:**

```json
{
  "user_id": 1,
  "amount": 500.00,
  "comment": "Initial deposit"
}
```

| Field | Type | Description |
|-------|------|-------------|
| user_id | integer | ID of the user |
| amount | number | Amount to deposit (must be positive) |
| comment | string | Optional description for the transaction |

**Response Example:**

```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "data": {
    "id": 1,
    "wallet_id": 1,
    "dest_wallet_id": null,
    "type": "DEPOSIT",
    "amount": 500.00,
    "balance_before": 1000.00,
    "balance_after": 1500.00,
    "description": "Initial deposit",
    "transaction_time": "2023-05-12T10:30:45Z",
    "created_at": "2023-05-12T10:30:45Z"
  }
}
```

#### 2. Withdraw Money

**Endpoint:** `POST /withdraw`

Withdraws money from a user's wallet.

**Request Body:**

```json
{
  "user_id": 1,
  "amount": 200.00,
  "comment": "ATM withdrawal"
}
```

| Field | Type | Description |
|-------|------|-------------|
| user_id | integer | ID of the user |
| amount | number | Amount to withdraw (must be positive) |
| comment | string | Optional description for the transaction |

#### 3. Transfer Money

**Endpoint:** `POST /transfer`

Transfers money from one user to another.

**Request Body:**

```json
{
  "sender_id": 1,
  "receiver_id": 2,
  "amount": 300.00,
  "comment": "Dinner payment"
}
```

| Field | Type | Description |
|-------|------|-------------|
| sender_id | integer | ID of the sending user |
| receiver_id | integer | ID of the receiving user |
| amount | number | Amount to transfer (must be positive) |
| comment | string | Optional description for the transaction |

#### 4. Get Wallet Balance

**Endpoint:** `GET /balance/{userID}`

Retrieves the current balance of a user's wallet.

**URL Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| userID | integer | ID of the user |

#### 5. Get Transaction History

**Endpoint:** `GET /transactions/{userID}`

Retrieves the transaction history for a user's wallet.

**URL Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| userID | integer | ID of the user |

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| limit | integer | Maximum number of transactions to return | 10 |
| offset | integer | Number of transactions to skip | 0 |

### Status Codes

The API uses the following status codes:

- `200 OK` - The request was successful
- `400 Bad Request` - The request was invalid or cannot be otherwise served
- `404 Not Found` - The requested resource does not exist
- `500 Internal Server Error` - Server error

### Data Types

#### Transaction Types

| Type | Description |
|------|-------------|
| DEPOSIT | Money added to wallet |
| WITHDRAWAL | Money removed from wallet |
| TRANSFER | Money sent to another user |

#### Currency

Currently, only `USD` is supported.

### Testing the API

A Postman collection is included in the repository as `wallet-app-postman-collection.json`. Import this file into Postman to test the API endpoints. The collection includes request ID tracking and test scripts to verify responses.

## Setup and Running

### Prerequisites

- Docker and Docker Compose

### Running the Application

```bash
# Start all services (app, postgres, migrations)
docker compose up

# Run in background
docker compose up -d
```

This single command will:
1. Start a PostgreSQL database
2. Start a Redis instance for caching and distributed locking
3. Run database migrations automatically
4. Seed the database with test data
5. Start the API server on port 8080

### Accessing the API

Once running, the API will be available at:
```
http://localhost:8080/api/v1
```

You can use the included Postman collection (`wallet-app-postman-collection-v2.json`) or the demo script (`./examples/demo.sh`) to test the API endpoints.

### Running Tests

Tests can be run inside the Docker container:

```bash
# Run tests within the container
docker compose exec app go test ./...
```

### Stopping the Application

```bash
# Stop all services
docker compose down

# Stop and remove volumes (will delete database data)
docker compose down -v
```

## Design Decisions

### Clean Architecture

The project follows clean architecture principles with clear separation of concerns:

- **Domain Layer**: Contains the business models and repository interfaces
- **Repository Layer**: Implements data access logic
- **Use Case Layer**: Implements business logic
- **Handler Layer**: Manages HTTP requests/responses

### Technology Choices

#### Go-Chi vs. Go-Kit

After careful consideration, I chose Go-Chi as the HTTP framework for this project:

**Why Go-Chi:**
- **Simplicity and Directness**: Go-Chi provides a lightweight, straightforward approach to HTTP routing without unnecessary abstractions
- **Standard Library Alignment**: Built on top of the standard `net/http` library, maintaining Go's simplicity
- **Middleware Support**: Excellent middleware system that's easy to extend (e.g., our request ID tracking)
- **Development Speed**: Lower learning curve and less boilerplate accelerates development
- **Maintainability**: Simpler codebase is easier for teams to maintain and understand

**Trade-offs Considered:**
- **Go-Kit** would offer more built-in features for microservices (service discovery, circuit breaking, multiple transports), but introduces additional complexity not needed at our current scale
- For future scaling, we could either:
  1. Gradually add the specific resilience patterns we need (circuit breakers, etc.)
  2. Refactor toward Go-Kit if a full microservices architecture becomes necessary
  3. Keep Go-Chi and use service mesh solutions (like Istio) for advanced network features

#### Project Structure

- **cmd/**: Application entry points
- **internal/**: Application-specific code (not meant for external reuse)
- **pkg/**: Shared utilities that could potentially be used by other services
  - Keeps cross-cutting concerns separate from business logic
  - Promotes code reuse for common patterns like response formatting

### Request ID Tracking

Every request flows through the system with a unique identifier:

- Accept client-provided `Request-Id` header or generate a UUID
- Include request ID in all logs for traceability
- Return request ID in all responses for correlation
- Standardized JSON responses with `request_id`, `data`, and `error` fields

### Transaction Handling

To ensure data consistency, money transfers ensure:

1. Sufficient balance checks before withdrawal
2. Atomicity in updating both sender and receiver wallets

### Error Handling

The service implements proper error handling with:
- Standardized error response format
- Contextual error messages with request IDs
- Appropriate HTTP status codes
- Detailed logging for debugging

## Database Design

The application uses PostgreSQL with the following schema:

### Entity Relationship Diagram

```
+----------------+       +----------------+       +------------------+
|     users      |       |    wallets     |       |   transactions   |
+----------------+       +----------------+       +------------------+
| id (PK)        |       | id (PK)        |       | id (PK)          |
| username       |       | user_id (FK)   |------>| wallet_id (FK)   |
| email          |       | balance        |       | dest_wallet_id   |
| created_at     |------>| currency       |       | type             |
| updated_at     |       | created_at     |       | amount           |
+----------------+       | updated_at     |       | balance_before   |
                         +----------------+       | balance_after    |
                                                  | description      |
                                                  | transaction_time |
                                                  | created_at       |
                                                  +------------------+
```

### Table Structures

#### Users Table
```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  username VARCHAR(255) NOT NULL UNIQUE,
  email VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
```

#### Wallets Table
```sql
CREATE TABLE wallets (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  balance DECIMAL(19, 4) NOT NULL DEFAULT 0,
  currency VARCHAR(10) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  UNIQUE(user_id)
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
```

#### Transactions Table
```sql
CREATE TABLE transactions (
  id SERIAL PRIMARY KEY,
  wallet_id INTEGER NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
  dest_wallet_id INTEGER REFERENCES wallets(id) ON DELETE SET NULL,
  type VARCHAR(20) NOT NULL,
  amount DECIMAL(19, 4) NOT NULL,
  balance_before DECIMAL(19, 4) NOT NULL,
  balance_after DECIMAL(19, 4) NOT NULL,
  description TEXT,
  transaction_time TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX idx_transactions_transaction_time ON transactions(transaction_time);
```

### Design Decisions

1. **One-to-One User to Wallet Relationship**:
   - Each user has exactly one wallet
   - Enforced by a unique constraint on `user_id` in the wallets table

2. **Transaction History**:
   - Transactions track both `wallet_id` and optional `dest_wallet_id` for transfers
   - Balance snapshots (`balance_before` and `balance_after`) provide audit capability
   - Transaction types (DEPOSIT, WITHDRAWAL, TRANSFER) define the operation

3. **Indexing Strategy**:
   - Indexed `wallet_id` for fast transaction lookups by wallet
   - Indexed `transaction_time` for efficient historical queries and pagination
   - Default indexing on all primary and foreign keys

4. **Precision for Financial Data**:
   - Used `DECIMAL(19, 4)` for all monetary values
   - Supports up to 15 digits before decimal point and 4 digits after
   - Ensures accurate financial calculations without floating-point errors

## Redis Integration

The application uses Redis for two key optimizations:

1. **Balance Caching**
   - Balance queries are cached with short TTL (5 seconds)
   - Cache is automatically invalidated on balance changes
   - Improves performance for frequent balance checks
   
   **5-Second TTL Rationale:**
   - Balances change frequently in a wallet application
   - Short TTL ensures users quickly see their latest balance
   - Provides a good balance between consistency and performance
   - Acts as a safety net for any missed explicit invalidations
   - Conservative approach appropriate for financial data

2. **Distributed Locking for Transfers**
   - Prevents race conditions during transfers between wallets
   - Uses a lock ordering strategy to prevent deadlocks
   - Ensures consistency when multiple transfers involve the same wallets
   
   **Lock Implementation Details:**
   - Orders locks by user ID to prevent deadlocks (lower ID first)
   - Uses 10-second timeout to prevent infinite blocking
   - Automatically releases locks using defer statements
   - Fails gracefully if locks cannot be acquired
   - Handles Redis connection issues without blocking operations

3. **Hybrid Caching Strategy**
   - Combines TTL-based expiration with explicit invalidation
   - Any balance-changing operation (deposit, withdraw, transfer) invalidates relevant caches
   - Transfers invalidate cache for both sender and receiver
   - Degrades gracefully if Redis is unavailable (falls back to database)
   - No stale data dependencies - application works correctly without Redis

## Areas for Improvement

- Authentication and Authorization
  - Complete the JWT token validation in auth middleware
  - Implement user registration and login endpoints
  - Add role-based access control (admin vs regular users)
  - Create token refresh mechanism
  - Add token revocation/blacklisting

- Redis Enhancements
  - Implement rate limiting using Redis
  - Add transaction idempotency keys to prevent duplicates
  - Create a circuit breaker using Redis health status
  - Implement smarter cache invalidation strategies

- Security Enhancements
  - Implement database transactions for atomicity
  - Add rate limiting to prevent abuse
  - Add request validation middleware
  - Implement HTTPS with proper certificate management

- Observability
  - Enhance logging with structured fields
  - Add metrics collection (Prometheus)
  - Implement distributed tracing
  - Create health check endpoints with detailed status

- Documentation
  - Add API documentation with Swagger/OpenAPI
  - Include authentication flows in documentation
  - Create postman collection with auth examples

## Planning and Implementation Approach

### Initial Planning

Before writing a single line of code, I spent time planning the application:

1. **Domain Modeling**
   - Identified core entities: Users, Wallets, Transactions
   - Established entity relationships (one user has one wallet, transactions reference wallets)
   - Defined constraints (e.g., non-negative balances, unique usernames)

2. **Architecture Design**
   - Selected Clean Architecture for clear separation of concerns
   - Set up layered structure: domain → repository → usecase → handler
   - Decided on dependency injection pattern for testability

3. **Database Schema Planning**
   - Created table designs with proper relationships and constraints
   - Added appropriate indexes for query performance
   - Ensured transaction tracking had necessary fields for auditing

4. **API Design**
   - Created RESTful endpoints for all required operations
   - Defined request/response formats
   - Planned authentication mechanism (with skeleton implementation)

5. **Technology Selection**
   - Go and Chi for simplicity and performance
   - PostgreSQL for reliable transactional data storage
   - Redis for performance optimizations
   - Docker for easy deployment and consistent environments

6. **Non-Functional Requirements**
   - Considered concurrency issues (distributed locking)
   - Planned for data consistency (cache invalidation)
   - Designed for observability (request IDs, structured logging)

By investing in planning before implementation, I was able to:
- Create a cohesive architecture that's easy to understand
- Avoid major refactoring during development
- Ensure all requirements were met systematically
- Future-proof the design for upcoming features

## Development Time

- Time spent: 12 hours

## Features Not Implemented

- Authentication and authorization
- Detailed transaction analytics
- Multiple currency support
- Pagination for large result sets
- File uploads (e.g., receipts)

## How to Review the Code

1. Start with domain models in `internal/domain/`
2. Review use cases in `internal/usecase/`
3. Check HTTP handlers in `internal/handler/`
4. Examine tests in `*_test.go` files