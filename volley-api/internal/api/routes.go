package api

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
			auth.POST("/refresh", h.RefreshToken)
		}
		// Games routes
		games := v1.Group("/games")
		{
			games.GET("", OptionalAuthMiddleware(), h.ListGames)
			games.POST("", AuthMiddleware(), h.CreateGame)
			games.GET("/:gameId", AuthMiddleware(), h.GetGame)
			games.PATCH("/:gameId", AuthMiddleware(), h.UpdateGame)
			games.DELETE("/:gameId", AuthMiddleware(), h.DeleteGame)
			games.POST("/:gameId/participation", AuthMiddleware(), h.JoinGame)
			games.DELETE("/:gameId/participation", AuthMiddleware(), h.DropGame)
			games.POST("/:gameId/cancel", AuthMiddleware(), h.CancelGame)
		}

		// Places routes (Google Places API v1 proxy)
		places := v1.Group("/places")
		places.Use(AuthMiddleware())
		{
			places.POST("/search", h.PlacesAutocomplete)
			places.GET("/:placeId", h.PlaceDetails)
		}
	}
}
