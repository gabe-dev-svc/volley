package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gabe-dev-svc/volley/internal/models"
)

// TestClient wraps http.Client with helper methods for testing
type TestClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string // Store auth token after login/register
}

// NewTestClient creates a new test HTTP client
func NewTestClient() *TestClient {
	return &TestClient{
		BaseURL: testBaseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// POST makes a POST request and unmarshals the response
func (c *TestClient) POST(path string, body interface{}, response interface{}) (*http.Response, error) {
	return c.request("POST", path, body, response)
}

// GET makes a GET request and unmarshals the response
func (c *TestClient) GET(path string, response interface{}) (*http.Response, error) {
	return c.request("GET", path, nil, response)
}

// PATCH makes a PATCH request and unmarshals the response
func (c *TestClient) PATCH(path string, body interface{}, response interface{}) (*http.Response, error) {
	return c.request("PATCH", path, body, response)
}

// DELETE makes a DELETE request
func (c *TestClient) DELETE(path string) (*http.Response, error) {
	return c.request("DELETE", path, nil, nil)
}

// request is the core HTTP request method
func (c *TestClient) request(method, path string, body interface{}, response interface{}) (*http.Response, error) {
	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Type", "mobile") // Get token in JSON response

	// Add auth token if available
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, fmt.Errorf("read response: %w", err)
	}

	// Unmarshal response if provided
	if response != nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, response); err != nil {
			return resp, fmt.Errorf("unmarshal response: %w (body: %s)", err, string(bodyBytes))
		}
	}

	return resp, nil
}

// RegisterUser creates a new user and returns the auth response
func (c *TestClient) RegisterUser(email, password, firstName, lastName string) (*models.AuthResponse, error) {
	req := models.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}

	var resp models.AuthResponse
	httpResp, err := c.POST("/v1/auth/register", req, &resp)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("registration failed: status %d", httpResp.StatusCode)
	}

	// Store token for authenticated requests
	if resp.Token != nil {
		c.Token = *resp.Token
	}

	return &resp, nil
}

// LoginUser logs in and stores the token
func (c *TestClient) LoginUser(email, password string) (*models.AuthResponse, error) {
	req := models.LoginRequest{
		Email:    email,
		Password: password,
	}

	var resp models.AuthResponse
	httpResp, err := c.POST("/v1/auth/login", req, &resp)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed: status %d", httpResp.StatusCode)
	}

	// Store token for authenticated requests
	if resp.Token != nil {
		c.Token = *resp.Token
	}

	return &resp, nil
}

// CreateGame creates a game for the authenticated user
func (c *TestClient) CreateGame(req models.CreateGameRequest) (*models.Game, error) {
	var game models.Game
	httpResp, err := c.POST("/v1/games", req, &game)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create game failed: status %d", httpResp.StatusCode)
	}

	return &game, nil
}

// CleanupUser deletes a user and all related data from the test database
func CleanupUser(ctx context.Context, userID string) error {
	// Delete in correct order due to foreign keys
	queries := []string{
		"DELETE FROM participants WHERE user_id = $1",
		"DELETE FROM games WHERE owner_id = $1",
		"DELETE FROM users WHERE id = $1",
	}

	for _, query := range queries {
		if _, err := testDBPool.Exec(ctx, query, userID); err != nil {
			return fmt.Errorf("cleanup user failed: %w", err)
		}
	}

	return nil
}

// CleanupGame deletes a game and all related data from the test database
func CleanupGame(ctx context.Context, gameID string) error {
	queries := []string{
		"DELETE FROM participants WHERE game_id = $1",
		"DELETE FROM teams WHERE game_id = $1",
		"DELETE FROM games WHERE id = $1",
	}

	for _, query := range queries {
		if _, err := testDBPool.Exec(ctx, query, gameID); err != nil {
			return fmt.Errorf("cleanup game failed: %w", err)
		}
	}

	return nil
}

// TestEmail generates a unique email for testing
func TestEmail(t *testing.T) string {
	return fmt.Sprintf("test_%s_%d@example.com", t.Name(), time.Now().UnixNano())
}

// AssertStatusCode fails the test if the status code doesn't match
func AssertStatusCode(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Fatalf("expected status %d, got %d", expected, actual)
	}
}

// AssertNoError fails the test if error is not nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Helper functions for pointer types
func strPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}
