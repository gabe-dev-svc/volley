package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/gabe-dev-svc/volley/internal/errors"
	"github.com/gabe-dev-svc/volley/internal/models"
	"github.com/gabe-dev-svc/volley/internal/service"
	"github.com/gabe-dev-svc/volley/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	gamesService    *service.GamesService
	userService     *service.UserService
	googlePlacesKey string
}

func NewHandler(gamesService *service.GamesService, userService *service.UserService, googlePlacesKey string) *Handler {
	return &Handler{
		gamesService:    gamesService,
		userService:     userService,
		googlePlacesKey: googlePlacesKey,
	}
}

// getUserID extracts and validates the authenticated user ID from the Gin context
func getUserID(c *gin.Context) (string, error) {
	userID := c.GetString("userID")
	if userID == "" {
		return "", fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// ListGames handles GET /games
func (h *Handler) ListGames(c *gin.Context) {
	logger := LoggerFromContext(c)
	ctx := logger.WithContext(c.Request.Context())

	// Parse query parameters
	categories := c.QueryArray("categories")
	if len(categories) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one sport category is required"})
		return
	}

	latitude := c.Query("latitude")
	if latitude == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "latitude is required"})
		return
	}

	longitude := c.Query("longitude")
	if longitude == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "longitude is required"})
		return
	}

	// Parse latitude and longitude
	var lat, lng float64
	if _, err := fmt.Sscanf(latitude, "%f", &lat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude"})
		return
	}
	if _, err := fmt.Sscanf(longitude, "%f", &lng); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude"})
		return
	}

	// Parse optional parameters
	var radius float64 = 16093.4 // Default 10 miles in meters
	if radiusStr := c.Query("radius"); radiusStr != "" {
		if _, err := fmt.Sscanf(radiusStr, "%f", &radius); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid radius"})
			return
		}
	}

	timeFilter := service.TimeFilterUpcoming // Default
	if timeFilterStr := c.Query("timeFilter"); timeFilterStr != "" {
		timeFilter = service.TimeFilter(timeFilterStr)
		if timeFilter != service.TimeFilterUpcoming &&
			timeFilter != service.TimeFilterPast &&
			timeFilter != service.TimeFilterAll {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timeFilter (must be: upcoming, past, or all)"})
			return
		}
	}

	var status *string
	if statusStr := c.Query("status"); statusStr != "" {
		status = &statusStr
	}

	var limit int = 20 // Default
	if limitStr := c.Query("limit"); limitStr != "" {
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
	}

	var offset int = 0 // Default
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
			return
		}
	}

	// Extract user ID from auth context (if authenticated)
	var userID *string
	if uid, exists := c.Get("userID"); exists {
		if uidStr, ok := uid.(string); ok && uidStr != "" {
			userID = &uidStr
		}
	}

	// Call service
	games, err := h.gamesService.ListGames(ctx, service.ListGamesFilters{
		Categories: categories,
		Latitude:   lat,
		Longitude:  lng,
		Radius:     radius,
		TimeFilter: timeFilter,
		Status:     status,
		Limit:      limit,
		Offset:     offset,
	}, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list games")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Ensure we return an empty array instead of null
	if games == nil {
		games = []models.GameSummary{}
	}

	c.JSON(http.StatusOK, models.ListGamesResponse{Games: games})
}

// CreateGame handles POST /games
func (h *Handler) CreateGame(c *gin.Context) {
	logger := log.With().
		Caller().
		Str("handler", "CreateGame").
		Str("path", c.Request.URL.Path).
		Logger()
	ctx := logger.WithContext(c.Request.Context())

	// Extract userID from auth middleware context
	userIDStr, err := getUserID(c)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to extract user ID from context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req models.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	logger.Info().
		Str("userID", userIDStr).
		Interface("request", req).
		Msg("Creating game")

	game, err := h.gamesService.CreateGame(ctx, userIDStr, req)
	if err != nil {
		logger.Error().Err(err).Str("userID", userIDStr).Msg("Failed to create game")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game"})
		return
	}

	logger.Info().
		Str("userID", userIDStr).
		Str("gameID", game.ID).
		Msg("Game created successfully")

	c.JSON(http.StatusCreated, game)
}

