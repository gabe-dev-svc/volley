package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/gabe-dev-svc/volley/internal/models"
)

func TestRegister_Success(t *testing.T) {
	client := NewTestClient()
	ctx := context.Background()

	email := TestEmail(t)

	// Register user
	resp, err := client.RegisterUser(email, "password123@", "John", "Doe")
	AssertNoError(t, err)

	// Cleanup
	defer CleanupUser(ctx, resp.User.ID)

	// Assertions
	if resp.User.Email != email {
		t.Errorf("expected email %s, got %s", email, resp.User.Email)
	}
	if resp.User.FirstName != "John" {
		t.Errorf("expected first name John, got %s", resp.User.FirstName)
	}
	if resp.User.LastName != "Doe" {
		t.Errorf("expected last name Doe, got %s", resp.User.LastName)
	}
	if resp.Token == nil {
		t.Error("expected token in response")
	}
	if resp.User.ID == "" {
		t.Error("expected user ID in response")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	client := NewTestClient()
	ctx := context.Background()

	email := TestEmail(t)

	// First registration
	resp1, err := client.RegisterUser(email, "password123@", "John", "Doe")
	AssertNoError(t, err)
	defer CleanupUser(ctx, resp1.User.ID)

	// Second registration with same email
	client2 := NewTestClient()
	req := models.RegisterRequest{
		Email:     email,
		Password:  "password456@",
		FirstName: "Jane",
		LastName:  "Smith",
	}

	var resp2 models.AuthResponse
	httpResp, _ := client2.POST("/v1/auth/register", req, &resp2)

	AssertStatusCode(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestRegister_InvalidEmail(t *testing.T) {
	client := NewTestClient()

	req := models.RegisterRequest{
		Email:     "not-an-email",
		Password:  "password123@",
		FirstName: "John",
		LastName:  "Doe",
	}

	var resp models.AuthResponse
	httpResp, _ := client.POST("/v1/auth/register", req, &resp)

	// Should fail validation
	AssertStatusCode(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestRegister_MissingFields(t *testing.T) {
	client := NewTestClient()

	tests := []struct {
		name string
		req  models.RegisterRequest
	}{
		{
			name: "missing email",
			req: models.RegisterRequest{
				Password:  "password123@",
				FirstName: "John",
				LastName:  "Doe",
			},
		},
		{
			name: "missing password",
			req: models.RegisterRequest{
				Email:     TestEmail(t),
				FirstName: "John",
				LastName:  "Doe",
			},
		},
		{
			name: "missing first name",
			req: models.RegisterRequest{
				Email:    TestEmail(t),
				Password: "password123@",
				LastName: "Doe",
			},
		},
		{
			name: "missing last name",
			req: models.RegisterRequest{
				Email:     TestEmail(t),
				Password:  "password123@",
				FirstName: "John",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp models.AuthResponse
			httpResp, _ := client.POST("/v1/auth/register", tt.req, &resp)

			AssertStatusCode(t, http.StatusBadRequest, httpResp.StatusCode)
		})
	}
}

func TestLogin_Success(t *testing.T) {
	client := NewTestClient()
	ctx := context.Background()

	email := TestEmail(t)
	password := "password123@"

	// Register user
	regResp, err := client.RegisterUser(email, password, "John", "Doe")
	AssertNoError(t, err)
	defer CleanupUser(ctx, regResp.User.ID)

	// Login with new client
	client2 := NewTestClient()
	loginResp, err := client2.LoginUser(email, password)
	AssertNoError(t, err)

	// Assertions
	if loginResp.User.Email != email {
		t.Errorf("expected email %s, got %s", email, loginResp.User.Email)
	}
	if loginResp.User.ID != regResp.User.ID {
		t.Errorf("expected user ID %s, got %s", regResp.User.ID, loginResp.User.ID)
	}
	if loginResp.Token == nil {
		t.Error("expected token in response")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	client := NewTestClient()
	ctx := context.Background()

	email := TestEmail(t)

	// Register user
	regResp, err := client.RegisterUser(email, "password123@", "John", "Doe")
	AssertNoError(t, err)
	defer CleanupUser(ctx, regResp.User.ID)

	// Login with wrong password
	client2 := NewTestClient()
	req := models.LoginRequest{
		Email:    email,
		Password: "wrongpassword",
	}

	var resp models.AuthResponse
	httpResp, _ := client2.POST("/v1/auth/login", req, &resp)

	AssertStatusCode(t, http.StatusUnauthorized, httpResp.StatusCode)
}

func TestLogin_NonexistentUser(t *testing.T) {
	client := NewTestClient()

	req := models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123@",
	}

	var resp models.AuthResponse
	httpResp, _ := client.POST("/v1/auth/login", req, &resp)

	AssertStatusCode(t, http.StatusUnauthorized, httpResp.StatusCode)
}

func TestLogin_EmptyCredentials(t *testing.T) {
	client := NewTestClient()

	req := models.LoginRequest{
		Email:    "",
		Password: "",
	}

	var resp models.AuthResponse
	httpResp, _ := client.POST("/v1/auth/login", req, &resp)

	AssertStatusCode(t, http.StatusBadRequest, httpResp.StatusCode)
}
