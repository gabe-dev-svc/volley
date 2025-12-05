# Mocks

This directory contains auto-generated mocks for testing, generated using [mockery v3](https://vektra.github.io/mockery/).

## Generating Mocks

Mocks are generated from interfaces defined in `ifaces/volley.go`.

To regenerate mocks after modifying interfaces:

```bash
mockery
```

The configuration is in `.mockery.yml` at the project root.

## Usage in Tests

```go
import (
    "testing"
    "github.com/gabe-dev-svc/volley/mocks"
)

func TestExample(t *testing.T) {
    // Create a new mock with automatic assertion cleanup
    mockQuerier := mocks.NewQuerier(t)

    // Set expectations
    mockQuerier.On("GetGame", mock.Anything, gameUUID).
        Return(repository.GetGameRow{...}, nil)

    // Use the mock
    service := NewGamesService(mockQuerier)

    // Assertions are automatically verified via t.Cleanup()
}
```

## Available Mocks

- **Querier** (`mocks.Querier`) - Database query interface mock
- **IFaceTest** (`mocks.IFaceTest`) - Test interface example

## Notes

- Mocks are automatically regenerated when interfaces change
- All mocks use testify/mock for assertions
- Cleanup and assertion verification happens automatically via `t.Cleanup()`
