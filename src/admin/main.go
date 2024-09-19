package main

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"

	"blacksmithlabs.dev/k8s-webauthn/admin/config"
)

var (
	sessionTimeout = config.GetSessionTimeout()
)

func main() {
	// Initialize Gin
	engine := gin.Default()

	// Enable CORS
	// if origins := config.GetRPOrigins(); len(origins) > 0 {
	// 	corsConfig := cors.DefaultConfig()
	// 	corsConfig.AllowOrigins = origins
	// 	corsConfig.AllowCredentials = true
	// 	engine.Use(cors.New(corsConfig))
	// }

	// Initialize Redis session store
	store, err := redis.NewStore(
		config.GetRedisPoolSize(),
		"tcp",
		config.GetRedisHost(),
		config.GetRedisPassword(),
		config.GetSessionSecret(),
	)
	if err != nil {
		panic(fmt.Errorf("failed to initialize redis session: %w", err))
	}
	engine.Use(sessions.Sessions("auth-it-admin", store))

	// Set up routes
	engine.GET("/_health", (func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	}))

	// Run Gin
	engine.Run(":8080")
}
