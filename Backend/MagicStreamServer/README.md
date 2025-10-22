# MagicStreamServer Backend - Architecture Documentation

## Table of Contents

1. [Project Overview](#project-overview)
2. [Technology Stack](#technology-stack)
3. [System Architecture](#system-architecture)
4. [Project Structure](#project-structure)
5. [Core Components](#core-components)
6. [Data Models](#data-models)
7. [Authentication & Authorization](#authentication--authorization)
8. [API Architecture](#api-architecture)
9. [Database Design](#database-design)
10. [Security Features](#security-features)
11. [Middleware Pipeline](#middleware-pipeline)
12. [Dependency Injection](#dependency-injection)

---

## Project Overview

**MagicStreamServer** is a RESTful API server for a streaming platform built with Go. The application provides user authentication, movie management, genre categorization, and personalized movie recommendations based on user preferences.

### Key Features

- JWT-based authentication with access and refresh tokens
- Role-based access control (User/Admin)
- Movie catalog management with IMDB integration
- Genre-based movie categorization
- Personalized movie recommendations
- Comprehensive API documentation with Swagger
- Secure headers and CORS configuration

---

## Technology Stack

### Backend Framework & Language

- **Language**: Go 1.24.4
- **Web Framework**: Gin Gonic v1.11.0
- **Router**: Gin built-in router with middleware support

### Database

- **Database**: MongoDB v2.3.1
- **Driver**: Official MongoDB Go Driver v2
- **ODM Pattern**: Repository pattern implementation

### Authentication & Security

- **JWT**: golang-jwt/jwt v5.3.0
- **Password Hashing**: golang.org/x/crypto (bcrypt)
- **Token Types**:
  - Access Token (short-lived, 15 minutes default)
  - Refresh Token (long-lived, 168 hours default)

### API Documentation

- **Swagger/OpenAPI**: Swaggo v1.16.6
- **UI**: Gin-Swagger v1.6.1
- **Documentation**: Auto-generated from code annotations

### Configuration Management

- **Environment Variables**: godotenv v1.5.1
- **Config Pattern**: Centralized configuration loader

### Additional Libraries

- **Compression**: klauspost/compress
- **Validation**: go-playground/validator v10
- **JSON Processing**: bytedance/sonic (high-performance)

---

## System Architecture

### Architectural Pattern

The application follows a **Layered Architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────┐
│         Presentation Layer              │
│    (Gin Handlers + Middleware)          │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         Business Logic Layer            │
│     (Services + Controllers)            │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         Data Access Layer               │
│     (Repositories + Models)             │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│           Database Layer                │
│           (MongoDB)                     │
└─────────────────────────────────────────┘
```

### Request Flow

```
Client Request
    ↓
Security Headers Middleware
    ↓
CORS Middleware
    ↓
Request ID Middleware
    ↓
Authentication Middleware (if protected route)
    ↓
Authorization Middleware (if admin route)
    ↓
Route Handler
    ↓
Service Layer
    ↓
Repository Layer
    ↓
MongoDB
    ↓
Response
```

---

## Project Structure

```
MagicStreamServer/
│
├── config/                      # Configuration management
│   └── config.go               # Environment config loader
│
├── controllers/                 # Business logic controllers
│   └── auth/
│       └── tokenService.go     # JWT token service
│
├── database/                    # Database connection
│   └── databaseConnection.go   # MongoDB connection handler
│
├── docs/                        # Swagger documentation
│   ├── docs.go                 # Generated documentation
│   ├── swagger.json            # OpenAPI JSON spec
│   └── swagger.yaml            # OpenAPI YAML spec
│
├── middleware/                  # HTTP middlewares
│   ├── auth.go                 # JWT authentication
│   ├── secureHeaders.go        # Security headers & CORS
│   └── swagger.go              # Swagger UI middleware
│
├── models/                      # Data models
│   ├── moviesModel.go          # Movie & Genre structures
│   ├── tokenModel.go           # Token structures
│   └── usersModel.go           # User structures
│
├── repositories/                # Data access layer
│   ├── genre_repository.go     # Genre repository interface
│   ├── movie_repository.go     # Movie repository interface
│   ├── refresh_token_repository.go  # Token repository
│   ├── user_repository.go      # User repository interface
│   └── impl/                   # Repository implementations
│
├── routes/                      # Route handlers
│   ├── authRoute.go            # Authentication endpoints
│   ├── genreRoute.go           # Genre endpoints
│   ├── helloRoute.go           # Health check
│   └── movieRoute.go           # Movie endpoints
│
├── utils/                       # Utility functions
│   ├── errors.go               # Custom error handling
│   └── helpers.go              # Helper functions
│
├── .env                        # Environment variables
├── .gitignore                  # Git ignore rules
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
└── main.go                     # Application entry point
```

---

## Core Components

### 1. Configuration Management (`config/`)

**Purpose**: Centralized configuration loading from environment variables.

**Key Features**:

- Environment variable loading with defaults
- Type-safe configuration structure
- Support for .env files via godotenv

**Configuration Parameters**:

```go
type Config struct {
    Port                 string  // Server port (default: 5000)
    MongoURI            string  // MongoDB connection string
    GinMode             string  // Gin mode (debug/release)
    DatabaseName        string  // MongoDB database name
    BackendServerURI    string  // Backend server URI
    JWTAccessSecret     string  // Access token secret
    JWTRefreshSecret    string  // Refresh token secret
    AccessTokenExpireMin int    // Access token expiry (minutes)
    RefreshTokenExpireHr int    // Refresh token expiry (hours)
}
```

### 2. Database Layer (`database/`)

**Connection Management**:

- Connection pooling via official MongoDB driver
- Graceful connection and disconnection
- Context-based operations with timeout
- Collection factory pattern

**Key Functions**:

- `Connect()`: Establishes MongoDB connection
- `Disconnect()`: Gracefully closes connection
- `OpenCollection(name)`: Returns collection reference

### 3. Authentication Service (`controllers/auth/`)

**TokenService** - Centralized JWT management:

**Key Responsibilities**:

- Token generation (access + refresh pairs)
- Token validation and verification
- Token revocation management
- Refresh token rotation (single-use policy)
- Expired token cleanup

**Token Flow**:

```
Login/Register
    ↓
Generate Token Pair
    ↓
Store Refresh Token in DB
    ↓
Return to Client
    ↓
Client uses Access Token
    ↓
Access Token Expires
    ↓
Client sends Refresh Token
    ↓
Validate & Revoke Old Token
    ↓
Generate New Token Pair
```

### 4. Repository Pattern (`repositories/`)

**Design Pattern**: Interface-based repository pattern for data abstraction.

**Benefits**:

- Decoupling from database implementation
- Testability with mock repositories
- Consistent error handling
- Type-safe operations

**Repository Interfaces**:

- `UserRepository`: User CRUD operations
- `MovieRepository`: Movie management
- `GenreRepository`: Genre operations
- `RefreshTokenRepository`: Token lifecycle management

---

## Data Models

### User Model

```go
type User struct {
    ID              bson.ObjectID  // MongoDB ObjectID
    UserID          string         // Custom user identifier
    FirstName       string         // User first name
    LastName        string         // User last name
    Email           string         // Unique email
    Password        string         // Bcrypt hashed password
    Role            string         // USER or ADMIN
    CreatedAt       time.Time      // Account creation timestamp
    UpdatedAt       time.Time      // Last update timestamp
    FavouriteGenres []Genre        // User preferred genres
}
```

**Validation Rules**:

- Email must be valid format
- Password minimum 6 characters
- Names: 2-100 characters
- FavouriteGenres required on registration

### Movie Model

```go
type Movie struct {
    ID          bson.ObjectID  // MongoDB ObjectID
    ImdbID      string         // IMDb identifier (9-10 chars)
    Title       string         // Movie title (2-500 chars)
    PosterPath  string         // Poster image URL
    YouTubeID   string         // YouTube trailer ID (11 chars)
    Genre       []Genre        // Associated genres
    AdminReview string         // Admin review (max 1000 chars)
    Ranking     Ranking        // Rating information
}

type Genre struct {
    GenreID   int     // Unique genre identifier
    GenreName string  // Genre name (2-100 chars)
}

type Ranking struct {
    RankingValue int    // Rating value (1-10)
    RankingName  string // Rating label (2-50 chars)
}
```

### Token Model

```go
type RefreshToken struct {
    ID        bson.ObjectID  // MongoDB ObjectID
    UserID    string         // User identifier
    Token     string         // JWT refresh token
    ExpiresAt time.Time      // Expiration timestamp
    CreatedAt time.Time      // Creation timestamp
    Revoked   bool           // Revocation status
}
```

---

## Authentication & Authorization

### JWT Strategy

#### Access Token

- **Purpose**: Short-lived token for API access
- **Lifetime**: 15 minutes (configurable)
- **Algorithm**: HMAC-SHA256
- **Storage**: Client-side only (memory/sessionStorage)
- **Claims**:
  ```json
  {
    "sub": "user_id",
    "exp": 1234567890,
    "typ": "access"
  }
  ```

#### Refresh Token

- **Purpose**: Long-lived token for obtaining new access tokens
- **Lifetime**: 168 hours / 7 days (configurable)
- **Algorithm**: HMAC-SHA256
- **Storage**: Database + client-side (httpOnly cookie recommended)
- **Single-Use Policy**: Revoked after each use
- **Claims**:
  ```json
  {
    "sub": "user_id",
    "exp": 1234567890,
    "typ": "refresh"
  }
  ```

### Authorization Flow

#### Authentication Middleware

```go
func AuthMiddleware(ts *TokenService) gin.HandlerFunc
```

**Process**:

1. Extract Authorization header
2. Validate "Bearer <token>" format
3. Parse and verify JWT signature
4. Check token type (must be "access")
5. Extract user ID from claims
6. Store user_id in context
7. Continue to next handler

#### Authorization Middleware

```go
func AdminOnly() gin.HandlerFunc
```

**Process**:

1. Retrieve user_role from context
2. Verify role equals "ADMIN"
3. Reject if not admin
4. Continue to next handler

### Role-Based Access Control (RBAC)

**Roles**:

- `USER`: Standard user with basic access
- `ADMIN`: Administrator with elevated privileges

**Permission Matrix**:

| Endpoint                    | Anonymous | USER | ADMIN |
| --------------------------- | --------- | ---- | ----- |
| POST /auth/register         | ✓         | ✓    | ✓     |
| POST /auth/login            | ✓         | ✓    | ✓     |
| POST /auth/refresh          | ✓         | ✓    | ✓     |
| POST /auth/logout           | ✗         | ✓    | ✓     |
| GET /auth/me                | ✗         | ✓    | ✓     |
| PUT /auth/favorite-genres   | ✗         | ✓    | ✓     |
| GET /movies                 | ✓         | ✓    | ✓     |
| GET /movies/:id             | ✓         | ✓    | ✓     |
| GET /movies/recommendations | ✗         | ✓    | ✓     |
| POST /movies                | ✗         | ✗    | ✓     |
| PUT /movies/:id             | ✗         | ✗    | ✓     |
| DELETE /movies/:id          | ✗         | ✗    | ✓     |
| GET /genres                 | ✓         | ✓    | ✓     |
| POST /genres/seed           | ✗         | ✗    | ✓     |

---

## API Architecture

### API Versioning

- **Pattern**: Path-based versioning
- **Current Version**: v1
- **Base Path**: `/api/v1`

### Endpoint Organization

#### Authentication Endpoints (`/api/v1/auth`)

```
POST   /register              - User registration
POST   /login                 - User login
POST   /refresh               - Refresh access token
POST   /logout                - User logout (authenticated)
GET    /me                    - Get user profile (authenticated)
PUT    /favorite-genres       - Update favorite genres (authenticated)
```

#### Movie Endpoints (`/api/v1/movies`)

```
GET    /                      - Get all movies
GET    /:id                   - Get movie by ID
GET    /genre/:genre_id       - Get movies by genre
GET    /recommendations       - Get personalized recommendations (authenticated)
POST   /                      - Create movie (admin)
PUT    /:id                   - Update movie (admin)
DELETE /:id                   - Delete movie (admin)
```

#### Genre Endpoints (`/api/v1/genres`)

```
GET    /                      - Get all genres
GET    /:id                   - Get genre by ID
POST   /seed                  - Seed initial genres (admin)
```

#### Health Check (`/api/v1`)

```
GET    /health                - Server health status
```

### Response Standards

#### Success Response

```json
{
  "status": "success",
  "data": {
    // Response payload
  },
  "message": "Operation successful"
}
```

#### Error Response

```json
{
  "error": "Error message",
  "details": "Detailed error information"
}
```

#### Authentication Response

```json
{
  "user_id": "507f1f77bcf86cd799439011",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "role": "USER",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "favourite_genres": [{ "genre_id": 1, "genre_name": "Action" }]
}
```

---

## Database Design

### MongoDB Collections

#### Users Collection

```javascript
{
  _id: ObjectId("..."),
  user_id: "uuid-string",
  first_name: "John",
  last_name: "Doe",
  email: "john@example.com",
  password: "$2a$10$hashed...",
  role: "USER",
  created_at: ISODate("2025-01-15T10:30:00Z"),
  updated_at: ISODate("2025-01-15T10:30:00Z"),
  favourite_genres: [
    { genre_id: 1, genre_name: "Action" },
    { genre_id: 2, genre_name: "Comedy" }
  ]
}
```

**Indexes**:

- `email`: Unique index for fast lookup
- `user_id`: Index for query optimization

#### Movies Collection

```javascript
{
  _id: ObjectId("..."),
  imdb_id: "tt0111161",
  title: "The Shawshank Redemption",
  poster_path: "https://image.tmdb.org/...",
  youtube_id: "6hB3S9bIaco",
  genre: [
    { genre_id: 18, genre_name: "Drama" }
  ],
  admin_review: "Excellent movie...",
  ranking: {
    ranking_value: 9,
    ranking_name: "Masterpiece"
  }
}
```

**Indexes**:

- `imdb_id`: Unique index
- `genre.genre_id`: Multi-key index for genre filtering
- `ranking.ranking_value`: Index for sorting

#### Genres Collection

```javascript
{
  _id: ObjectId("..."),
  genre_id: 1,
  genre_name: "Action"
}
```

**Indexes**:

- `genre_id`: Unique index

#### Refresh Tokens Collection

```javascript
{
  _id: ObjectId("..."),
  user_id: "uuid-string",
  token: "eyJhbGciOiJIUzI1NiIs...",
  expires_at: ISODate("2025-01-22T10:30:00Z"),
  created_at: ISODate("2025-01-15T10:30:00Z"),
  revoked: false
}
```

**Indexes**:

- `user_id`: Index for user token lookup
- `token`: Index for token validation
- `expires_at`: Index for cleanup operations

### Database Operations

**Context Management**:

- All operations use context with timeout
- Default timeout: 10 seconds
- Graceful cancellation support

**Error Handling**:

- Custom error types per repository
- Consistent error mapping
- MongoDB error translation

---

## Security Features

### Security Headers (via SecureHeaders Middleware)

```http
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

### CORS Configuration

**Allowed Origins**:

- http://localhost:5000
- http://localhost:5173
- localhost

**Allowed Methods**: GET, POST, PUT, PATCH, DELETE, OPTIONS

**Allowed Headers**: Content-Type, Authorization, X-Requested-With, X-CSRF-Token

**Credentials**: Enabled

**Max Age**: 86400 seconds (24 hours)

### Password Security

**Hashing Algorithm**: Bcrypt

- Automatically handles salt generation
- Configurable cost factor
- One-way hash (cannot be reversed)

**Best Practices**:

- Minimum password length: 6 characters
- Passwords never logged or exposed
- Hash verification on login

### Token Security

**Storage Recommendations**:

- Access Token: Memory or sessionStorage (client)
- Refresh Token: httpOnly cookie or secure storage

**Token Rotation**:

- Refresh tokens are single-use
- Old tokens revoked on refresh
- Automatic cleanup of expired tokens

---

## Middleware Pipeline

### Global Middlewares (Applied to All Routes)

1. **SecureHeaders**: Adds security headers to responses
2. **CORS**: Handles cross-origin requests
3. **RequestID**: Adds unique request identifier

### Route-Specific Middlewares

4. **AuthMiddleware**: Validates JWT access token
5. **AdminOnly**: Restricts access to admin users

### Middleware Execution Order

```
Request → SecureHeaders → CORS → RequestID → [AuthMiddleware] → [AdminOnly] → Handler
```

**Example Protected Route**:

```go
movies.POST("",
    middleware.AuthMiddleware(ts),
    middleware.AdminOnly(),
    movieHandler.Create,
)
```

---

## Dependency Injection

### Service Initialization Pattern

The application uses **constructor-based dependency injection** to manage dependencies:

```go
// main.go
func main() {
    // 1. Load configuration
    cfg := config.LoadConfig()

    // 2. Connect to database
    database.Connect()
    defer database.Disconnect()

    // 3. Initialize repositories
    userRepo := repositories.NewUserRepository(
        database.OpenCollection("users")
    )
    movieRepo := repositories.NewMovieRepository(
        database.OpenCollection("movies")
    )
    genreRepo := repositories.NewGenreRepository(
        database.OpenCollection("genres")
    )
    refreshTokenRepo := repositories.NewRefreshTokenRepository(
        database.OpenCollection("refresh_token")
    )

    // 4. Initialize services
    tokenService := authservice.NewTokenService(cfg, refreshTokenRepo)

    // 5. Setup routes with dependencies
    setupRoutes(router, tokenService, userRepo, movieRepo, genreRepo)
}
```

### Benefits of This Pattern

1. **Testability**: Easy to mock dependencies
2. **Flexibility**: Simple to swap implementations
3. **Clarity**: Explicit dependency requirements
4. **Lifecycle Management**: Clear initialization order
5. **No Global State**: All dependencies passed explicitly

---

## Development & Deployment

### Environment Variables

Required environment variables:

```env
PORT=5000
MONGO_URI=mongodb://localhost:27017
DATABASE_NAME=magic_stream
GIN_MODE=debug
JWT_ACCESS_SECRET=your-access-secret
JWT_REFRESH_SECRET=your-refresh-secret
ACCESS_TOKEN_EXPIRE_MINUTES=15
REFRESH_TOKEN_EXPIRE_HOURS=168
BACKEND_URI=http://localhost:5000
```

### Running the Application

**Development Mode**:

```bash
go run main.go
```

**Production Build**:

```bash
go build -o magic-stream-server
./magic-stream-server
```

### API Documentation Access

Once the server is running:

- **Swagger UI**: http://localhost:5000/swagger/index.html
- **Health Check**: http://localhost:5000/api/v1/health

### Generating Swagger Documentation

```bash
swag init -g main.go --output ./docs
```

---

## Design Patterns & Best Practices

### Patterns Implemented

1. **Repository Pattern**: Data access abstraction
2. **Dependency Injection**: Constructor injection
3. **Factory Pattern**: Repository and service creation
4. **Middleware Chain**: Request processing pipeline
5. **Configuration Pattern**: Centralized config management

### Code Organization Principles

1. **Separation of Concerns**: Clear layer boundaries
2. **Single Responsibility**: Each component has one job
3. **Interface Segregation**: Focused, minimal interfaces
4. **Dependency Inversion**: Depend on abstractions

### Error Handling Strategy

1. **Custom Errors**: Domain-specific error types
2. **Error Wrapping**: Preserve error context
3. **Consistent Responses**: Standardized error format
4. **Graceful Degradation**: Fail safely

---

## Performance Considerations

### Database Optimization

- **Connection Pooling**: Managed by MongoDB driver
- **Indexes**: Strategic indexing on frequent queries
- **Projections**: Fetch only required fields
- **Query Optimization**: Efficient filter design

### API Performance

- **JSON Serialization**: High-performance Sonic library
- **Middleware Efficiency**: Minimal overhead
- **Response Caching**: (Recommended for production)
- **Connection Reuse**: HTTP keep-alive enabled

---

## Future Enhancements

### Recommended Improvements

1. **Rate Limiting**: Implement per-user/IP rate limits
2. **Logging**: Structured logging with correlation IDs
3. **Monitoring**: Metrics collection and alerting
4. **Testing**: Unit and integration test coverage
5. **Caching**: Redis for session and response caching
6. **Database Migrations**: Versioned schema management
7. **API Gateway**: Centralized API management
8. **Microservices**: Service decomposition for scale

---

## Conclusion

MagicStreamServer is a well-architected Go application that demonstrates best practices in API design, security, and code organization. The layered architecture, repository pattern, and dependency injection make it maintainable, testable, and scalable for future growth.

**Key Strengths**:

- Clean separation of concerns
- Robust authentication system
- Type-safe operations with Go
- Comprehensive API documentation
- Security-first approach
- Scalable architecture

**Version**: 1.0  
**Last Updated**: October 2025  
**Go Version**: 1.24.4
