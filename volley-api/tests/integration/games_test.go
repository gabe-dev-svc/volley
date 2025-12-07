package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gabe-dev-svc/volley/internal/models"
)

func TestCreateGame_Success(t *testing.T) {
	client := NewTestClient()
	ctx := context.Background()

	// Register and login user
	email := TestEmail(t)
	authResp, err := client.RegisterUser(email, "password123@", "John", "Doe")
	AssertNoError(t, err)
	defer CleanupUser(ctx, authResp.User.ID)

	// Create game request
	startTime := time.Now().Add(24 * time.Hour)
	skillLevel := models.SkillLevelAll

	req := models.CreateGameRequest{
		Category:    models.GameCategoryBasketball,
		Title:       strPtr("Pickup Basketball"),
		Description: strPtr("Casual game at the park"),
		Location: models.Location{
			Name:      "Central Park Basketball Court",
			Address:   strPtr("123 Main St"),
			Latitude:  floatPtr(40.7829),
			Longitude: floatPtr(-73.9654),
		},
		StartTime:       startTime,
		DurationMinutes: 90,
		MaxParticipants: 10,
		Pricing: models.Pricing{
			Type:        models.PricingTypeFree,
			AmountCents: 0,
			Currency:    "USD",
		},
		SkillLevel: &skillLevel,
	}

	// Create game
	game, err := client.CreateGame(req)
	AssertNoError(t, err)
	defer CleanupGame(ctx, game.ID)

	// Assertions
	if game.Category != models.GameCategoryBasketball {
		t.Errorf("expected category %s, got %s", models.GameCategoryBasketball, game.Category)
	}
	if game.Owner.ID != authResp.User.ID {
		t.Errorf("expected owner %s, got %s", authResp.User.ID, game.Owner.ID)
	}
	if game.Status != models.GameStatusOpen {
		t.Errorf("expected status %s, got %s", models.GameStatusOpen, game.Status)
	}
	if game.MaxParticipants != 10 {
		t.Errorf("expected max participants 10, got %d", game.MaxParticipants)
	}
	if game.DurationMinutes != 90 {
		t.Errorf("expected duration 90, got %d", game.DurationMinutes)
	}
	if *game.Title != "Pickup Basketball" {
		t.Errorf("expected title 'Pickup Basketball', got %s", *game.Title)
	}
}

func TestCreateGame_Unauthorized(t *testing.T) {
	client := NewTestClient()

	startTime := time.Now().Add(24 * time.Hour)
	skillLevel := models.SkillLevelAll

	req := models.CreateGameRequest{
		Category:        models.GameCategoryBasketball,
		StartTime:       startTime,
		DurationMinutes: 90,
		MaxParticipants: 10,
		Location: models.Location{
			Name: "Central Park",
		},
		Pricing: models.Pricing{
			Type:        models.PricingTypeFree,
			AmountCents: 0,
			Currency:    "USD",
		},
		SkillLevel: &skillLevel,
	}

	// Try to create game without token
	var game models.Game
	httpResp, _ := client.POST("/v1/games", req, &game)

	AssertStatusCode(t, http.StatusUnauthorized, httpResp.StatusCode)
}

