package models

// PlaceSearchRequest represents a request to search for places (Google Places API v1)
type PlaceSearchRequest struct {
	TextQuery string `json:"textQuery" binding:"required,min=3"` // Search query (minimum 3 characters)
}

// PlacePrediction represents a single place prediction
type PlacePrediction struct {
	PlaceID     string                    `json:"place_id"`              // Google Place ID
	Description string                    `json:"description"`           // Full description
	Formatting  PlaceStructuredFormatting `json:"structured_formatting"` // Structured text
}

// PlaceStructuredFormatting represents formatted place text
type PlaceStructuredFormatting struct {
	MainText      string `json:"main_text"`      // Primary text (e.g., "Memorial Park")
	SecondaryText string `json:"secondary_text"` // Secondary text (e.g., "Houston, TX, USA")
}

// PlaceAutocompleteResponse represents the response from search endpoint
type PlaceAutocompleteResponse struct {
	Predictions []PlacePrediction `json:"predictions"` // List of place predictions
}

// PlaceDetailsResponse represents detailed information about a place
type PlaceDetailsResponse struct {
	PlaceID          string  `json:"place_id"`          // Google Place ID
	Name             string  `json:"name"`              // Place name
	FormattedAddress string  `json:"formatted_address"` // Full formatted address
	Latitude         float64 `json:"latitude"`          // Latitude coordinate
	Longitude        float64 `json:"longitude"`         // Longitude coordinate
}

// GooglePlacesSearchResponse represents Google's new Places API v1 searchText response
type GooglePlacesSearchResponse struct {
	Places []struct {
		ID               string `json:"id"`               // Place ID (format: places/{place_id})
		FormattedAddress string `json:"formattedAddress"` // Full address
		DisplayName      struct {
			Text string `json:"text"` // Display name
		} `json:"displayName"`
		Location struct {
			Latitude  float64 `json:"latitude"`  // Latitude
			Longitude float64 `json:"longitude"` // Longitude
		} `json:"location"`
	} `json:"places"`
}

// GooglePlaceDetailsResponse represents Google's new Places API v1 place details response
type GooglePlaceDetailsResponse struct {
	ID               string `json:"id"`               // Place ID (format: places/{place_id})
	FormattedAddress string `json:"formattedAddress"` // Full address
	DisplayName      struct {
		Text string `json:"text"` // Display name
	} `json:"displayName"`
	Location struct {
		Latitude  float64 `json:"latitude"`  // Latitude
		Longitude float64 `json:"longitude"` // Longitude
	} `json:"location"`
}
