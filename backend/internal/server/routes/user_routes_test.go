package routes

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUserRoutesRegisterAvailableModelsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")

	RegisterUserRoutes(
		v1,
		&handler.Handlers{AvailableChannel: &handler.AvailableChannelHandler{}},
		middleware.JWTAuthMiddleware(func(c *gin.Context) { c.Next() }),
		nil,
		nil,
		nil,
	)

	registered := false
	for _, route := range router.Routes() {
		if route.Method == http.MethodGet && route.Path == "/api/v1/models/available" {
			registered = true
			break
		}
	}

	require.True(t, registered, "GET /api/v1/models/available should be registered")
}