// GetGame handles GET /games/:gameId
func (h *Handler) GetGame(c *gin.Context) {
	logger := LoggerFromContext(c)
	ctx := logger.WithContext(c.Request.Context())

	gameID := c.Param("gameId")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	logger = logger.With().Str("gameId", gameID).Logger()
	ctx = logger.WithContext(ctx)

	game, err := h.gamesService.GetGame(ctx, gameID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			logger.Warn().Err(err).Msg("Game not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to get game")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve game"})
		return
	}

	c.JSON(http.StatusOK, game)
}

// UpdateGame handles PATCH /games/:gameId
func (h *Handler) UpdateGame(c *gin.Context) {
	// TODO: Implement
}

// DeleteGame handles DELETE /games/:gameId
func (h *Handler) DeleteGame(c *gin.Context) {
	// TODO: Implement
}

// JoinGame handles POST /games/:gameId/participation
func (h *Handler) JoinGame(c *gin.Context) {
	logger := LoggerFromContext(c)
	ctx := logger.WithContext(c.Request.Context())

	userID, err := getUserID(c)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to extract user ID from context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	gameID := c.Param("gameId")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	logger = logger.With().Str("userId", userID).Str("gameId", gameID).Logger()
	ctx = logger.WithContext(ctx)

	participant, err := h.gamesService.JoinGame(ctx, gameID, userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to join game")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info().Str("gameID", gameID).Msg("User joined game")
	c.JSON(http.StatusOK, participant)
}

// DropGame handles POST /games/:gameId/drop
func (h *Handler) DropGame(c *gin.Context) {
	logger := LoggerFromContext(c)
	ctx := logger.WithContext(c.Request.Context())

	userID, err := getUserID(c)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to extract user ID from context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	gameID := c.Param("gameId")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	logger = logger.With().Str("userId", userID).Str("gameId", gameID).Logger()
	ctx = logger.WithContext(ctx)

	result, err := h.gamesService.DropParticipantFromGame(ctx, gameID, userID)
	if err != nil {
		// Handle specific error types
		if errors.Is(err, service.ErrTooLate) {
			logger.Warn().Err(err).Msg("Drop deadline has passed")
			c.JSON(http.StatusForbidden, gin.H{"error": "Drop deadline has passed"})
			return
		}
		if errors.Is(err, service.ErrGameFinished) {
			logger.Warn().Err(err).Msg("Game has already finished")
			c.JSON(http.StatusForbidden, gin.H{"error": "Game has already finished"})
			return
		}
		if errors.Is(err, service.ErrNotParticipant) {
			logger.Warn().Err(err).Msg("User is not a participant")
			c.JSON(http.StatusBadRequest, gin.H{"error": "You are not a participant of this game"})
			return
		}

		logger.Error().Err(err).Msg("Failed to drop from game")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to drop from game"})
		return
	}

	logger.Info().Msg("User dropped from game successfully")

	// TODO: Send push notification to promoted user if applicable
	if result.PromotedUser != nil {
		logger.Info().
			Str("promotedUserId", result.PromotedUser.ID).
			Str("promotedUserEmail", result.PromotedUser.Email).
			Msg("TODO: Send push notification to promoted user")
		// Future: h.notificationService.SendWaitlistPromotion(ctx, result.PromotedUser, gameID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully dropped from game"})
}

// CancelGame handles POST /games/:gameId/cancel
func (h *Handler) CancelGame(c *gin.Context) {
	logger := LoggerFromContext(c)
	ctx := logger.WithContext(c.Request.Context())

	userID, err := getUserID(c)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to extract user ID from context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	gameID := c.Param("gameId")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game ID is required"})
		return
	}

	logger = logger.With().Str("userId", userID).Str("gameId", gameID).Logger()
	ctx = logger.WithContext(ctx)

	result, err := h.gamesService.CancelGame(ctx, gameID, userID)
	if err != nil {
		// Handle specific error types
		if errors.Is(err, service.ErrNotOwner) {
			logger.Warn().Err(err).Msg("User is not the game owner")
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the game owner can cancel the game"})
			return
		}
		if errors.Is(err, service.ErrGameAlreadyStarted) {
			logger.Warn().Err(err).Msg("Cannot cancel a game that has already started")
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot cancel a game that has already started"})
			return
		}
		if errors.Is(err, service.ErrGameFinished) {
			logger.Warn().Err(err).Msg("Cannot cancel a game that has already finished")
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot cancel a game that has already finished"})
			return
		}
		if strings.Contains(err.Error(), "game not found") {
			logger.Warn().Err(err).Msg("Game not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}

		logger.Error().Err(err).Msg("Failed to cancel game")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel game"})
		return
	}

	logger.Info().Int("participantCount", len(result.ParticipantsToNotify)).Msg("Game cancelled successfully")

	// TODO: Send push notifications to all participants
	if len(result.ParticipantsToNotify) > 0 {
		logger.Info().
			Int("participantCount", len(result.ParticipantsToNotify)).
			Msg("TODO: Send push notifications to all participants about game cancellation")
		// Future: h.notificationService.SendGameCancellation(ctx, result.ParticipantsToNotify, gameID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game cancelled successfully"})
}

// Register handles POST /auth/register
func (h *Handler) Register(c *gin.Context) {

	// get logger from gin context and register it with background context
	logger := LoggerFromContext(c)
	ctx := logger.WithContext(c.Request.Context())

	var req models.RegisterRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user
	user, err := h.userService.CreateUser(ctx, service.CreateUserRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrAlreadyExists) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		logger.Error().Err(err).Msg("CreateUser failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := util.GenerateToken(user.ID, req.Email, req.FirstName, req.LastName, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Prepare response
	resp := models.AuthResponse{
		User: *user,
	}

	// Detect if request is from mobile or web
	if isMobileClient(c) {
		// For mobile clients: include token in JSON response
		resp.Token = &token
	} else {
		// For web clients: set token in HTTP-only secure cookie
		setAuthCookie(c, token)
	}

	c.JSON(http.StatusCreated, resp)
}

// isMobileClient determines if the request is from a mobile app
func isMobileClient(c *gin.Context) bool {
	// Check for custom header (recommended)
	if c.GetHeader("X-Client-Type") == "mobile" {
		return true
	}

	// Fallback: check User-Agent for mobile app identifiers
	userAgent := c.GetHeader("User-Agent")
	return strings.Contains(userAgent, "VolleyMobile") ||
		strings.Contains(userAgent, "Flutter") ||
		strings.Contains(userAgent, "Dart")
}

// setAuthCookie sets an HTTP-only secure cookie with the JWT token
func setAuthCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"auth_token",         // name
		token,                // value
		60*60*24*7,           // maxAge in seconds (7 days)
		"/",                  // path
		"",                   // domain (empty means current domain)
		c.Request.TLS != nil, // secure (true if HTTPS)
		true,                 // httpOnly
	)
}

func (h *Handler) Login(c *gin.Context) {
	logger := LoggerFromContext(c)
	ctx := logger.WithContext(c.Request.Context())

	var req models.LoginRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate user
	user, err := h.userService.Login(ctx, req.Email, req.Password)
	if err != nil {
		logger.Warn().Err(err).Str("email", req.Email).Msg("Login failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := util.GenerateToken(user.ID, user.Email, user.FirstName, user.LastName, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Prepare response
	resp := models.AuthResponse{
		User: *user,
	}

	// Detect if request is from mobile or web
	if isMobileClient(c) {
		// For mobile clients: include token in JSON response
		resp.Token = &token
	} else {
		// For web clients: set token in HTTP-only secure cookie
		setAuthCookie(c, token)
	}

	logger.Info().Str("email", user.Email).Msg("User logged in successfully")
	c.JSON(http.StatusOK, resp)
}

// PlacesAutocomplete handles POST /places/search (Google Places API v1)
func (h *Handler) PlacesAutocomplete(c *gin.Context) {
	logger := LoggerFromContext(c)

	// Bind and validate request body
	var req models.PlaceSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if API key is configured
	if h.googlePlacesKey == "" {
		logger.Error().Msg("Google Places API key not configured")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Location service not configured"})
		return
	}

	// Build Google Places API v1 searchText request
	apiURL := "https://places.googleapis.com/v1/places:searchText"

	requestBody, err := json.Marshal(map[string]any{
		"textQuery":      req.TextQuery,
		"maxResultCount": 5,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal request body")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", apiURL, strings.NewReader(string(requestBody)))
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create HTTP request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Goog-Api-Key", h.googlePlacesKey)
	httpReq.Header.Set("X-Goog-FieldMask", "places.id,places.displayName,places.formattedAddress,places.location")

	// Make request to Google Places API
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to call Google Places API")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch location suggestions"})
		return
	}
	defer resp.Body.Close()

	// Parse Google's response
	var googleResp models.GooglePlacesSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		logger.Error().Err(err).Msg("Failed to parse Google Places API response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse location data"})
		return
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		logger.Error().
			Int("httpStatus", resp.StatusCode).
			Msg("Google Places API returned error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Location service error"})
		return
	}

	// Convert to our response format
	predictions := make([]models.PlacePrediction, len(googleResp.Places))
	for i, place := range googleResp.Places {
		addressParts := strings.Split(place.FormattedAddress, ",")
		mainText := place.DisplayName.Text
		if mainText == "" && len(addressParts) > 0 {
			mainText = strings.TrimSpace(addressParts[0])
		}

		secondaryText := ""
		if len(addressParts) > 1 {
			secondaryText = strings.TrimSpace(strings.Join(addressParts[1:], ","))
		}

		predictions[i] = models.PlacePrediction{
			PlaceID:     place.ID,
			Description: place.FormattedAddress,
			Formatting: models.PlaceStructuredFormatting{
				MainText:      mainText,
				SecondaryText: secondaryText,
			},
		}
	}

	// Return response
	c.JSON(http.StatusOK, models.PlaceAutocompleteResponse{
		Predictions: predictions,
	})
}

