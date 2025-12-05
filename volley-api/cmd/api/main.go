package main

import (
	"github.com/gabe-dev-svc/volley/internal/api"
)

func main() {
	api.NewServer().Run("8080")
}