func TestCreateGame_InvalidDuration(t *testing.T) {
	client := NewTestClient()
	ctx := context.Background()

	// Register user
	email := TestEmail(t)
	authResp, err := client.RegisterUser(email, "password123@", "John", "Doe")
	AssertNoError(t, err)
	defer CleanupUser(ctx, authResp.User.ID)

	startTime := time.Now().Add(24 * time.Hour)
	skillLevel := models.SkillLevelAll

	req := models.CreateGameRequest{
		Category:        models.GameCategoryBasketball,
		StartTime:       startTime,
		DurationMinutes: 5, // Too short, min is 15
		MaxParticipants: 10,
		Location: models.Location{
			Name: "Central Park",
		},
		Pricing: models.Pricing{
			Type:        models.PricingTypeFree,
			AmountCents: 0,
			Currency:    "USD",
		},
		SkillLevel: &skillLevel,
	}

	var game models.Game
	httpResp, _ := client.POST("/v1/games", req, &game)

	AssertStatusCode(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestCreateGame_InvalidMaxParticipants(t *testing.T) {
	client := NewTestClient()
	ctx := context.Background()

	// Register user
	email := TestEmail(t)
	authResp, err := client.RegisterUser(email, "password123@", "John", "Doe")
	AssertNoError(t, err)
	defer CleanupUser(ctx, authResp.User.ID)

	startTime := time.Now().Add(24 * time.Hour)
	skillLevel := models.SkillLevelAll

	req := models.CreateGameRequest{
		Category:        models.GameCategoryBasketball,
		StartTime:       startTime,
		DurationMinutes: 90,
		MaxParticipants: 1, // Too few, min is 2
		Location: models.Location{
			Name: "Central Park",
		},
		Pricing: models.Pricing{
			Type:        models.PricingTypeFree,
			AmountCents: 0,
			Currency:    "USD",
		},
		SkillLevel: &skillLevel,
	}

	var game models.Game
	httpResp, _ := client.POST("/v1/games", req, &game)

	AssertStatusCode(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestListGames_Success(t *testing.T) {
	client := NewTestClient()
	ctx := context.Background()

	// Register user and create game
	email := TestEmail(t)
	authResp, err := client.RegisterUser(email, "password123@", "John", "Doe")
	AssertNoError(t, err)
	defer CleanupUser(ctx, authResp.User.ID)

	startTime := time.Now().Add(24 * time.Hour)
	skillLevel := models.SkillLevelAll

	gameReq := models.CreateGameRequest{
		Category:        models.GameCategoryBasketball,
		StartTime:       startTime,
		DurationMinutes: 90,
		MaxParticipants: 10,
		Location: models.Location{
			Name:      "Central Park",
			Latitude:  floatPtr(40.7829),
			Longitude: floatPtr(-73.9654),
		},
		Pricing: models.Pricing{
			Type:        models.PricingTypeFree,
			AmountCents: 0,
			Currency:    "USD",
		},
		SkillLevel: &skillLevel,
	}

	game, err := client.CreateGame(gameReq)
	AssertNoError(t, err)
	defer CleanupGame(ctx, game.ID)

	// List games near the location
	var listResp models.ListGamesResponse
	httpResp, err := client.GET("/v1/games?categories=basketball&latitude=40.7829&longitude=-73.9654&radius=10000", &listResp)
	AssertNoError(t, err)
	AssertStatusCode(t, http.StatusOK, httpResp.StatusCode)

	// Should find at least the game we created
	if len(listResp.Games) == 0 {
		t.Error("expected at least one game in list")
	}

	found := false
	for _, g := range listResp.Games {
		if g.ID == game.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("created game not found in list")
	}
}

func TestListGames_MultipleCategories(t *testing.T) {
	client := NewTestClient()
	ctx := context.Background()

	// Register user
	email := TestEmail(t)
	authResp, err := client.RegisterUser(email, "password123@", "John", "Doe")
	AssertNoError(t, err)
	defer CleanupUser(ctx, authResp.User.ID)

	startTime := time.Now().Add(24 * time.Hour)
	skillLevel := models.SkillLevelAll

	// Create basketball game
	basketballGame, err := client.CreateGame(models.CreateGameRequest{
		Category:        models.GameCategoryBasketball,
		StartTime:       startTime,
		DurationMinutes: 90,
		MaxParticipants: 10,
		Location: models.Location{
			Name:      "Central Park",
			Latitude:  floatPtr(40.7829),
			Longitude: floatPtr(-73.9654),
		},
		Pricing: models.Pricing{
			Type:        models.PricingTypeFree,
			AmountCents: 0,
			Currency:    "USD",
		},
		SkillLevel: &skillLevel,
	})
	AssertNoError(t, err)
	defer CleanupGame(ctx, basketballGame.ID)

	// Create soccer game
	soccerGame, err := client.CreateGame(models.CreateGameRequest{
		Category:        models.GameCategorySoccer,
		StartTime:       startTime,
		DurationMinutes: 90,
		MaxParticipants: 22,
		Location: models.Location{
			Name:      "Soccer Field",
			Latitude:  floatPtr(40.7830),
			Longitude: floatPtr(-73.9655),
		},
		Pricing: models.Pricing{
			Type:        models.PricingTypeFree,
			AmountCents: 0,
			Currency:    "USD",
		},
		SkillLevel: &skillLevel,
	})
	AssertNoError(t, err)
	defer CleanupGame(ctx, soccerGame.ID)

	// List games for both categories
	var listResp models.ListGamesResponse
	httpResp, err := client.GET("/v1/games?categories=basketball&categories=soccer&latitude=40.7829&longitude=-73.9654&radius=10000", &listResp)
	AssertNoError(t, err)
	AssertStatusCode(t, http.StatusOK, httpResp.StatusCode)

	// Should find both games
	foundBasketball := false
	foundSoccer := false
	for _, g := range listResp.Games {
		if g.ID == basketballGame.ID {
			foundBasketball = true
		}
		if g.ID == soccerGame.ID {
			foundSoccer = true
		}
	}

	if !foundBasketball {
		t.Error("basketball game not found in list")
	}
	if !foundSoccer {
		t.Error("soccer game not found in list")
	}
}

func TestListGames_MissingRequiredParams(t *testing.T) {
	client := NewTestClient()

	tests := []struct {
		name string
		path string
	}{
		{
			name: "missing categories",
			path: "/v1/games?latitude=40.7829&longitude=-73.9654",
		},
		{
			name: "missing latitude",
			path: "/v1/games?categories=basketball&longitude=-73.9654",
		},
		{
			name: "missing longitude",
			path: "/v1/games?categories=basketball&latitude=40.7829",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var listResp models.ListGamesResponse
			httpResp, _ := client.GET(tt.path, &listResp)

			AssertStatusCode(t, http.StatusBadRequest, httpResp.StatusCode)
		})
	}
}

func TestListGames_EmptyResult(t *testing.T) {
	client := NewTestClient()

	// List games in the middle of the ocean (no games there)
	var listResp models.ListGamesResponse
	httpResp, err := client.GET("/v1/games?categories=basketball&latitude=0.0&longitude=0.0&radius=1000", &listResp)
	AssertNoError(t, err)
	AssertStatusCode(t, http.StatusOK, httpResp.StatusCode)

	// Should return empty array, not null
	if listResp.Games == nil {
		t.Error("expected empty array, got nil")
	}
	if len(listResp.Games) != 0 {
		t.Errorf("expected 0 games, got %d", len(listResp.Games))
	}
}
