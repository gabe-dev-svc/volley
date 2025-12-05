# Integration Tests

This directory contains integration tests for the Volley API that use Testcontainers to spin up a real PostgreSQL database with PostGIS extension.

## Prerequisites

- Docker Desktop must be running
- Go 1.21 or higher

## Running the Tests

### Run all integration tests

```bash
cd tests/integration
go test -v -count=1 -timeout=180s
```

### Run a specific test

```bash
go test -v -count=1 -run TestRegister_Success
```

### Run tests for a specific area

```bash
# Auth tests
go test -v -count=1 -run "^TestRegister|^TestLogin"

# Game tests
go test -v -count=1 -run "^TestCreateGame|^TestListGames"

# Participant tests
go test -v -count=1 -run "^TestJoinGame|^TestDropGame"
```

## How it Works

1. **TestMain setup** (`setup_test.go`):
   - Starts a PostgreSQL 16 + PostGIS 3.4 container using Testcontainers
   - Runs database migrations from `internal/repository/schema.sql`
   - Starts the Gin API server on port 8081
   - All tests run against this server instance
   - Cleanup happens automatically after all tests complete

2. **Test isolation**:
   - Each test creates its own test users and games
   - Cleanup is done via `defer` statements using direct database queries
   - Tests can run in parallel (though currently they run sequentially)

3. **HTTP Client** (`helpers.go`):
   - `TestClient` wraps Go's `http.Client` with convenience methods
   - Automatically handles auth token injection
   - Helper methods like `RegisterUser()`, `CreateGame()`, etc.

## Test Structure

- `setup_test.go` - TestMain and container setup
- `helpers.go` - HTTP client, cleanup utilities, assertion helpers
- `auth_test.go` - Tests for /auth/register and /auth/login endpoints
- `games_test.go` - Tests for game CRUD operations and listing
- `participants_test.go` - Tests for joining, dropping, and waitlist functionality

## Test Coverage

### Auth Endpoints
- ✅ Register with valid data
- ✅ Register with duplicate email (should fail)
- ✅ Register with invalid email (should fail)
- ✅ Register with missing required fields (should fail)
- ✅ Login with valid credentials
- ✅ Login with invalid password (should fail)
- ✅ Login with nonexistent user (should fail)
- ✅ Login with empty credentials (should fail)

### Game Endpoints
- ✅ Create game with valid data
- ✅ Create game without authentication (should fail)
- ✅ Create game with invalid duration (should fail)
- ✅ Create game with invalid max participants (should fail)
- ✅ List games with spatial filtering
- ✅ List games with multiple categories
- ✅ List games with missing required parameters (should fail)
- ✅ List games returns empty array when no results

### Participant Endpoints
- ✅ Join game successfully
- ✅ Join game without authentication (should fail)
- ✅ Join game is idempotent
- ✅ Join game that has finished (should fail)
- ✅ Drop from game successfully
- ✅ Drop from game is idempotent
- ✅ Drop from game when not a participant (should fail)
- ✅ Drop from game without authentication (should fail)
- ✅ Drop after deadline (should fail)

### Waitlist Functionality
- ✅ User is waitlisted when game is full
- ✅ Multiple waitlisted users have correct positions
- ✅ Waitlist promotion when confirmed player drops (automatic)
- ✅ Second waitlisted player has position 2
- ✅ Dropping from waitlist does not affect confirmed participants

## Troubleshooting

**Tests fail with "cannot connect to Docker"**
- Ensure Docker Desktop is running
- Check that Docker socket is accessible

**Tests timeout**
- Increase the timeout: `go test -timeout=300s`
- Check if container is starting correctly
- Verify your Docker has enough resources allocated

**Database migration errors**
- Check that `internal/repository/schema.sql` exists and is valid
- Ensure PostGIS extension is being created successfully

**Port 8081 already in use**
- Stop any other process using port 8081
- Or modify the port in `setup_test.go`

## CI/CD Integration

To run these tests in CI/CD:

```yaml
# GitHub Actions example
steps:
  - uses: actions/checkout@v3
  - uses: actions/setup-go@v4
    with:
      go-version: '1.21'
  - name: Run integration tests
    run: |
      cd tests/integration
      go test -v -count=1 -timeout=300s
```

The tests will automatically pull the PostGIS Docker image and manage the container lifecycle.
