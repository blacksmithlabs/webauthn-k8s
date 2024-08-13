package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	"blacksmithlabs.dev/webauthn-k8s/shared/dto"
)

var (
	RedisHost     = os.Getenv("REDIS_HOST")
	RedisPassword = os.Getenv("REDIS_PASSWORD")
)

var (
	webAuthn *webauthn.WebAuthn
	err      error
	logger   *slog.Logger
)

func main() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Initialize WebAuthn
	wconfig := &webauthn.Config{
		RPDisplayName: "Blacksmith Labs",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	}

	if webAuthn, err = webauthn.New(wconfig); err != nil {
		fmt.Println("Failed to create Webauthn handler", err)
		return
	}

	// Initialize Gin
	engine := gin.Default()

	/* @@ */
	logger.Info("Redis Info", "host", RedisHost, "password", RedisPassword)
	/* ## */

	// Initialize Redis session store
	store, err := redis.NewStore(10, "tcp", RedisHost, RedisPassword, []byte("secret"))
	if err != nil {
		fmt.Println("Failed to initialize redis session", err)
		return
	}
	engine.Use(sessions.Sessions("webauthn", store))

	engine.POST("/credentials/", beginCreateCredential)
	engine.PUT("/credentials/:requestId", finishCreateCredential)
	engine.POST("/authentication/", beginAuthentication)
	engine.PUT("/authentication/:requestId", finishAuthentication)

	// Run Gin
	engine.Run(":8080")
}

func getSession(c *gin.Context) sessions.Session {
	session := sessions.Default(c)
	session.Options(sessions.Options{
		Path:   "/",
		MaxAge: 180,
	})
	return session
}

func beginCreateCredential(c *gin.Context) {
	var requestPayload dto.CreateRegistrationRequest
	if err := c.BindJSON(&requestPayload); err != nil {
		logger.Error("Invalid request format", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid request format"})
		return
	}
	if err := requestPayload.Validate(); err != nil {
		logger.Error("Invalid request payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid request payload"})
		return
	}

	logger.Info("Creating credential for user", "userId", requestPayload.User.UserId)

	options, sessionData, err := webAuthn.BeginRegistration(requestPayload.User)

	if err != nil {
		logger.Error("Failed to create registration options", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to create registration options"})
		return
	}

	requestId := uuid.New().String()
	if sessionDataJson, err := json.Marshal(sessionData); err != nil {
		logger.Error("Failed to save internal session data", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to save internal session data"})
		return
	} else {
		logger.Info("Saving internal session data", "requestId", requestId, "sessionData", string(sessionDataJson))

		session := getSession(c)
		session.Set("requestId", requestId)
		session.Set(requestId, string(sessionDataJson))
		session.Save()
	}

	c.JSON(http.StatusOK, dto.CreateRegistrationResponse{
		RequestId: requestId,
		Options:   *options,
	})
}

func finishCreateCredential(c *gin.Context) {
	// requestId := c.Param("requestId")
	// c.JSON(http.StatusOK, gin.H{"requestId": requestId})
}

func beginAuthentication(c *gin.Context) {
	if len(c.Request.Cookies()) == 0 {
		logger.Info("No cookies found")
	} else {
		for _, cookie := range c.Request.Cookies() {
			logger.Info("Cookie", "name", cookie.Name, "value", cookie.Value)
		}
	}

	session := getSession(c)
	requestId := session.Get("requestId")
	if requestId == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No request ID found in session"})
		return
	}

	sessionDataJson := session.Get(requestId).(string)

	c.JSON(http.StatusOK, gin.H{"requestId": requestId, "sessionData": sessionDataJson})
	// userId := c.Param("userId")
	// c.JSON(http.StatusOK, gin.H{"userId": userId})
}

func finishAuthentication(c *gin.Context) {
	// requestId := c.Param("requestId")
	// c.JSON(http.StatusOK, gin.H{"requestId": requestId})
}
