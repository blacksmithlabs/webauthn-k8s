package controllers

import (
	"net/http"

	credential_service "blacksmithlabs.dev/webauthn-k8s/app/services/credential"
	"blacksmithlabs.dev/webauthn-k8s/shared/dto"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// POST /authentication/ end point to handle getting the params for authenticating a credential
func BeginAuthentication(c *gin.Context) {
	var requestPayload dto.StartAuthenticationRequest
	if err := c.BindJSON(&requestPayload); err != nil {
		logger.Error("Invalid request format", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid request format"})
		return
	}
	if err := requestPayload.Validate(); err != nil {
		logger.Error("Invalid request payload", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid request payload"})
		return
	}

	service, err := credential_service.New(c)
	if err != nil {
		logger.Error("Failed to get credentials service", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Database error"})
		return
	}
	user, err := service.GetUserWithCredentialsByRef(requestPayload.User.UserID)
	if err != nil {
		logger.Error("Failed to get user", "error", err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error(), "message": "User not found"})
		return
	}
	if !user.Credentials.Loaded || len(user.Credentials.Value) == 0 {
		logger.Error("User has no credentials", "user", user)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "No credentials found", "message": "User has no credentials"})
		return
	}

	webAuthn := c.MustGet("webauthn").(*webauthn.WebAuthn)
	options, sessionData, err := webAuthn.BeginLogin(user)
	if err != nil {
		logger.Error("Failed to create authentication options", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err, "message": "Failed to create authentication options"})
		return
	}

	requestId := uuid.New().String()
	session := getSession(c)
	logger.Info("Saving request data to session", "requestId", requestId, "user", requestPayload.User, "session", sessionData)
	session.Set(requestId, gin.H{"userId": user.ID, "session": sessionData})
	if err := session.Save(); err != nil {
		logger.Error("Failed to save session data", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to save session data"})
		return
	}

	c.JSON(http.StatusOK, dto.StartAuthenticationResponse{
		RequestID: requestId,
		Options:   *options,
	})
}

// PUT /authentication/:requestId end point to handle actually authenticating a credential
func FinishAuthentication(c *gin.Context) {
	requestId := c.Param("requestId")
	logger.Info("Finish authentication", "requestId", requestId)

	session := getSession(c)
	sessionPayload := session.Get(requestId)

	if sessionPayload == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Request Not Found", "requestId": requestId})
		return
	}

	userId := sessionPayload.(gin.H)["userId"].(int64)
	sessionData := sessionPayload.(gin.H)["session"].(webauthn.SessionData)

	// Get the user for this credential
	service, err := credential_service.New(c)
	if err != nil {
		logger.Error("Failed to get credentials service", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Database error"})
		return
	}

	user, err := service.GetUserWithCredentialsByID(userId)
	if err != nil {
		logger.Error("Failed to get user", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "User lookup failed"})
		return
	}

	logger.Info("Loaded user for request", "user", user)

	webAuthn := c.MustGet("webauthn").(*webauthn.WebAuthn)
	var requestPayload dto.FinishAuthenticationRequest
	if err := c.BindJSON(&requestPayload); err != nil {
		logger.Error("Invalid request format", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid request format"})
		return
	}

	logger.Info("Preparing to parse assertion", "assertion", requestPayload.Assertion)

	parsedAssertion, err := requestPayload.Assertion.Parse()
	if err != nil {
		logger.Error("Failed to parse assertion", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err, "message": "Failed to parse assertion"})
		return
	}

	credential, err := webAuthn.ValidateLogin(user, sessionData, parsedAssertion)
	if err != nil {
		logger.Error("Failed to validate login", "error", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err, "message": "Failed to validate login"})
		return
	}

	count, err := service.IncrementCredentialUseCounter(credential.ID)
	if err != nil {
		logger.Error("Failed to increment credential use counter", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err, "message": "Failed to increment credential use counter"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully authenticated", "credential": credential, "useCount": count})
}