// PlaceDetails handles GET /places/{placeId} (Google Places API v1)
func (h *Handler) PlaceDetails(c *gin.Context) {
	logger := LoggerFromContext(c)

	// Get place ID parameter (format: places/ChIJ...)
	placeID := c.Param("placeId")
	if placeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "placeId parameter is required"})
		return
	}

	// Check if API key is configured
	if h.googlePlacesKey == "" {
		logger.Error().Msg("Google Places API key not configured")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Location service not configured"})
		return
	}

	// Build Google Places API v1 request
	apiURL := fmt.Sprintf("https://places.googleapis.com/v1/places/%s", placeID)

	// Create HTTP request
	httpReq, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create HTTP request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	httpReq.Header.Set("X-Goog-Api-Key", h.googlePlacesKey)
	httpReq.Header.Set("X-Goog-FieldMask", "displayName,formattedAddress,location")

	// Make request to Google Places API
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to call Google Places API")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch place details"})
		return
	}
	defer resp.Body.Close()

	// Parse Google's response
	var googleResp models.GooglePlaceDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		logger.Error().Err(err).Int("statusCode", resp.StatusCode).Msg("Failed to parse Google Places API response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse location data"})
		return
	}

	// Check HTTP status
	if resp.StatusCode == http.StatusNotFound {
		logger.Warn().Str("placeId", placeID).Msg("Place not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Place not found"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error().
			Int("httpStatus", resp.StatusCode).
			Str("placeId", placeID).
			Msg("Google Places API returned error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Location service error"})
		return
	}

	// Build response
	response := models.PlaceDetailsResponse{
		PlaceID:          googleResp.ID,
		Name:             googleResp.DisplayName.Text,
		FormattedAddress: googleResp.FormattedAddress,
		Latitude:         googleResp.Location.Latitude,
		Longitude:        googleResp.Location.Longitude,
	}

	c.JSON(http.StatusOK, response)
}
