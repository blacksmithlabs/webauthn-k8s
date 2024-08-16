package main

import (
	"encoding/gob"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"

	"blacksmithlabs.dev/webauthn-k8s/app/config"
	"blacksmithlabs.dev/webauthn-k8s/app/controllers"
	"blacksmithlabs.dev/webauthn-k8s/shared/dto"
)

var (
	webAuthn       *webauthn.WebAuthn
	err            error
	sessionTimeout time.Duration = config.GetSessionTimeout()
)

func main() {
	// Initialize code depenedencies
	gob.Register(gin.H{})
	gob.Register(dto.RegistrationUserInfo{})
	gob.Register(webauthn.SessionData{})

	// Initialize WebAuthn
	timeoutConfig := webauthn.TimeoutConfig{
		Enforce:    true,
		Timeout:    sessionTimeout,
		TimeoutUVD: sessionTimeout,
	}
	wconfig := &webauthn.Config{
		RPDisplayName: config.GetRPDisplayName(),
		RPID:          config.GetRPID(),
		RPOrigins:     config.GetRPOrigins(),
		Timeouts: webauthn.TimeoutsConfig{
			Login:        timeoutConfig,
			Registration: timeoutConfig,
		},
	}

	if webAuthn, err = webauthn.New(wconfig); err != nil {
		fmt.Println("Failed to create Webauthn handler", err)
		return
	}

	// Initialize Gin
	engine := gin.Default()
	// Bind the WebAuthn instance to the context
	engine.Use(func(ctx *gin.Context) {
		ctx.Set("webauthn", webAuthn)
	})

	// Enable CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173"}
	corsConfig.AllowCredentials = true
	engine.Use(cors.New(corsConfig))

	// Initialize Redis session store
	store, err := redis.NewStore(
		config.GetRedisPoolSize(),
		"tcp",
		config.GetRedisHost(),
		config.GetRedisPassword(),
		config.GetSessionSecret(),
	)
	if err != nil {
		fmt.Println("Failed to initialize redis session", err)
		return
	}
	engine.Use(sessions.Sessions("webauthn", store))

	// Set up routes
	engine.GET("/users/:userId/credentials/", controllers.GetUserCredentials)
	engine.POST("/credentials/", controllers.BeginCreateCredential)
	engine.PUT("/credentials/:requestId", controllers.FinishCreateCredential)
	engine.POST("/authentication/", controllers.BeginAuthentication)
	engine.PUT("/authentication/:requestId", controllers.FinishAuthentication)

	// Run Gin
	engine.Run(":" + config.GetAppPort())
}
